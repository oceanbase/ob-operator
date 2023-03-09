/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantBackupSpec defines the desired state of TenantBackup
type TenantBackupSpec struct {
	SourceCluster      SourceClusterSpec        `json:"source"`
	Tenants            []TenantConfigSpec       `json:"tenant"`
	DeleteBackupPolicy []DeleteBackupPolicySpec `json:"deleteBackupPolicy,omitempty"`
}

type TenantConfigSpec struct {
	Name                string         `json:"name"`
	UserSecret          string         `json:"userSecret"`
	BackupSecret        string         `json:"backupSecret,omitempty"`
	LogArchiveDest      string         `json:"logArchiveDest"`
	Binding             string         `json:"binding,omitempty"`
	PieceSwitchInterval string         `json:"pieceSwitchInterval,omitempty"`
	DataBackupDest      string         `json:"dataBackupDest"`
	Schedule            []ScheduleSpec `json:"schedule"`
}

type DeleteBackupPolicySpec struct {
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	Tenants        []string `json:"tenants"`
	RecoveryWindow string   `json:"recoveryWindow"`
}

// TenantBackupStatus defines the observed state of TenantBackup
type TenantBackupStatus struct {
	TenantBackupSet []TenantBackupSetStatus `json:"backup set"`
}

type TenantBackupSetStatus struct {
	TenantName  string            `json:"tenantName"`
	ClusterName string            `json:"clusterName"`
	BackupJobs  []BackupJobStatus `json:"backupJobs,omitempty"`
	Interval    []IntervalSpec    `json:"interval,omitempty"`
	Schedule    []ScheduleSpec    `json:"schedule,omitempty"`
}

type BackupJobStatus struct {
	BackupSetId int    `json:"backupSetId"`
	BackupType  string `json:"backupType"`
	Status      string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TenantBackup is the Schema for the tenantbackups API
type TenantBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantBackupSpec   `json:"spec,omitempty"`
	Status TenantBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantBackupList contains a list of TenantBackup
type TenantBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantBackup{}, &TenantBackupList{})
}
