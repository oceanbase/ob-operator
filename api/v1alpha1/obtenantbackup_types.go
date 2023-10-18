/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantBackupSpec defines the desired state of OBTenantBackup
type OBTenantBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Type          apitypes.BackupJobType `json:"type"`
	TenantName    string                 `json:"tenantName"`
	TenantSecret  string                 `json:"tenantSecret"`
	ObClusterName string                 `json:"obClusterName"`
	Path          string                 `json:"path,omitempty"`

	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}

// +kubebuilder:object:generate=false
// OBTenantBackupStatus defines the observed state of OBTenantBackup
type OBTenantBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           apitypes.BackupJobStatus `json:"status"`
	Progress         string                   `json:"progress,omitempty"`
	OperationContext *OperationContext        `json:"operationContext,omitempty"`
	StartedAt        string                   `json:"startedAt,omitempty"`
	EndedAt          string                   `json:"endedAt,omitempty"`
	BackupJob        *model.OBBackupJob       `json:"backupJob,omitempty"`
	ArchiveLogJob    *model.OBArchiveLogJob   `json:"archiveLogJob,omitempty"`
	DataCleanJob     *model.OBBackupCleanJob  `json:"dataCleanJob,omitempty"`
}

// fix: implementation of DeepCopyInto needed by zz_generated.deepcopy.go
// controller-gen can not generate DeepCopyInto method for struct with pointer field
func (in *OBTenantBackupStatus) DeepCopyInto(out *OBTenantBackupStatus) {
	*out = *in
	if in.OperationContext != nil {
		in, out := &in.OperationContext, &out.OperationContext
		*out = new(OperationContext)
		(*in).DeepCopyInto(*out)
	}
	if in.BackupJob != nil {
		in, out := &in.BackupJob, &out.BackupJob
		*out = new(model.OBBackupJob)
		**out = **in
	}
	if in.ArchiveLogJob != nil {
		in, out := &in.ArchiveLogJob, &out.ArchiveLogJob
		*out = new(model.OBArchiveLogJob)
		**out = **in
	}
	if in.DataCleanJob != nil {
		in, out := &in.DataCleanJob, &out.DataCleanJob
		*out = new(model.OBBackupCleanJob)
		**out = **in
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="TenantName",type=string,JSONPath=`.spec.tenantName`
//+kubebuilder:printcolumn:name="Path",type=string,JSONPath=`.spec.path`,priority=100
//+kubebuilder:printcolumn:name="StartedAt",type=string,JSONPath=`.status.startedAt`
//+kubebuilder:printcolumn:name="EndedAt",type=string,JSONPath=`.status.endedAt`,description="In ArchiveLogJob, EndedAt is CheckpointScnDisplay field, in other jobs, EndedAt is EndTimestamp field"

// OBTenantBackup is the Schema for the obtenantbackups API.
// An instance of OBTenantBackup stands for a tenant backup job
type OBTenantBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantBackupSpec   `json:"spec,omitempty"`
	Status OBTenantBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBTenantBackupList contains a list of OBTenantBackup
type OBTenantBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenantBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenantBackup{}, &OBTenantBackupList{})
}
