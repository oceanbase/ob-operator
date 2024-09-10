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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// K8sClusterCredentialSpec defines the desired state of K8sClusterCredential
type K8sClusterCredentialSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	K8sCluster string `json:"k8sCluster"`
	KubeConfig string `json:"kubeConfig"`
}

// K8sClusterCredentialStatus defines the observed state of K8sClusterCredential
type K8sClusterCredentialStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=kcc

// K8sClusterCredential is the Schema for the k8sclustercredentials API
type K8sClusterCredential struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K8sClusterCredentialSpec   `json:"spec,omitempty"`
	Status K8sClusterCredentialStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// K8sClusterCredentialList contains a list of K8sClusterCredential
type K8sClusterCredentialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K8sClusterCredential `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K8sClusterCredential{}, &K8sClusterCredentialList{})
}
