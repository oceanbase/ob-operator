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
	constants "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBTenantBackupPolicySpec defines the desired state of OBTenantBackupPolicy
type OBTenantBackupPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ObClusterName string           `json:"obClusterName"`
	TenantName    string           `json:"tenantName"`
	LogArchive    LogArchiveConfig `json:"logArchive"`
	DataBackup    DataBackupConfig `json:"dataBackup"`
	DataClean     CleanPolicy      `json:"dataClean,omitempty"`
}

// OBTenantBackupPolicyStatus defines the observed state of OBTenantBackupPolicy
type OBTenantBackupPolicyStatus struct {
	Status                 constants.BackupPolicyStatusType `json:"status"`
	LogArchiveDestDisabled bool                             `json:"logArchiveDestDisabled"`
	TenantInfo             *model.OBTenant                  `json:"tenantInfo,omitempty"`
	OperationContext       *OperationContext                `json:"operationContext,omitempty"`
	NextFull               string                           `json:"nextFull,omitempty"`
	NextIncremental        string                           `json:"nextIncremental,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
//+kubebuilder:printcolumn:name="TenantName",type=string,JSONPath=`.spec.tenantName`
//+kubebuilder:printcolumn:name="NextFull",type=string,JSONPath=`.status.nextFull`
//+kubebuilder:printcolumn:name="NextIncremental",type=string,JSONPath=`.status.nextIncremental`
//+kubebuilder:printcolumn:name="FullCrontab",type=string,JSONPath=`.spec.dataBackup.fullCrontab`
//+kubebuilder:printcolumn:name="IncrementalCrontab",type=string,JSONPath=`.spec.dataBackup.incrementalCrontab`

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
	Destination         constants.BackupDestination `json:"destination"`
	SwitchPieceInterval string                      `json:"switchPieceInterval"`
	Binding             ArchiveBinding              `json:"binding,omitempty"`
	DestDisabled        bool                        `json:"destDisabled,omitempty"`
	Concurrency         int                         `json:"concurrency,omitempty"`
}

type ArchiveBinding string

const (
	ArchiveBindingOptional  = "Optional"
	ArchiveBindingMandatory = "Mandatory"
)

// DataBackupConfig contains the configuration for data backup progress
type DataBackupConfig struct {
	Destination        constants.BackupDestination `json:"destination"`
	FullCrontab        string                      `json:"fullCrontab,omitempty"`
	IncrementalCrontab string                      `json:"incrementalCrontab,omitempty"`
}

type CleanPolicy struct {
	Name           string `json:"name,omitempty"`
	RecoveryWindow string `json:"recoveryWindow,omitempty"`
	Disabled       string `json:"disabled,omitempty"`
}
