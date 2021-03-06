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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/riotkit-org/volume-syncing-controller/pkg/apis/riotkit.org/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePodFilesystemSyncs implements PodFilesystemSyncInterface
type FakePodFilesystemSyncs struct {
	Fake *FakeRiotkitV1alpha1
	ns   string
}

var podfilesystemsyncsResource = schema.GroupVersionResource{Group: "riotkit.org", Version: "v1alpha1", Resource: "podfilesystemsyncs"}

var podfilesystemsyncsKind = schema.GroupVersionKind{Group: "riotkit.org", Version: "v1alpha1", Kind: "PodFilesystemSync"}

// Get takes name of the podFilesystemSync, and returns the corresponding podFilesystemSync object, and an error if there is any.
func (c *FakePodFilesystemSyncs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PodFilesystemSync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podfilesystemsyncsResource, c.ns, name), &v1alpha1.PodFilesystemSync{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodFilesystemSync), err
}

// List takes label and field selectors, and returns the list of PodFilesystemSyncs that match those selectors.
func (c *FakePodFilesystemSyncs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PodFilesystemSyncList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podfilesystemsyncsResource, podfilesystemsyncsKind, c.ns, opts), &v1alpha1.PodFilesystemSyncList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PodFilesystemSyncList{ListMeta: obj.(*v1alpha1.PodFilesystemSyncList).ListMeta}
	for _, item := range obj.(*v1alpha1.PodFilesystemSyncList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podFilesystemSyncs.
func (c *FakePodFilesystemSyncs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podfilesystemsyncsResource, c.ns, opts))

}

// Create takes the representation of a podFilesystemSync and creates it.  Returns the server's representation of the podFilesystemSync, and an error, if there is any.
func (c *FakePodFilesystemSyncs) Create(ctx context.Context, podFilesystemSync *v1alpha1.PodFilesystemSync, opts v1.CreateOptions) (result *v1alpha1.PodFilesystemSync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podfilesystemsyncsResource, c.ns, podFilesystemSync), &v1alpha1.PodFilesystemSync{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodFilesystemSync), err
}

// Update takes the representation of a podFilesystemSync and updates it. Returns the server's representation of the podFilesystemSync, and an error, if there is any.
func (c *FakePodFilesystemSyncs) Update(ctx context.Context, podFilesystemSync *v1alpha1.PodFilesystemSync, opts v1.UpdateOptions) (result *v1alpha1.PodFilesystemSync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podfilesystemsyncsResource, c.ns, podFilesystemSync), &v1alpha1.PodFilesystemSync{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodFilesystemSync), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePodFilesystemSyncs) UpdateStatus(ctx context.Context, podFilesystemSync *v1alpha1.PodFilesystemSync, opts v1.UpdateOptions) (*v1alpha1.PodFilesystemSync, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podfilesystemsyncsResource, "status", c.ns, podFilesystemSync), &v1alpha1.PodFilesystemSync{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodFilesystemSync), err
}

// Delete takes name of the podFilesystemSync and deletes it. Returns an error if one occurs.
func (c *FakePodFilesystemSyncs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(podfilesystemsyncsResource, c.ns, name), &v1alpha1.PodFilesystemSync{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodFilesystemSyncs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podfilesystemsyncsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PodFilesystemSyncList{})
	return err
}

// Patch applies the patch and returns the patched podFilesystemSync.
func (c *FakePodFilesystemSyncs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PodFilesystemSync, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podfilesystemsyncsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PodFilesystemSync{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodFilesystemSync), err
}
