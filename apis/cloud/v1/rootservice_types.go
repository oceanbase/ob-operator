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

// RootServiceSpec defines the desired state of RootService
type RootServiceSpec struct {
	Topology []Cluster `json:"topology"`
}

// RootServiceStatus defines the observed state of RootService
type RootServiceStatus struct {
	Topology []ClusterRootServiceStatus `json:"topology"`
}

type ClusterRootServiceStatus struct {
	Cluster string                  `json:"cluster"`
	Zone    []ZoneRootServiceStatus `json:"zoneRootService"`
}

type ZoneRootServiceStatus struct {
	Name     string `json:"name"`
	ServerIP string `json:"serverIP"`
	Role     int64  `json:"role"`
	Status   string `json:"status"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RootService is the Schema for the rootservices API
type RootService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RootServiceSpec   `json:"spec,omitempty"`
	Status RootServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RootServiceList contains a list of RootService
type RootServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RootService `json:"items"`
}
