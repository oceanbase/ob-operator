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

package v1alpha1

import (
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBClusterOperationSpec defines the desired state of OBClusterOperation
type OBClusterOperationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	OBCluster string                        `json:"obcluster"`
	Type      apitypes.ClusterOperationType `json:"type"`
	Force     bool                          `json:"force,omitempty"`
	//+kubebuilder:default="7d"
	TTL              string                    `json:"ttl,omitempty"`
	AddZones         []apitypes.OBZoneTopology `json:"addZones,omitempty"`
	DeleteZones      []string                  `json:"deleteZones,omitempty"`
	AdjustReplicas   []AlterZoneReplicas       `json:"adjustReplicas,omitempty"`
	RestartOBServers *RestartOBServersConfig   `json:"restartOBServers,omitempty"`
	Upgrade          *UpgradeConfig            `json:"upgrade,omitempty"`
	ModifyOBServers  *ModifyOBServersConfig    `json:"modifyOBServers,omitempty"`
	SetParameters    []apitypes.Parameter      `json:"setParameters,omitempty"`
}

type ModifyOBServersConfig struct {
	Resource           *apitypes.ResourceSpec     `json:"resource,omitempty"`
	ExpandStorageSize  *ExpandStorageSizeConfig   `json:"expandStorageSize,omitempty"`
	ModifyStorageClass *ModifyStorageClassConfig  `json:"modifyStorageClass,omitempty"`
	AddingMonitor      *apitypes.MonitorTemplate  `json:"addingMonitor,omitempty"`
	AddingBackupVolume *apitypes.BackupVolumeSpec `json:"addingBackupVolume,omitempty"`
}

type RestartOBServersConfig struct {
	OBServers []string `json:"observers,omitempty"`
	OBZones   []string `json:"obzones,omitempty"`
	All       bool     `json:"all,omitempty"`
}

type ExpandStorageSizeConfig struct {
	DataStorage    *resource.Quantity `json:"dataStorage,omitempty"`
	LogStorage     *resource.Quantity `json:"logStorage,omitempty"`
	RedoLogStorage *resource.Quantity `json:"redoLogStorage,omitempty"`
}

type ModifyStorageClassConfig struct {
	DataStorage    string `json:"dataStorage,omitempty"`
	LogStorage     string `json:"logStorage,omitempty"`
	RedoLogStorage string `json:"redoLogStorage,omitempty"`
}

type UpgradeConfig struct {
	Image string `json:"image"`
}

type AlterZoneReplicas struct {
	Zones []string `json:"zones"`
	To    int      `json:"to,omitempty"`
	By    int      `json:"by,omitempty"`
}

// OBClusterOperationStatus defines the observed state of OBClusterOperation
type OBClusterOperationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           apitypes.ClusterOperationStatus `json:"status"`
	OperationContext *tasktypes.OperationContext     `json:"operationContext,omitempty"`
	ClusterSnapshot  *OBClusterSnapshot              `json:"clusterSnapshot,omitempty"`
}

type OBClusterSnapshot struct {
	Spec   *OBClusterSpec   `json:"spec,omitempty"`
	Status *OBClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Tasks",type="string",JSONPath=".status.operationContext.tasks",priority=1
//+kubebuilder:printcolumn:name="Task",type="string",JSONPath=".status.operationContext.task",priority=1

// OBClusterOperation is the Schema for the obclusteroperations API
type OBClusterOperation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBClusterOperationSpec   `json:"spec,omitempty"`
	Status OBClusterOperationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBClusterOperationList contains a list of OBClusterOperation
type OBClusterOperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBClusterOperation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBClusterOperation{}, &OBClusterOperationList{})
}

func (o *OBClusterOperation) ShouldBeCleaned() bool {
	ttl, err := strconv.Atoi(strings.TrimRight(o.Spec.TTL, "d"))
	if err != nil {
		return false
	}
	return o.CreationTimestamp.AddDate(0, 0, ttl).Before(metav1.Now().Time)
}
