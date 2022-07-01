package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestPodFilesystemSync_IsPodMatching(t *testing.T) {
	// case A
	definitionA := PodFilesystemSync{}
	definitionA.Spec.PodSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"variant": "with-dynamic-directory-name",
		},
	}
	firstPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"variant":                             "with-dynamic-directory-name",
			},
		},
	}

	// case B
	definitionB := PodFilesystemSync{}
	definitionB.Spec.PodSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"my-pod-label": "test",
		},
	}
	secondPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"my-pod-label":                        "test",
			},
		},
	}

	assert.True(t, definitionA.IsPodMatching(&firstPod))
	assert.False(t, definitionA.IsPodMatching(&secondPod))

	assert.True(t, definitionB.IsPodMatching(&secondPod))
	assert.False(t, definitionB.IsPodMatching(&firstPod))
}

func TestPodFilesystemSync_WasAlreadySynchronized(t *testing.T) {
	testPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lenin-was-a-dickhead",
			Namespace: "kronstadt",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"variant":                             "with-dynamic-directory-name",
			},
		},
	}

	definition := PodFilesystemSync{}
	definition.Spec.RemotePath = "/something"

	// check if at the beginning it does not give false result
	result, _ := definition.WasAlreadySynchronized(&testPod)
	assert.False(t, result)

	// claim it
	assert.Nil(t, definition.ClaimDirectoryByPod(&testPod))

	// after claimed it should be in the status
	result, _ = definition.WasAlreadySynchronized(&testPod)
	assert.True(t, result)

	// check status if was populated
	assert.Len(t, definition.Status.Locations, 1)
}

func TestPodFilesystemSync_ShouldRestoreFilesFromRemote_DisabledDirection(t *testing.T) {
	testPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "lenin-was-a-dickhead",
			Namespace: "kronstadt",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"variant":                             "with-dynamic-directory-name",
			},
		},
	}

	definition := PodFilesystemSync{}
	definition.Spec.SyncOptions.AllowedDirections.FromRemote = false

	result, err := definition.ShouldRestoreFilesFromRemote(&testPod)

	assert.Nil(t, err)
	assert.False(t, result)
}

func TestPodFilesystemSync_ShouldRestoreFilesFromRemote_WhenAlreadySynchronized(t *testing.T) {
	// assumptions:
	//   .Spec.SyncOptions.AllowedDirections.FromRemote = true
	//   wasAlreadySynchronized = true (claimed)
	//

	definition := PodFilesystemSync{}
	definition.Spec.SyncOptions.AllowedDirections.FromRemote = true
	definition.Spec.RemotePath = "/lenin-was-a-dickhead"
	assert.Nil(t, definition.ClaimDirectoryByPod(&v1.Pod{}))

	result, err := definition.ShouldRestoreFilesFromRemote(&v1.Pod{})

	assert.Nil(t, err)
	assert.True(t, result)
}

func TestPodFilesystemSync_ShouldRestoreFilesFromRemote_NotSynchronizedButShouldBeRestoredAtStartup(t *testing.T) {
	// assumptions:
	//   .Spec.SyncOptions.AllowedDirections.FromRemote = true
	//   .Spec.SyncOptions.RestoreRemoteOnFirstRun = true
	//   wasAlreadySynchronized = false (no claim)
	//

	definition := PodFilesystemSync{}
	definition.Spec.SyncOptions.AllowedDirections.FromRemote = true
	definition.Spec.SyncOptions.RestoreRemoteOnFirstRun = true
	definition.Spec.RemotePath = "/lenin-was-a-dickhead"

	result, err := definition.ShouldRestoreFilesFromRemote(&v1.Pod{})

	assert.Nil(t, err)
	assert.True(t, result)
}
