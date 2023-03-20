/*
Copyright © 2023 The vm-crd-create Authors.

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
	v1alpha1 "vm-crd-create/pkg/apis/cloudnative/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVirtualMachines implements VirtualMachineInterface
type FakeVirtualMachines struct {
	Fake *FakeCloudnativeV1alpha1
	ns   string
}

var virtualmachinesResource = schema.GroupVersionResource{Group: "cloudnative.lich", Version: "v1alpha1", Resource: "virtualmachines"}

var virtualmachinesKind = schema.GroupVersionKind{Group: "cloudnative.lich", Version: "v1alpha1", Kind: "VirtualMachine"}

// Get takes name of the virtualMachine, and returns the corresponding virtualMachine object, and an error if there is any.
func (c *FakeVirtualMachines) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VirtualMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(virtualmachinesResource, c.ns, name), &v1alpha1.VirtualMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtualMachine), err
}

// List takes label and field selectors, and returns the list of VirtualMachines that match those selectors.
func (c *FakeVirtualMachines) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VirtualMachineList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(virtualmachinesResource, virtualmachinesKind, c.ns, opts), &v1alpha1.VirtualMachineList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VirtualMachineList{ListMeta: obj.(*v1alpha1.VirtualMachineList).ListMeta}
	for _, item := range obj.(*v1alpha1.VirtualMachineList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested virtualMachines.
func (c *FakeVirtualMachines) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(virtualmachinesResource, c.ns, opts))

}

// Create takes the representation of a virtualMachine and creates it.  Returns the server's representation of the virtualMachine, and an error, if there is any.
func (c *FakeVirtualMachines) Create(ctx context.Context, virtualMachine *v1alpha1.VirtualMachine, opts v1.CreateOptions) (result *v1alpha1.VirtualMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(virtualmachinesResource, c.ns, virtualMachine), &v1alpha1.VirtualMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtualMachine), err
}

// Update takes the representation of a virtualMachine and updates it. Returns the server's representation of the virtualMachine, and an error, if there is any.
func (c *FakeVirtualMachines) Update(ctx context.Context, virtualMachine *v1alpha1.VirtualMachine, opts v1.UpdateOptions) (result *v1alpha1.VirtualMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(virtualmachinesResource, c.ns, virtualMachine), &v1alpha1.VirtualMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtualMachine), err
}

// Delete takes name of the virtualMachine and deletes it. Returns an error if one occurs.
func (c *FakeVirtualMachines) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(virtualmachinesResource, c.ns, name, opts), &v1alpha1.VirtualMachine{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVirtualMachines) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(virtualmachinesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.VirtualMachineList{})
	return err
}

// Patch applies the patch and returns the patched virtualMachine.
func (c *FakeVirtualMachines) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VirtualMachine, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(virtualmachinesResource, c.ns, name, pt, data, subresources...), &v1alpha1.VirtualMachine{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VirtualMachine), err
}
