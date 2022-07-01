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
