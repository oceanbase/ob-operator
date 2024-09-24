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
	"encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// K8sClusterSpec defines the desired state of K8sCluster
type K8sClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	KubeConfig  string `json:"kubeConfig"`
}

// K8sClusterStatus defines the observed state of K8sCluster
type K8sClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,shortName=kc
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ClusterName",type="string",JSONPath=".spec.name"
//+kubebuilder:printcolumn:name="Description",type="string",JSONPath=".spec.description",priority=1

// K8sCluster is the Schema for the k8sclusters API
type K8sCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K8sClusterSpec   `json:"spec,omitempty"`
	Status K8sClusterStatus `json:"status,omitempty"`
}

// EncodeKubeConfig returns base64 encoded kubeconfig
func (s *K8sCluster) EncodeKubeConfig() string {
	base64Str := base64.StdEncoding.EncodeToString([]byte(s.Spec.KubeConfig))
	return base64Str
}

// DecodeKubeConfig returns kubeconfig from base64 encoded kubeconfig
func (s *K8sCluster) DecodeKubeConfig() ([]byte, error) {
	kubeConfig, err := base64.StdEncoding.DecodeString(s.Spec.KubeConfig)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

//+kubebuilder:object:root=true

// K8sClusterList contains a list of K8sCluster
type K8sClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K8sCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K8sCluster{}, &K8sClusterList{})
}
