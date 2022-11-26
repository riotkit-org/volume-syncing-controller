package mutation

import (
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-controller/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-controller/pkg/server/context"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

// MutatePodByInjectingContainers returns a new mutated pod according to set env rules
func MutatePodByInjectingContainers(pod *corev1.Pod, image string, preSynchronizeFromRemoteOnStart bool, canSynchronizeToRemote bool, params context.SynchronizationParameters) error {
	nLogger := logrus.WithField("mutation", "Mutating pod")

	if preSynchronizeFromRemoteOnStart && !hasInitContainer(pod) {
		nLogger.Infof("`kind: Pod` '%s' has no initContainer present", pod.ObjectMeta.Name)
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, createContainer(true, context.InitContainerName, pod, image, params, params.CreateCommandlineArgumentsForInitContainer()))

		// make sure the initContainer is placed in a proper order
		var err error
		pod.Spec.InitContainers, err = reorderInitContainer(pod.Spec.InitContainers, context.InitContainerName, params.InitContainerPlacement)
		if err != nil {
			return errors.Wrapf(err, "Cannot reorder initContainers to place '%s' in proper place", context.InitContainerName)
		}
	}

	if canSynchronizeToRemote && !hasSideCar(pod) {
		nLogger.Infof("`kind: Pod` '%s' has no side car present", pod.ObjectMeta.Name)
		pod.Spec.Containers = append(pod.Spec.Containers, createContainer(false, context.SideCarName, pod, image, params, params.CreateCommandlineArgumentsForSideCar()))
	}

	return nil
}

// createContainer injects an initContainer
func createContainer(isInitContainer bool, containerName string, pod *corev1.Pod, image string, params context.SynchronizationParameters, commandlineArgs []string) corev1.Container {
	container := corev1.Container{
		Name:         containerName,
		Image:        image,
		Command:      []string{"/usr/bin/volume-syncing-controller"},
		Args:         commandlineArgs,
		WorkingDir:   "/",
		Env:          buildEnvironment(params.Env),
		EnvFrom:      buildSecretRefs(params.EnvSecrets),
		VolumeMounts: mergeVolumeMounts(pod.Spec.Containers, params.LocalPath),
		// VolumeDevices:            nil,
		ImagePullPolicy: "Always",
	}

	if !isInitContainer {
		// lifecycle hook allows to perform a last synchronization before the Pod will be terminated
		container.Lifecycle = &corev1.Lifecycle{
			PreStop: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/usr/bin/volume-syncing-controller", "interrupt"},
				},
			},
		}
	}

	// run container as specified user to operate on volume with given permissions
	if params.Owner != "" && params.Group != "" {
		logrus.Infof("Using UID=%v, GID=%v", params.Owner, params.Group)

		// RunAsNonRoot
		iOwner, _ := strconv.Atoi(params.Owner)
		asNonRoot := iOwner > 0

		// RunAsUser
		iUser, _ := strconv.Atoi(params.Owner)
		runAsUser := int64(iUser)

		// RunAsGroup
		iGroup, _ := strconv.Atoi(params.Group)
		runAsGroup := int64(iGroup)

		// ReadOnlyRootFilesystem
		roFilesystem := false

		container.SecurityContext = &corev1.SecurityContext{
			RunAsUser:              &runAsUser,
			RunAsGroup:             &runAsGroup,
			RunAsNonRoot:           &asNonRoot,
			ReadOnlyRootFilesystem: &roFilesystem,
		}
	}

	return container
}

// buildSecretRefs Builds a `envFrom` to reference all `kind: Secret` objects
func buildSecretRefs(envSecrets []string) []corev1.EnvFromSource {
	var envFrom []corev1.EnvFromSource
	for _, secretName := range envSecrets {
		envFrom = append(envFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
			},
		})
	}
	return envFrom
}

// buildEnvironment Converts a map to Pod's environment syntax
func buildEnvironment(env map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for k, v := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}

// mergeVolumeMounts merges volume mounts of multiple containers
func mergeVolumeMounts(containers []corev1.Container, targetPath string) []corev1.VolumeMount {
	var merged []corev1.VolumeMount
	var appendedPaths []string

	for _, container := range containers {
		for _, volume := range container.VolumeMounts {
			for _, existingPath := range appendedPaths {
				// already collected
				if existingPath == volume.MountPath {
					continue
				}
			}

			// do not collect non-related mount points at all
			if !strings.HasPrefix(targetPath, volume.MountPath) {
				continue
			}

			logrus.Debugf("Collecting VolumeMount: %v", volume.String())
			appendedPaths = append(appendedPaths, volume.MountPath)
			merged = append(merged, volume)
		}
	}

	return merged
}

func hasInitContainer(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.InitContainers {
		if container.Name == context.InitContainerName {
			return true
		}
	}
	return false
}

func hasSideCar(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == context.SideCarName {
			return true
		}
	}
	return false
}

// reorderInitContainer is making sure that the container will be placed in a proper order
func reorderInitContainer(containers []corev1.Container, containerName string, placement v1alpha1.InitContainerPlacementSpec) ([]corev1.Container, error) {
	foundAtIndex := 0
	var copied []corev1.Container // list without our container
	var ourContainer corev1.Container

	// at first find our container
	for num, container := range containers {
		if container.Name == containerName {
			foundAtIndex = num
			ourContainer = container
			continue
		}
		copied = append(copied, container)
	}

	if placement.GetPlacement() == "first" && foundAtIndex != 0 {
		return append([]corev1.Container{ourContainer}, copied...), nil

	} else if placement.GetPlacement() == "last" && foundAtIndex+1 != len(containers) {
		return append(copied, ourContainer), nil

	} else if placement.GetPlacement() == "before" || placement.GetPlacement() == "after" {
		var final []corev1.Container

		for _, container := range copied {
			if container.Name == placement.ContainerReference && placement.GetPlacement() == "before" {
				final = append(final, ourContainer)
			}
			final = append(final, container)
			if container.Name == placement.ContainerReference && placement.GetPlacement() == "after" {
				final = append(final, ourContainer)
			}
		}
		return final, nil
	}

	return containers, nil
}
