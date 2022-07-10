package mutation

import (
	"github.com/riotkit-org/volume-syncing-controller/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-controller/pkg/server/context"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestMutatePodByInjectingContainers_ReordersContainer(t *testing.T) {
	type testData struct {
		ContainerReference string
		Placement          string
		ExpectedOrder      []string
	}

	// test cases
	table := []testData{
		{
			ContainerReference: "existing",
			Placement:          "before",
			ExpectedOrder:      []string{"init-volume-restore", "existing", "second-existing"},
		},
		{
			ContainerReference: "existing",
			Placement:          "after",
			ExpectedOrder:      []string{"existing", "init-volume-restore", "second-existing"},
		},
		{
			ContainerReference: "second-existing",
			Placement:          "after",
			ExpectedOrder:      []string{"existing", "second-existing", "init-volume-restore"},
		},
		{
			ContainerReference: "",
			Placement:          "last",
			ExpectedOrder:      []string{"existing", "second-existing", "init-volume-restore"},
		},
		{
			ContainerReference: "",
			Placement:          "first",
			ExpectedOrder:      []string{"init-volume-restore", "existing", "second-existing"},
		},
	}

	for _, test := range table {
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "second-pod",
				Namespace: "default",
				Labels: map[string]string{
					"riotkit.org/volume-syncing-controller": "true",
					"my-pod-label":                          "test",
				},
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{Name: "existing"},
					{Name: "second-existing"},
				},
			},
		}

		params := context.SynchronizationParameters{}
		params.InitContainerPlacement.ContainerReference = test.ContainerReference
		params.InitContainerPlacement.Placement = v1alpha1.ContainerPlacement(test.Placement)
		err := MutatePodByInjectingContainers(&pod, "ghcr.io/blahblahblah", true, true, params)

		var containersNames []string
		for _, container := range pod.Spec.InitContainers {
			containersNames = append(containersNames, container.Name)
		}

		assert.Nil(t, err)
		assert.Equal(t, test.ExpectedOrder, containersNames)
	}
}
