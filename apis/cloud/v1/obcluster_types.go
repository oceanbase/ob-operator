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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OBClusterSpec defines the desired state of OBCluster
type OBClusterSpec struct {
	ImageRepo string `json:"imageRepo"`
	Tag       string `json:"tag"`
	// +kubebuilder:validation:Minimum=1
	ClusterID    int           `json:"clusterID"`
	ImageObagent string        `json:"imageObagent"`
	Topology     []Cluster     `json:"topology"`
	Resources    ResourcesSpec `json:"resources"`
}

type Cluster struct {
	Cluster    string      `json:"cluster"`
	Zone       []Subset    `json:"zone"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResourcesSpec struct {
	CPU     resource.Quantity `json:"cpu"`
	Memory  resource.Quantity `json:"memory"`
	Storage []StorageSpec     `json:"storage"`
	Volume  VolumeSpec        `json:"volume"`
}

type VolumeSpec struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type StorageSpec struct {
	Name             string            `json:"name"`
	StorageClassName string            `json:"storageClassName"`
	Size             resource.Quantity `json:"size"`
}

// OBClusterStatus defines the observed state of OBCluster
type OBClusterStatus struct {
	Status   string          `json:"status"`
	Topology []ClusterStatus `json:"topology"`
}

type ClusterStatus struct {
	Cluster            string       `json:"cluster"`
	ClusterStatus      string       `json:"clusterStatus"`
	LastTransitionTime metav1.Time  `json:"lastTransitionTime"`
	Zone               []ZoneStatus `json:"zone"`
}

type ZoneStatus struct {
	Name              string `json:"name"`
	Region            string `json:"region"`
	ZoneStatus        string `json:"zoneStatus"`
	ExpectedReplicas  int    `json:"expectedReplicas"`
	AvailableReplicas int    `json:"availableReplicas"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OBCluster is the Schema for the obclusters API
type OBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   OBClusterSpec   `json:"spec"`
	Status OBClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OBClusterList contains a list of OBCluster
type OBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBCluster `json:"items"`
}
