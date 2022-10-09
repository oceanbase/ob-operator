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

// OBClusterSpec defines the desired state of OBCluster
type BackupSpec struct {
	// +kubebuilder:validation:Minimum=1
	SourceCluster []SourceClusterSpec `json:"source"`
	DestPath      string              `json:"dest_path"`
	Schedule      []ScheduleSpec      `json:"schedule"`
	Parameters    []Parameter         `json:"parameters"`
}

// SourceCluster defines the source cluster
type SourceClusterSpec struct {
	ClusterID   int    `json:"clusterID"`
	ClusterName string `json:"clusterName"`
}

// ScheduleSpec defines the schedule strategy
type ScheduleSpec struct {
	BackupType string `json:"backupType"`
	Schedule   string `json:"schedule"`
}

// BackupStatus defines the observed state of backup
type BackupStatus struct {
	BackupSet []BackupSetStatus `json:"backup set"`
	Interval  []IntervalSpec    `json:"interval"`
	Schedule  []ScheduleSpec    `json:"schedule"`
}

type BackupSetStatus struct {
	TenantID    int    `json:"tenantID"`
	BSKey       int    `json:"bsKey"`
	ClusterName string `json:"clusterName"`
	BackupType  string `json:"backupType"`
	Status      string `json:"status"`
}

type IntervalSpec struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Backup is the Schema for the backups API
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   BackupSpec   `json:"spec"`
	Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupList contains a list of backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}
