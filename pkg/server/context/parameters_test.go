package context

import (
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestNewSynchronizationParameters_UIDAndGIDCanBeSetWithAnnotations is checking that Pod can override values using annotations
func TestNewSynchronizationParameters_UIDAndGIDCanBeSetWithAnnotations(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-pod",
			Namespace: "default",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
			},
			Annotations: map[string]string{
				"riotkit.org/volume-user-id":  "1312",
				"riotkit.org/volume-group-id": "161",
			},
		},
	}

	definition := v1alpha1.PodFilesystemSync{}

	params, err := NewSynchronizationParameters(&pod, &definition, map[string]string{})
	assert.Nil(t, err)

	assert.Equal(t, "161", params.Group)
	assert.Equal(t, "1312", params.Owner)
}

func TestNewSynchronizationParameters_ResolvesRemotePathFromTemplate(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-pod",
			Namespace: "default",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"some":                                "hitler-was-a-dickhead",
			},
		},
	}
	definition := v1alpha1.PodFilesystemSync{}
	definition.Spec.RemotePath = "/mnt/{{ pod.ObjectMeta.Labels[\"some\"] }}/and-lenin-too"

	params, err := NewSynchronizationParameters(&pod, &definition, map[string]string{})
	assert.Nil(t, err)
	assert.Equal(t, "/mnt/hitler-was-a-dickhead/and-lenin-too", params.RemotePath)
}
