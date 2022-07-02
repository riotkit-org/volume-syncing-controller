package mutation

import (
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/context"
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
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, createContainer(context.InitContainerName, pod, image, params, params.CreateCommandlineArgumentsForInitContainer()))
	}

	if canSynchronizeToRemote && !hasSideCar(pod) {
		nLogger.Infof("`kind: Pod` '%s' has no side car present", pod.ObjectMeta.Name)
		pod.Spec.Containers = append(pod.Spec.Containers, createContainer(context.SideCarName, pod, image, params, params.CreateCommandlineArgumentsForSideCar()))
	}

	return nil
}

// createContainer injects an initContainer
func createContainer(containerName string, pod *corev1.Pod, image string, params context.SynchronizationParameters, commandlineArgs []string) corev1.Container {
	container := corev1.Container{
		Name:         containerName,
		Image:        image,
		Command:      []string{"/usr/bin/volume-syncing-operator"},
		Args:         commandlineArgs,
		WorkingDir:   "/",
		Env:          buildEnvironment(params.Env),
		EnvFrom:      buildSecretRefs(params.EnvSecrets),
		VolumeMounts: mergeVolumeMounts(pod.Spec.Containers, params.LocalPath),
		// VolumeDevices:            nil,
		ImagePullPolicy: "Always",
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
			Prefix:       "",
			ConfigMapRef: nil,
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
