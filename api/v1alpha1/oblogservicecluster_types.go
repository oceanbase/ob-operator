/*
Copyright (c) 2025 OceanBase
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
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type OBLogServiceClusterSpec struct {
	ClusterId      int64                             `json:"clusterId"`
	Image          string                            `json:"image"`
	Topology       []apitypes.LogServiceZoneTopology `json:"topology"`
	ObjectStoreURL apitypes.ObjectStoreConfig        `json:"objectStoreUrl"`
	Storage        *apitypes.LogServiceStorageSpec   `json:"storage"`
	//+kubebuilder:default=default
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

type OBLogServiceClusterStatus struct {
	Status           string                                 `json:"status"`
	OperationContext *tasktypes.OperationContext            `json:"operationContext,omitempty"`
	ZoneStatus       []apitypes.LogServiceZoneReplicaStatus `json:"zoneStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ClusterId",type="integer",JSONPath=".spec.clusterId"
//+kubebuilder:printcolumn:name="Tasks",type="string",JSONPath=".status.operationContext.tasks",priority=1
//+kubebuilder:printcolumn:name="Task",type="string",JSONPath=".status.operationContext.task",priority=1
//+kubebuilder:printcolumn:name="TaskStatus",type="string",JSONPath=".status.operationContext.taskStatus",priority=1

type OBLogServiceCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBLogServiceClusterSpec   `json:"spec,omitempty"`
	Status OBLogServiceClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type OBLogServiceClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBLogServiceCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBLogServiceCluster{}, &OBLogServiceClusterList{})
}
