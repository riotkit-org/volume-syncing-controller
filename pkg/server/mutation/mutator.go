package mutation

import (
	goCtx "context"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-operator/pkg/client/clientset/versioned"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/cache"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/context"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodMutationController struct {
	cache         *cache.Cache
	riotkitClient *versioned.Clientset
}

func (m *PodMutationController) parseRequestIntoModels(request *admissionv1.AdmissionRequest) (*corev1.Pod, *corev1.Pod, *v1alpha1.PodFilesystemSync, error) {
	pod, podResolveErr := ResolvePod(request)
	if podResolveErr != nil {
		return &corev1.Pod{}, &corev1.Pod{}, &v1alpha1.PodFilesystemSync{}, errors.Wrap(podResolveErr, "Cannot process AdmissionRequest, cannot resolve Pod")
	}

	originalPod := pod.DeepCopy()

	// then retrieve a matching `kind: PodFilesystemSync` object to know how to set up synchronization for the `kind: Pod`
	matchingPodFilesystemSync, matchedAny, matchingErr := m.cache.FindMatchingForPod(pod)
	if matchingErr != nil {
		return &corev1.Pod{}, &corev1.Pod{}, &v1alpha1.PodFilesystemSync{}, errors.Wrap(matchingErr, "Cannot match any `kind: PodFilesystemSync` for selected `kind: Pod`")
	}
	if !matchedAny {
		return &corev1.Pod{}, &corev1.Pod{}, &v1alpha1.PodFilesystemSync{}, errors.New("No matching `kind: PodFilesystemSync` found for Pod")
	}

	return pod, originalPod, matchingPodFilesystemSync, nil
}

// ProcessAdmissionRequest is retrieving all the required data, calling to resolve, then calling a mutation action
func (m *PodMutationController) ProcessAdmissionRequest(review *admissionv1.AdmissionReview, image string) (corev1.Pod, corev1.Pod, error) {
	pod, originalPod, matchingPodFilesystemSync, resolveErr := m.parseRequestIntoModels(review.Request)
	if resolveErr != nil {
		return corev1.Pod{}, corev1.Pod{}, resolveErr
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

	params, paramsErr := context.NewSynchronizationParameters(pod, matchingPodFilesystemSync, env)
	if paramsErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(paramsErr, "Cannot create patch for `kind: Pod`")
	}

	// finally try to patch the `kind: Pod` using definition from `kind: PodFilesystemSync`
	if err := m.applyPatchToPod(pod, image, matchingPodFilesystemSync, params); err != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(err, "Cannot mutate `kind: Pod`")
	}

	// change status
	if claimErr := matchingPodFilesystemSync.ClaimDirectoryByPod(pod); claimErr != nil {
		return corev1.Pod{}, corev1.Pod{}, errors.Wrap(claimErr, "Cannot claim directory for `kind: Pod`")
	}
	if err := m.updateStatus(matchingPodFilesystemSync); err != nil {
		return corev1.Pod{}, corev1.Pod{}, err
	}

	return *pod, *originalPod, nil
}

// updateStatus is updating status field of a PodFilesystemSync object
func (m *PodMutationController) updateStatus(syncDefinition *v1alpha1.PodFilesystemSync) error {
	logrus.Debug("Updating status")

	client := m.riotkitClient.RiotkitV1alpha1().PodFilesystemSyncs(syncDefinition.Namespace)
	clusterDefinition, getErr := client.Get(goCtx.TODO(), syncDefinition.Name, v1.GetOptions{})
	if getErr != nil {
		return errors.Wrap(getErr, "Cannot update status field - error retrieving object")
	}

	syncDefinition.SetResourceVersion(clusterDefinition.GetResourceVersion())
	_, statusUpdateErr := client.UpdateStatus(
		goCtx.TODO(), syncDefinition, v1.UpdateOptions{})
	if statusUpdateErr != nil {
		return errors.Wrap(statusUpdateErr, "Cannot update status field")
	}
	return nil
}

// applyPatchToPod is applying a patch to `kind: Pod` before it gets scheduled
func (m *PodMutationController) applyPatchToPod(pod *corev1.Pod, image string, syncDefinition *v1alpha1.PodFilesystemSync, params context.SynchronizationParameters) error {
	// decide if we should start the init container with remote-to-local-sync
	shouldRestoreFromRemoteOnInit, configErr := syncDefinition.ShouldRestoreFilesFromRemote(pod)
	if configErr != nil {
		return errors.Wrap(configErr, "Error creating patch for `kind: Pod` - cannot decide if the `kind: Pod` should restore files from remote on init")
	}

	mutationErr := MutatePodByInjectingContainers(pod, image, shouldRestoreFromRemoteOnInit, syncDefinition.ShouldSynchronizeToRemote(), params)
	if mutationErr != nil {
		return errors.Wrap(mutationErr, "Cannot patch `kind: Pod`")
	}
	return nil
}

func NewPodMutator(cache *cache.Cache, riotkitClient *versioned.Clientset) PodMutationController {
	return PodMutationController{
		cache:         cache,
		riotkitClient: riotkitClient,
	}
}
