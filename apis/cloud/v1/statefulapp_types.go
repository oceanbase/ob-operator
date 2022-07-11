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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulAppSpec defines the desired state of StatefulApp
type StatefulAppSpec struct {
	Cluster          string            `json:"cluster"`
	Subsets          []Subset          `json:"subsets"`
	PodTemplate      corev1.PodSpec    `json:"podTemplate"`
	StorageTemplates []StorageTemplate `json:"storageTemplates"`
}

type Subset struct {
	Name         string            `json:"name"`
	Region       string            `json:"region,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector"`
	// +kubebuilder:validation:Minimum=1
	Replicas int `json:"replicas"`
}

type StorageTemplate struct {
	Name string                           `json:"name"`
	PVC  corev1.PersistentVolumeClaimSpec `json:"pvc"`
}

// StatefulAppStatus defines the observed state of StatefulApp
type StatefulAppStatus struct {
	Cluster       string         `json:"cluster"`
	ClusterStatus string         `json:"clusterStatus"`
	Subsets       []SubsetStatus `json:"subsets"`
}

type SubsetStatus struct {
	Name              string      `json:"name"`
	Region            string      `json:"region,omitempty"`
	ExpectedReplicas  int         `json:"expectedReplicas"`
	AvailableReplicas int         `json:"availableReplicas"`
	Pods              []PodStatus `json:"pods"`
}

type PodStatus struct {
	Name     string          `json:"name"`
	Index    int             `json:"index"`
	PodPhase corev1.PodPhase `json:"podPhase"`
	PodIP    string          `json:"podIP"`
	NodeIP   string          `json:"nodeIP"`
	PVCs     []PVCStatus     `json:"pvcs,omitempty"`
}

type PVCStatus struct {
	Name  string                       `json:"name"`
	Phase corev1.PersistentVolumePhase `json:"phase"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// StatefulApp is the Schema for the statefulapps API
type StatefulApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   StatefulAppSpec   `json:"spec"`
	Status StatefulAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StatefulAppList contains a list of StatefulApp
type StatefulAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []StatefulApp `json:"items"`
}
