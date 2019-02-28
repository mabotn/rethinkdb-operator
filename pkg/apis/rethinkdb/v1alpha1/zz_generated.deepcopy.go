// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RethinkDBCluster) DeepCopyInto(out *RethinkDBCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RethinkDBCluster.
func (in *RethinkDBCluster) DeepCopy() *RethinkDBCluster {
	if in == nil {
		return nil
	}
	out := new(RethinkDBCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RethinkDBCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RethinkDBClusterList) DeepCopyInto(out *RethinkDBClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RethinkDBCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RethinkDBClusterList.
func (in *RethinkDBClusterList) DeepCopy() *RethinkDBClusterList {
	if in == nil {
		return nil
	}
	out := new(RethinkDBClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RethinkDBClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RethinkDBClusterSpec) DeepCopyInto(out *RethinkDBClusterSpec) {
	*out = *in
	if in.Pod != nil {
		in, out := &in.Pod, &out.Pod
		*out = new(RethinkDBPodPolicy)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RethinkDBClusterSpec.
func (in *RethinkDBClusterSpec) DeepCopy() *RethinkDBClusterSpec {
	if in == nil {
		return nil
	}
	out := new(RethinkDBClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RethinkDBClusterStatus) DeepCopyInto(out *RethinkDBClusterStatus) {
	*out = *in
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RethinkDBClusterStatus.
func (in *RethinkDBClusterStatus) DeepCopy() *RethinkDBClusterStatus {
	if in == nil {
		return nil
	}
	out := new(RethinkDBClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RethinkDBPodPolicy) DeepCopyInto(out *RethinkDBPodPolicy) {
	*out = *in
	in.Resources.DeepCopyInto(&out.Resources)
	if in.PersistentVolumeClaimSpec != nil {
		in, out := &in.PersistentVolumeClaimSpec, &out.PersistentVolumeClaimSpec
		*out = new(v1.PersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RethinkDBPodPolicy.
func (in *RethinkDBPodPolicy) DeepCopy() *RethinkDBPodPolicy {
	if in == nil {
		return nil
	}
	out := new(RethinkDBPodPolicy)
	in.DeepCopyInto(out)
	return out
}
