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

// OBTenantBackupPolicySpec defines the desired state of OBTenantBackupPolicy
type OBTenantBackupPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of OBTenantBackupPolicy. Edit obtenantbackuppolicy_types.go to remove/update
	TenantName string           `json:"tenantName"`
	LogArchive LogArchiveConfig `json:"logArchive"`
	DataBackup DataBackupConfig `json:"dataBackup"`
	DataClean  CleanPolicy      `json:"dataClean"`
}

// OBTenantBackupPolicyStatus defines the observed state of OBTenantBackupPolicy
type OBTenantBackupPolicyStatus struct {
	Status                 BackupPolicyStatusType `json:"status"`
	LogArchiveDestDisabled bool                   `json:"logArchiveDestDisabled"`
	TenantInfo             *model.OBTenant        `json:"tenantInfo"`
	OperationContext       *OperationContext      `json:"operationContext,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBTenantBackupPolicy is the Schema for the obtenantbackuppolicies API
type OBTenantBackupPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBTenantBackupPolicySpec   `json:"spec,omitempty"`
	Status OBTenantBackupPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBTenantBackupPolicyList contains a list of OBTenantBackupPolicy
type OBTenantBackupPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBTenantBackupPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBTenantBackupPolicy{}, &OBTenantBackupPolicyList{})
}

// LogArchiveConfig contains the configuration for log archive progress
type LogArchiveConfig struct {
	Destination         string `json:"destination"`
	SwitchPieceInterval string `json:"switchPieceInterval"`
	DestDisabled        bool   `json:"destDisabled"`
	Concurrency         int    `json:"concurrency"`
}

type BackupType string

const (
	BackupFull BackupType = "FULL"
	BackupIncr BackupType = "INCR"
)

// DataBackupConfig contains the configuration for data backup progress
type DataBackupConfig struct {
	Destination string     `json:"destination"`
	Type        BackupType `json:"type"`
	Crontab     string     `json:"crontab"`
}

type CleanPolicy struct {
	Name          string `json:"name"`
	RecoverWindow string `json:"recoverWindow"`
	Disabled      string `json:"Disabled"`
}

type BackupPolicyStatusType string

const (
	BackupPolicyStatusPreparing BackupPolicyStatusType = "PREPARING"
	BackupPolicyStatusPrepared  BackupPolicyStatusType = "PREPARED"
	BackupPolicyStatusRunning   BackupPolicyStatusType = "RUNNING"
	BackupPolicyStatusFailed    BackupPolicyStatusType = "FAILED"
	BackupPolicyStatusPaused    BackupPolicyStatusType = "PAUSED"
	BackupPolicyStatusStopped   BackupPolicyStatusType = "STOPPED"
)
