package cache_test

import (
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-operator/pkg/client/clientset/versioned/fake"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/cache"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1Fake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func createData() (v1alpha1.PodFilesystemSync, v1alpha1.PodFilesystemSync, v1.Pod, v1.Pod, v1.Namespace) {
	// case A
	definitionA := v1alpha1.PodFilesystemSync{}
	definitionA.ObjectMeta.Name = "definition-a"
	definitionA.ObjectMeta.Namespace = "default"
	definitionA.Spec.PodSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"variant": "with-dynamic-directory-name",
		},
	}
	firstPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "first-pod",
			Namespace: "default",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"variant":                             "with-dynamic-directory-name",
			},
		},
	}

	// case B
	definitionB := v1alpha1.PodFilesystemSync{}
	definitionB.ObjectMeta.Name = "definition-b"
	definitionB.ObjectMeta.Namespace = "default"
	definitionB.Spec.PodSelector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"my-pod-label": "test",
		},
	}
	secondPod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "second-pod",
			Namespace: "default",
			Labels: map[string]string{
				"riotkit.org/volume-syncing-operator": "true",
				"my-pod-label":                        "test",
			},
		},
	}

	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}

	return definitionA, definitionB, firstPod, secondPod, ns
}

// TestCache_FunctionalTest is more functional test that checks working of a whole Cache struct
func TestCache_FunctionalTest(t *testing.T) {
	definitionA, definitionB, firstPod, secondPod, ns := createData()

	rktClient := fake.NewSimpleClientset(&v1alpha1.PodFilesystemSyncList{Items: []v1alpha1.PodFilesystemSync{
		definitionA, definitionB,
	}})
	v1Client := v1Fake.NewSimpleClientset(&firstPod, &secondPod, &ns)

	c := cache.Cache{}

	// populate the cache from API Server
	assert.Nil(t, c.Populate(rktClient, v1Client))

	// then try to delete
	c.Delete(&definitionA)
	c.Delete(&definitionB)

	// and add again
	c.Add(&definitionA)
	c.Add(&definitionB)

	// then try to use those definitions
	_, firstFound, firstErr := c.FindMatchingForPod(&firstPod)
	assert.Nil(t, firstErr)
	assert.True(t, firstFound)

	_, secondFound, secondErr := c.FindMatchingForPod(&firstPod)
	assert.Nil(t, secondErr)
	assert.True(t, secondFound)
}
