//go:build !ignore_autogenerated

/*
Copyright 2023.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/oceanbase/ob-operator/api/types"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CleanPolicy) DeepCopyInto(out *CleanPolicy) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CleanPolicy.
func (in *CleanPolicy) DeepCopy() *CleanPolicy {
	if in == nil {
		return nil
	}
	out := new(CleanPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataBackupConfig) DeepCopyInto(out *DataBackupConfig) {
	*out = *in
	out.Destination = in.Destination
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataBackupConfig.
func (in *DataBackupConfig) DeepCopy() *DataBackupConfig {
	if in == nil {
		return nil
	}
	out := new(DataBackupConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalityType) DeepCopyInto(out *LocalityType) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalityType.
func (in *LocalityType) DeepCopy() *LocalityType {
	if in == nil {
		return nil
	}
	out := new(LocalityType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogArchiveConfig) DeepCopyInto(out *LogArchiveConfig) {
	*out = *in
	out.Destination = in.Destination
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogArchiveConfig.
func (in *LogArchiveConfig) DeepCopy() *LogArchiveConfig {
	if in == nil {
		return nil
	}
	out := new(LogArchiveConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MigrateServerStatus) DeepCopyInto(out *MigrateServerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MigrateServerStatus.
func (in *MigrateServerStatus) DeepCopy() *MigrateServerStatus {
	if in == nil {
		return nil
	}
	out := new(MigrateServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBCluster) DeepCopyInto(out *OBCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBCluster.
func (in *OBCluster) DeepCopy() *OBCluster {
	if in == nil {
		return nil
	}
	out := new(OBCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBClusterList) DeepCopyInto(out *OBClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBClusterList.
func (in *OBClusterList) DeepCopy() *OBClusterList {
	if in == nil {
		return nil
	}
	out := new(OBClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBClusterSpec) DeepCopyInto(out *OBClusterSpec) {
	*out = *in
	if in.OBServerTemplate != nil {
		in, out := &in.OBServerTemplate, &out.OBServerTemplate
		*out = new(OBServerTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.MonitorTemplate != nil {
		in, out := &in.MonitorTemplate, &out.MonitorTemplate
		*out = (*in).DeepCopy()
	}
	if in.BackupVolume != nil {
		in, out := &in.BackupVolume, &out.BackupVolume
		*out = (*in).DeepCopy()
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]types.Parameter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Topology != nil {
		in, out := &in.Topology, &out.Topology
		*out = make([]types.OBZoneTopology, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.UserSecrets != nil {
		in, out := &in.UserSecrets, &out.UserSecrets
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBClusterSpec.
func (in *OBClusterSpec) DeepCopy() *OBClusterSpec {
	if in == nil {
		return nil
	}
	out := new(OBClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBClusterStatus) DeepCopyInto(out *OBClusterStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = (*in).DeepCopy()
	}
	if in.OBZoneStatus != nil {
		in, out := &in.OBZoneStatus, &out.OBZoneStatus
		*out = make([]types.OBZoneReplicaStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make([]types.Parameter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.UserSecrets != nil {
		in, out := &in.UserSecrets, &out.UserSecrets
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBClusterStatus.
func (in *OBClusterStatus) DeepCopy() *OBClusterStatus {
	if in == nil {
		return nil
	}
	out := new(OBClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBParameter) DeepCopyInto(out *OBParameter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBParameter.
func (in *OBParameter) DeepCopy() *OBParameter {
	if in == nil {
		return nil
	}
	out := new(OBParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBParameter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBParameterList) DeepCopyInto(out *OBParameterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBParameter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBParameterList.
func (in *OBParameterList) DeepCopy() *OBParameterList {
	if in == nil {
		return nil
	}
	out := new(OBParameterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBParameterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBParameterSpec) DeepCopyInto(out *OBParameterSpec) {
	*out = *in
	if in.Parameter != nil {
		in, out := &in.Parameter, &out.Parameter
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBParameterSpec.
func (in *OBParameterSpec) DeepCopy() *OBParameterSpec {
	if in == nil {
		return nil
	}
	out := new(OBParameterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBParameterStatus) DeepCopyInto(out *OBParameterStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = (*in).DeepCopy()
	}
	if in.Parameter != nil {
		in, out := &in.Parameter, &out.Parameter
		*out = make([]types.ParameterValue, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBParameterStatus.
func (in *OBParameterStatus) DeepCopy() *OBParameterStatus {
	if in == nil {
		return nil
	}
	out := new(OBParameterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBServer) DeepCopyInto(out *OBServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBServer.
func (in *OBServer) DeepCopy() *OBServer {
	if in == nil {
		return nil
	}
	out := new(OBServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBServerList) DeepCopyInto(out *OBServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBServerList.
func (in *OBServerList) DeepCopy() *OBServerList {
	if in == nil {
		return nil
	}
	out := new(OBServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBServerSpec) DeepCopyInto(out *OBServerSpec) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(v1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.OBServerTemplate != nil {
		in, out := &in.OBServerTemplate, &out.OBServerTemplate
		*out = new(OBServerTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.MonitorTemplate != nil {
		in, out := &in.MonitorTemplate, &out.MonitorTemplate
		*out = (*in).DeepCopy()
	}
	if in.BackupVolume != nil {
		in, out := &in.BackupVolume, &out.BackupVolume
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBServerSpec.
func (in *OBServerSpec) DeepCopy() *OBServerSpec {
	if in == nil {
		return nil
	}
	out := new(OBServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBServerStatus) DeepCopyInto(out *OBServerStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBServerStatus.
func (in *OBServerStatus) DeepCopy() *OBServerStatus {
	if in == nil {
		return nil
	}
	out := new(OBServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBServerTemplate) DeepCopyInto(out *OBServerTemplate) {
	*out = *in
	if in.Resource != nil {
		in, out := &in.Resource, &out.Resource
		*out = (*in).DeepCopy()
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBServerTemplate.
func (in *OBServerTemplate) DeepCopy() *OBServerTemplate {
	if in == nil {
		return nil
	}
	out := new(OBServerTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenant) DeepCopyInto(out *OBTenant) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenant.
func (in *OBTenant) DeepCopy() *OBTenant {
	if in == nil {
		return nil
	}
	out := new(OBTenant)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenant) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackup) DeepCopyInto(out *OBTenantBackup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackup.
func (in *OBTenantBackup) DeepCopy() *OBTenantBackup {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantBackup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackupList) DeepCopyInto(out *OBTenantBackupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBTenantBackup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackupList.
func (in *OBTenantBackupList) DeepCopy() *OBTenantBackupList {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantBackupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackupPolicy) DeepCopyInto(out *OBTenantBackupPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackupPolicy.
func (in *OBTenantBackupPolicy) DeepCopy() *OBTenantBackupPolicy {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackupPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantBackupPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackupPolicyList) DeepCopyInto(out *OBTenantBackupPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBTenantBackupPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackupPolicyList.
func (in *OBTenantBackupPolicyList) DeepCopy() *OBTenantBackupPolicyList {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackupPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantBackupPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackupPolicySpec) DeepCopyInto(out *OBTenantBackupPolicySpec) {
	*out = *in
	out.LogArchive = in.LogArchive
	out.DataBackup = in.DataBackup
	out.DataClean = in.DataClean
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackupPolicySpec.
func (in *OBTenantBackupPolicySpec) DeepCopy() *OBTenantBackupPolicySpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackupPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantBackupSpec) DeepCopyInto(out *OBTenantBackupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantBackupSpec.
func (in *OBTenantBackupSpec) DeepCopy() *OBTenantBackupSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantBackupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantList) DeepCopyInto(out *OBTenantList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBTenant, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantList.
func (in *OBTenantList) DeepCopy() *OBTenantList {
	if in == nil {
		return nil
	}
	out := new(OBTenantList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOpChangePwdSpec) DeepCopyInto(out *OBTenantOpChangePwdSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOpChangePwdSpec.
func (in *OBTenantOpChangePwdSpec) DeepCopy() *OBTenantOpChangePwdSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantOpChangePwdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOpFailoverSpec) DeepCopyInto(out *OBTenantOpFailoverSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOpFailoverSpec.
func (in *OBTenantOpFailoverSpec) DeepCopy() *OBTenantOpFailoverSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantOpFailoverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOpSwitchoverSpec) DeepCopyInto(out *OBTenantOpSwitchoverSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOpSwitchoverSpec.
func (in *OBTenantOpSwitchoverSpec) DeepCopy() *OBTenantOpSwitchoverSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantOpSwitchoverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOperation) DeepCopyInto(out *OBTenantOperation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOperation.
func (in *OBTenantOperation) DeepCopy() *OBTenantOperation {
	if in == nil {
		return nil
	}
	out := new(OBTenantOperation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantOperation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOperationList) DeepCopyInto(out *OBTenantOperationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBTenantOperation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOperationList.
func (in *OBTenantOperationList) DeepCopy() *OBTenantOperationList {
	if in == nil {
		return nil
	}
	out := new(OBTenantOperationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantOperationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOperationSpec) DeepCopyInto(out *OBTenantOperationSpec) {
	*out = *in
	if in.Switchover != nil {
		in, out := &in.Switchover, &out.Switchover
		*out = new(OBTenantOpSwitchoverSpec)
		**out = **in
	}
	if in.Failover != nil {
		in, out := &in.Failover, &out.Failover
		*out = new(OBTenantOpFailoverSpec)
		**out = **in
	}
	if in.ChangePwd != nil {
		in, out := &in.ChangePwd, &out.ChangePwd
		*out = new(OBTenantOpChangePwdSpec)
		**out = **in
	}
	if in.ReplayUntil != nil {
		in, out := &in.ReplayUntil, &out.ReplayUntil
		*out = new(RestoreUntilConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.TargetTenant != nil {
		in, out := &in.TargetTenant, &out.TargetTenant
		*out = new(string)
		**out = **in
	}
	if in.AuxillaryTenant != nil {
		in, out := &in.AuxillaryTenant, &out.AuxillaryTenant
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOperationSpec.
func (in *OBTenantOperationSpec) DeepCopy() *OBTenantOperationSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantOperationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantOperationStatus) DeepCopyInto(out *OBTenantOperationStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = (*in).DeepCopy()
	}
	if in.PrimaryTenant != nil {
		in, out := &in.PrimaryTenant, &out.PrimaryTenant
		*out = new(OBTenant)
		(*in).DeepCopyInto(*out)
	}
	if in.SecondaryTenant != nil {
		in, out := &in.SecondaryTenant, &out.SecondaryTenant
		*out = new(OBTenant)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantOperationStatus.
func (in *OBTenantOperationStatus) DeepCopy() *OBTenantOperationStatus {
	if in == nil {
		return nil
	}
	out := new(OBTenantOperationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantRestore) DeepCopyInto(out *OBTenantRestore) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantRestore.
func (in *OBTenantRestore) DeepCopy() *OBTenantRestore {
	if in == nil {
		return nil
	}
	out := new(OBTenantRestore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantRestore) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantRestoreList) DeepCopyInto(out *OBTenantRestoreList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBTenantRestore, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantRestoreList.
func (in *OBTenantRestoreList) DeepCopy() *OBTenantRestoreList {
	if in == nil {
		return nil
	}
	out := new(OBTenantRestoreList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBTenantRestoreList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantRestoreSpec) DeepCopyInto(out *OBTenantRestoreSpec) {
	*out = *in
	in.Source.DeepCopyInto(&out.Source)
	if in.PrimaryTenant != nil {
		in, out := &in.PrimaryTenant, &out.PrimaryTenant
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantRestoreSpec.
func (in *OBTenantRestoreSpec) DeepCopy() *OBTenantRestoreSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantRestoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBTenantSpec) DeepCopyInto(out *OBTenantSpec) {
	*out = *in
	if in.Pools != nil {
		in, out := &in.Pools, &out.Pools
		*out = make([]ResourcePoolSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Source != nil {
		in, out := &in.Source, &out.Source
		*out = new(TenantSourceSpec)
		(*in).DeepCopyInto(*out)
	}
	out.Credentials = in.Credentials
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBTenantSpec.
func (in *OBTenantSpec) DeepCopy() *OBTenantSpec {
	if in == nil {
		return nil
	}
	out := new(OBTenantSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBZone) DeepCopyInto(out *OBZone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBZone.
func (in *OBZone) DeepCopy() *OBZone {
	if in == nil {
		return nil
	}
	out := new(OBZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBZone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBZoneList) DeepCopyInto(out *OBZoneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OBZone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBZoneList.
func (in *OBZoneList) DeepCopy() *OBZoneList {
	if in == nil {
		return nil
	}
	out := new(OBZoneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OBZoneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBZoneSpec) DeepCopyInto(out *OBZoneSpec) {
	*out = *in
	in.Topology.DeepCopyInto(&out.Topology)
	if in.OBServerTemplate != nil {
		in, out := &in.OBServerTemplate, &out.OBServerTemplate
		*out = new(OBServerTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.MonitorTemplate != nil {
		in, out := &in.MonitorTemplate, &out.MonitorTemplate
		*out = (*in).DeepCopy()
	}
	if in.BackupVolume != nil {
		in, out := &in.BackupVolume, &out.BackupVolume
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBZoneSpec.
func (in *OBZoneSpec) DeepCopy() *OBZoneSpec {
	if in == nil {
		return nil
	}
	out := new(OBZoneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OBZoneStatus) DeepCopyInto(out *OBZoneStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = (*in).DeepCopy()
	}
	if in.OBServerStatus != nil {
		in, out := &in.OBServerStatus, &out.OBServerStatus
		*out = make([]types.OBServerReplicaStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OBZoneStatus.
func (in *OBZoneStatus) DeepCopy() *OBZoneStatus {
	if in == nil {
		return nil
	}
	out := new(OBZoneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcePoolSpec) DeepCopyInto(out *ResourcePoolSpec) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(LocalityType)
		**out = **in
	}
	if in.UnitConfig != nil {
		in, out := &in.UnitConfig, &out.UnitConfig
		*out = new(UnitConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcePoolSpec.
func (in *ResourcePoolSpec) DeepCopy() *ResourcePoolSpec {
	if in == nil {
		return nil
	}
	out := new(ResourcePoolSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourcePoolStatus) DeepCopyInto(out *ResourcePoolStatus) {
	*out = *in
	if in.Units != nil {
		in, out := &in.Units, &out.Units
		*out = make([]UnitStatus, len(*in))
		copy(*out, *in)
	}
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(LocalityType)
		**out = **in
	}
	if in.UnitConfig != nil {
		in, out := &in.UnitConfig, &out.UnitConfig
		*out = new(UnitConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourcePoolStatus.
func (in *ResourcePoolStatus) DeepCopy() *ResourcePoolStatus {
	if in == nil {
		return nil
	}
	out := new(ResourcePoolStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestoreSourceSpec) DeepCopyInto(out *RestoreSourceSpec) {
	*out = *in
	if in.ArchiveSource != nil {
		in, out := &in.ArchiveSource, &out.ArchiveSource
		*out = new(types.BackupDestination)
		**out = **in
	}
	if in.BakDataSource != nil {
		in, out := &in.BakDataSource, &out.BakDataSource
		*out = new(types.BackupDestination)
		**out = **in
	}
	in.Until.DeepCopyInto(&out.Until)
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.ReplayLogUntil != nil {
		in, out := &in.ReplayLogUntil, &out.ReplayLogUntil
		*out = new(RestoreUntilConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestoreSourceSpec.
func (in *RestoreSourceSpec) DeepCopy() *RestoreSourceSpec {
	if in == nil {
		return nil
	}
	out := new(RestoreSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RestoreUntilConfig) DeepCopyInto(out *RestoreUntilConfig) {
	*out = *in
	if in.Timestamp != nil {
		in, out := &in.Timestamp, &out.Timestamp
		*out = new(string)
		**out = **in
	}
	if in.Scn != nil {
		in, out := &in.Scn, &out.Scn
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RestoreUntilConfig.
func (in *RestoreUntilConfig) DeepCopy() *RestoreUntilConfig {
	if in == nil {
		return nil
	}
	out := new(RestoreUntilConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantCredentials) DeepCopyInto(out *TenantCredentials) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantCredentials.
func (in *TenantCredentials) DeepCopy() *TenantCredentials {
	if in == nil {
		return nil
	}
	out := new(TenantCredentials)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantRecordInfo) DeepCopyInto(out *TenantRecordInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantRecordInfo.
func (in *TenantRecordInfo) DeepCopy() *TenantRecordInfo {
	if in == nil {
		return nil
	}
	out := new(TenantRecordInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantSourceSpec) DeepCopyInto(out *TenantSourceSpec) {
	*out = *in
	if in.Tenant != nil {
		in, out := &in.Tenant, &out.Tenant
		*out = new(string)
		**out = **in
	}
	if in.Restore != nil {
		in, out := &in.Restore, &out.Restore
		*out = new(RestoreSourceSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantSourceSpec.
func (in *TenantSourceSpec) DeepCopy() *TenantSourceSpec {
	if in == nil {
		return nil
	}
	out := new(TenantSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantSourceStatus.
func (in *TenantSourceStatus) DeepCopy() *TenantSourceStatus {
	if in == nil {
		return nil
	}
	out := new(TenantSourceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UnitConfig) DeepCopyInto(out *UnitConfig) {
	*out = *in
	out.MaxCPU = in.MaxCPU.DeepCopy()
	out.MemorySize = in.MemorySize.DeepCopy()
	out.MinCPU = in.MinCPU.DeepCopy()
	out.LogDiskSize = in.LogDiskSize.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnitConfig.
func (in *UnitConfig) DeepCopy() *UnitConfig {
	if in == nil {
		return nil
	}
	out := new(UnitConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UnitStatus) DeepCopyInto(out *UnitStatus) {
	*out = *in
	out.Migrate = in.Migrate
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnitStatus.
func (in *UnitStatus) DeepCopy() *UnitStatus {
	if in == nil {
		return nil
	}
	out := new(UnitStatus)
	in.DeepCopyInto(out)
	return out
}
