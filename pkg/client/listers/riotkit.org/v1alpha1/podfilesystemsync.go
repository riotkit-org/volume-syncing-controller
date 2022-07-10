/*
Copyright Riotkit.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/riotkit-org/volume-syncing-controller/pkg/apis/riotkit.org/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// PodFilesystemSyncLister helps list PodFilesystemSyncs.
// All objects returned here must be treated as read-only.
type PodFilesystemSyncLister interface {
	// List lists all PodFilesystemSyncs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PodFilesystemSync, err error)
	// PodFilesystemSyncs returns an object that can list and get PodFilesystemSyncs.
	PodFilesystemSyncs(namespace string) PodFilesystemSyncNamespaceLister
	PodFilesystemSyncListerExpansion
}

// podFilesystemSyncLister implements the PodFilesystemSyncLister interface.
type podFilesystemSyncLister struct {
	indexer cache.Indexer
}

// NewPodFilesystemSyncLister returns a new PodFilesystemSyncLister.
func NewPodFilesystemSyncLister(indexer cache.Indexer) PodFilesystemSyncLister {
	return &podFilesystemSyncLister{indexer: indexer}
}

// List lists all PodFilesystemSyncs in the indexer.
func (s *podFilesystemSyncLister) List(selector labels.Selector) (ret []*v1alpha1.PodFilesystemSync, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PodFilesystemSync))
	})
	return ret, err
}

// PodFilesystemSyncs returns an object that can list and get PodFilesystemSyncs.
func (s *podFilesystemSyncLister) PodFilesystemSyncs(namespace string) PodFilesystemSyncNamespaceLister {
	return podFilesystemSyncNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// PodFilesystemSyncNamespaceLister helps list and get PodFilesystemSyncs.
// All objects returned here must be treated as read-only.
type PodFilesystemSyncNamespaceLister interface {
	// List lists all PodFilesystemSyncs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.PodFilesystemSync, err error)
	// Get retrieves the PodFilesystemSync from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.PodFilesystemSync, error)
	PodFilesystemSyncNamespaceListerExpansion
}

// podFilesystemSyncNamespaceLister implements the PodFilesystemSyncNamespaceLister
// interface.
type podFilesystemSyncNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all PodFilesystemSyncs in the indexer for a given namespace.
func (s podFilesystemSyncNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.PodFilesystemSync, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.PodFilesystemSync))
	})
	return ret, err
}

// Get retrieves the PodFilesystemSync from the indexer for a given namespace and name.
func (s podFilesystemSyncNamespaceLister) Get(name string) (*v1alpha1.PodFilesystemSync, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("podfilesystemsync"), name)
	}
	return obj.(*v1alpha1.PodFilesystemSync), nil
}
