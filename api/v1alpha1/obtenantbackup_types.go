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
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantBackupSpec defines the desired state of OBTenantBackup
type OBTenantBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of OBTenantBackup. Edit obtenantbackup_types.go to remove/update
	Type           BackupJobType `json:"type"`
	TenantName     string        `json:"tenantName"`
	LogArchiveDest string        `json:"logArchiveDest,omitempty"`
	DataBackupDest string        `json:"dataBackupDest,omitempty"`
}

// OBTenantBackupStatus defines the observed state of OBTenantBackup
type OBTenantBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           BackupJobStatus    `json:"status"`
	Progress         string             `json:"progress"`
	OperationContext *OperationContext  `json:"operationContext,omitempty"`
	JobId            int64              `json:"jobId,omitempty"`
	BackupJob        *model.OBBackupJob `json:"backupJob,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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

type BackupJobType string

const (
	BackupJobTypeFull    BackupJobType = "FULL"
	BackupJobTypeIncr    BackupJobType = "INC"
	BackupJobTypeClean   BackupJobType = "CLEAN"
	BackupJobTypeArchive BackupJobType = "ARCHIVE"
)

type BackupJobStatus string

const (
	BackupJobStatusRunning      BackupJobStatus = "RUNNING"
	BackupJobStatusInitializing BackupJobStatus = "INITIALIZING"
	BackupJobStatusSuccessful   BackupJobStatus = "SUCCESSFUL"
	BackupJobStatusFailed       BackupJobStatus = "FAILED"
)
