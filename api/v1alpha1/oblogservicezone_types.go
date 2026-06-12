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

type OBLogServiceZoneSpec struct {
	ClusterName    string                          `json:"clusterName"`
	ClusterId      int64                           `json:"clusterId"`
	Image          string                          `json:"image"`
	Topology       apitypes.LogServiceZoneTopology `json:"topology"`
	ObjectStoreURL apitypes.ObjectStoreConfig      `json:"objectStoreUrl"`
	Storage        *apitypes.LogServiceStorageSpec `json:"storage"`
	ServiceAccount string                          `json:"serviceAccount,omitempty"`
}

type OBLogServiceZoneStatus struct {
	Status           string                                 `json:"status"`
	OperationContext *tasktypes.OperationContext            `json:"operationContext,omitempty"`
	NodeStatus       []apitypes.LogServiceNodeReplicaStatus `json:"nodeStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Zone",type="string",JSONPath=".spec.topology.zone"
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.clusterName"

type OBLogServiceZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBLogServiceZoneSpec   `json:"spec,omitempty"`
	Status OBLogServiceZoneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type OBLogServiceZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBLogServiceZone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBLogServiceZone{}, &OBLogServiceZoneList{})
}
