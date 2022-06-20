package cache

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-operator/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Cache struct {
	specsIndexed map[string]v1alpha1.PodFilesystemSync
}

// Populate is fetching initially a list of existing objects before the application was started
func (c Cache) Populate(riotkitClient *versioned.Clientset, kubeClient *kubernetes.Clientset) error {
	ctx := context.TODO()

	namespaces, err := kubeClient.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Cannot list namespaces to populate PodFilesystemSync objects")
	}

	// from every namespace collect "PodFilesystemSync" type objects into the local cache
	for _, ns := range namespaces.Items {
		objects, listingErr := riotkitClient.RiotkitV1alpha1().PodFilesystemSyncs(ns.Name).List(ctx, v1.ListOptions{})
		if listingErr != nil {
			return errors.Wrapf(listingErr, "Cannot list PodFilesystemSync objects inside '%v' namespace. Are there permission issues maybe?", ns.Name)
		}
		logrus.Debugf("[%s] Collecting %v resources of PodFilesystemSync type", ns.Name, len(objects.Items))
		for _, podFilesystemSync := range objects.Items {
			c.Add(podFilesystemSync)
		}
	}
	return nil
}

// Add adds element to cache, making sure it will not be duplicated
func (c *Cache) Add(element v1alpha1.PodFilesystemSync) {
	logrus.Infof("[%s] Updating cache for '%s'", element.Namespace, element.Name)
	indent := c.createCacheIdent(element)
	c.specsIndexed[indent] = element
}

func (c *Cache) createCacheIdent(element v1alpha1.PodFilesystemSync) string {
	return fmt.Sprintf("%v_%v", element.Namespace, element.Name)
}

func (c *Cache) FindMatchingForPod(pod corev1.Pod) (v1alpha1.PodFilesystemSync, error, bool) {
	var matched v1alpha1.PodFilesystemSync
	found := false

	for _, definition := range c.specsIndexed {
		if found {
			return v1alpha1.PodFilesystemSync{}, errors.New("ambiguous match. At least two `kind: PodFilesystemSync` objects are matching the same `kind: Pod` using PodSelector"), false
		}
		if definition.IsPodMatching(pod) {
			matched = definition
			found = true
		}
	}

	return matched, nil, found
}
