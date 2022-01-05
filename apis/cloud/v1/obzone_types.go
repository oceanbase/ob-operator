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

// OBZoneSpec defines the desired state of OBZone
type OBZoneSpec struct {
	Topology []Cluster `json:"topology"`
}

// OBZoneStatus defines the observed state of OBZone
type OBZoneStatus struct {
	Topology []ClusterOBZoneStatus `json:"topology"`
}

type ClusterOBZoneStatus struct {
	Cluster string       `json:"cluster"`
	Zone    []OBZoneInfo `json:"zone"`
}

type OBZoneInfo struct {
	Name  string   `json:"name"`
	Nodes []OBNode `json:"nodes"`
}

type OBNode struct {
	ServerIP string `json:"serverIP"`
	Status   string `json:"status"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OBZone is the Schema for the obzones API
type OBZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBZoneSpec   `json:"spec,omitempty"`
	Status OBZoneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OBZoneList contains a list of OBZone
type OBZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBZone `json:"items"`
}
