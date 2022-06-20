//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomaticEncryptionSpec) DeepCopyInto(out *AutomaticEncryptionSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomaticEncryptionSpec.
func (in *AutomaticEncryptionSpec) DeepCopy() *AutomaticEncryptionSpec {
	if in == nil {
		return nil
	}
	out := new(AutomaticEncryptionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CleanUpSpec) DeepCopyInto(out *CleanUpSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CleanUpSpec.
func (in *CleanUpSpec) DeepCopy() *CleanUpSpec {
	if in == nil {
		return nil
	}
	out := new(CleanUpSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PermissionsSpec) DeepCopyInto(out *PermissionsSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PermissionsSpec.
func (in *PermissionsSpec) DeepCopy() *PermissionsSpec {
	if in == nil {
		return nil
	}
	out := new(PermissionsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in PodEnvironment) DeepCopyInto(out *PodEnvironment) {
	{
		in := &in
		*out = make(PodEnvironment, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodEnvironment.
func (in PodEnvironment) DeepCopy() PodEnvironment {
	if in == nil {
		return nil
	}
	out := new(PodEnvironment)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in PodEnvironmentFromSecrets) DeepCopyInto(out *PodEnvironmentFromSecrets) {
	{
		in := &in
		*out = make(PodEnvironmentFromSecrets, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodEnvironmentFromSecrets.
func (in PodEnvironmentFromSecrets) DeepCopy() PodEnvironmentFromSecrets {
	if in == nil {
		return nil
	}
	out := new(PodEnvironmentFromSecrets)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodFilesystemSync) DeepCopyInto(out *PodFilesystemSync) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodFilesystemSync.
func (in *PodFilesystemSync) DeepCopy() *PodFilesystemSync {
	if in == nil {
		return nil
	}
	out := new(PodFilesystemSync)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodFilesystemSync) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodFilesystemSyncList) DeepCopyInto(out *PodFilesystemSyncList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PodFilesystemSync, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodFilesystemSyncList.
func (in *PodFilesystemSyncList) DeepCopy() *PodFilesystemSyncList {
	if in == nil {
		return nil
	}
	out := new(PodFilesystemSyncList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PodFilesystemSyncList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodFilesystemSyncSpec) DeepCopyInto(out *PodFilesystemSyncSpec) {
	*out = *in
	if in.PodSelector != nil {
		in, out := &in.PodSelector, &out.PodSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	out.SyncOptions = in.SyncOptions
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make(PodEnvironment, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.EnvFromSecrets != nil {
		in, out := &in.EnvFromSecrets, &out.EnvFromSecrets
		*out = make(PodEnvironmentFromSecrets, len(*in))
		copy(*out, *in)
	}
	out.AutomaticEncryption = in.AutomaticEncryption
	out.CleanUp = in.CleanUp
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodFilesystemSyncSpec.
func (in *PodFilesystemSyncSpec) DeepCopy() *PodFilesystemSyncSpec {
	if in == nil {
		return nil
	}
	out := new(PodFilesystemSyncSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in PodSelector) DeepCopyInto(out *PodSelector) {
	{
		in := &in
		*out = make(PodSelector, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSelector.
func (in PodSelector) DeepCopy() PodSelector {
	if in == nil {
		return nil
	}
	out := new(PodSelector)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncOptionsSpec) DeepCopyInto(out *SyncOptionsSpec) {
	*out = *in
	out.Permissions = in.Permissions
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncOptionsSpec.
func (in *SyncOptionsSpec) DeepCopy() *SyncOptionsSpec {
	if in == nil {
		return nil
	}
	out := new(SyncOptionsSpec)
	in.DeepCopyInto(out)
	return out
}