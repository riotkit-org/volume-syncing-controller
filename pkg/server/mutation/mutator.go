package mutation

import (
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/cache"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/context"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

type PodMutator struct {
	cache *cache.Cache
}

// ProcessAdmissionRequest is retrieving all the required data, calling to resolve, then calling a mutation action
func (m *PodMutator) ProcessAdmissionRequest(review *admissionv1.AdmissionReview, image string) (corev1.Pod, corev1.Pod, error) {
	// retrieve `kind: Pod`
	pod, podResolveErr := ResolvePod(review.Request)
	if podResolveErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(podResolveErr, "Cannot process AdmissionRequest, cannot resolve Pod")
	}

	originalPod := pod.DeepCopy()

	// then retrieve a matching `kind: PodFilesystemSync` object to know how to set up synchronization for the `kind: Pod`
	matchingPodFilesystemSync, matchingErr, matchedAny := m.cache.FindMatchingForPod(pod)
	if matchingErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(matchingErr, "Cannot match any `kind: PodFilesystemSync` for selected `kind: Pod`")
	}
	if !matchedAny {
		return corev1.Pod{}, corev1.Pod{}, errors.New("No matching `kind: PodFilesystemSync` found for Pod")
	}

	// verify secrets
	secretsVerificationErr := VerifySecrets(matchingPodFilesystemSync, pod.Namespace)
	if secretsVerificationErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(secretsVerificationErr, "The secrets are invalid")
	}

	// prepare environment variables
	// DO NOT CONFUSE WITH SECRETS - those are mounted to not leak sensitive information in `kind: Pod` definition
	env, envResolveErr := ResolveTemplatedEnvironment(pod, matchingPodFilesystemSync)
	if envResolveErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(envResolveErr, "Cannot resolve environment variables")
	}

	// finally try to patch the `kind: Pod` using definition from `kind: PodFilesystemSync`
	if err := m.applyPatchToPod(pod, image, matchingPodFilesystemSync, env); err != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(err, "Cannot mutate `kind: Pod`")
	}
	return *pod, *originalPod, nil
}

// applyPatchToPod is applying a patch to `kind: Pod` before it gets scheduled
func (m *PodMutator) applyPatchToPod(pod *corev1.Pod, image string, syncDefinition *v1alpha1.PodFilesystemSync, env map[string]string) error {
	mutationErr := MutatePodByInjectingInitContainer(pod, image, context.NewSynchronizationParameters(pod, syncDefinition, env))
	if mutationErr != nil {
		return errors.Wrap(mutationErr, "Cannot patch `kind: Pod`")
	}
	return nil
}

func NewPodMutator(cache *cache.Cache) PodMutator {
	return PodMutator{
		cache: cache,
	}
}
