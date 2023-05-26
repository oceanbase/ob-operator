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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBServerSpec defines the desired state of OBServer
type OBServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName      string             `json:"clusterName"`
	ClusterId        int64              `json:"clusterId,omitempty"`
	Zone             string             `json:"zone"`
	NodeSelector     *map[string]string `json:"nodeSelector,omitempty"`
	OBServerTemplate *OBServerTemplate  `json:"observerTemplate"`
	MonitorTemplate  *MonitorTemplate   `json:"monitorTemplate,omitempty"`
	BackupVolume     *BackupVolumeSpec  `json:"backupVolume,omitempty"`
}

// OBServerStatus defines the observed state of OBServer
type OBServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationContext *OperationContext `json:"operationContext,omitempty"`
	Image            string            `json:"image"`
	Status           string            `json:"status"`
	PodPhase         corev1.PodPhase   `json:"podPhase"`
	Ready            bool              `json:"ready"`
	PodIp            string            `json:"podIp"`
	NodeIp           string            `json:"nodeIp"`
	// TODO uncomment this
	// Storage          []PVCStatus       `json:"storage"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OBServer is the Schema for the observers API
type OBServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBServerSpec   `json:"spec,omitempty"`
	Status OBServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBServerList contains a list of OBServer
type OBServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBServer{}, &OBServerList{})
}
