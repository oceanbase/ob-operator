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
	apitypes "github.com/oceanbase/ob-operator/api/types"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBClusterSpec defines the desired state of OBCluster
type OBClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName      string                     `json:"clusterName"`
	ClusterId        int64                      `json:"clusterId,omitempty"`
	OBServerTemplate *apitypes.OBServerTemplate `json:"observer"`
	MonitorTemplate  *apitypes.MonitorTemplate  `json:"monitor,omitempty"`
	BackupVolume     *apitypes.BackupVolumeSpec `json:"backupVolume,omitempty"`
	Parameters       []apitypes.Parameter       `json:"parameters,omitempty"`
	Topology         []apitypes.OBZoneTopology  `json:"topology"`
	UserSecrets      *apitypes.OBUserSecrets    `json:"userSecrets"`
	//+kubebuilder:default=default
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// OBClusterStatus defines the observed state of OBCluster
type OBClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image            string                         `json:"image"`
	OperationContext *tasktypes.OperationContext    `json:"operationContext,omitempty"`
	Status           string                         `json:"status"`
	OBZoneStatus     []apitypes.OBZoneReplicaStatus `json:"obzones"`
	Parameters       []apitypes.Parameter           `json:"parameters"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// OBCluster is the Schema for the obclusters API
type OBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBClusterSpec   `json:"spec,omitempty"`
	Status OBClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBClusterList contains a list of OBCluster
type OBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBCluster{}, &OBClusterList{})
}
