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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OBClusterReference identifies an OBCluster resource that this OBProxy connects to.
type OBClusterReference struct {
	// Namespace of the referenced OBCluster. Defaults to the namespace of the OBProxy resource.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Name of the OBCluster resource.
	Name string `json:"name"`
}

// OBProxySpec defines the desired state of OBProxy
type OBProxySpec struct {
	// Reference to the target OBCluster.
	OBCluster OBClusterReference `json:"obCluster"`

	// ProxyClusterName is the OBProxy cluster (application) name used inside OceanBase proxy.
	ProxyClusterName string `json:"proxyClusterName"`

	// ProxySysSecret is the name of a Secret in the same namespace as this OBProxy that holds
	// the root@proxysys password. The Secret must contain a key named "password".
	ProxySysSecret string `json:"proxySysSecret"`

	// Image is the OBProxy container image (e.g. oceanbase/obproxy-ce:tag).
	Image string `json:"image"`

	// ServiceType is the Kubernetes Service type exposed for SQL access.
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer;ExternalName
	ServiceType string `json:"serviceType,omitempty"`

	// Replicas is the number of OBProxy pods.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`

	// Resource limits and requests for each OBProxy pod.
	Resource *apitypes.ResourceSpec `json:"resource"`

	// Parameters are optional OBProxy configuration key-value pairs (mapped to proxy startup config).
	// +optional
	Parameters []apitypes.Parameter `json:"parameters,omitempty"`

	// ServiceAccount is the name of the ServiceAccount to use for OBProxy pods.
	// +kubebuilder:default=default
	// +optional
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// NodeSelector constrains OBProxy pods to nodes matching these labels.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Affinity defines scheduling affinity rules for OBProxy pods.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations allow OBProxy pods to be scheduled on tainted nodes.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// OBProxyStatus defines the observed state of OBProxy
type OBProxyStatus struct {
	// High-level phase, e.g. Running, Pending, Failed.
	Status string `json:"status,omitempty"`

	// Image is the container image currently running in the Deployment (observed).
	Image string `json:"image,omitempty"`

	// Replicas is the replica count currently set in the Deployment (observed).
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready OBProxy pods (observed).
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// RSList is the RS_LIST env value currently set in the Deployment (observed).
	// +optional
	RSList string `json:"rsList,omitempty"`

	// RSListSource indicates how RSList was last resolved: "k8s" or "sql".
	// +optional
	RSListSource string `json:"rsListSource,omitempty"`

	// ServiceIP is the ClusterIP of the managed Service (observed).
	// +optional
	ServiceIP string `json:"serviceIP,omitempty"`

	// Conditions holds granular state for diagnostic purposes.
	// Known types: OBClusterAvailable, OBClusterReady, RSListAvailable.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// OperationContext carries long-running task progress when using the operator task flow.
	// +optional
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=obp
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas"
//+kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas"
//+kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".spec.obCluster.name"
//+kubebuilder:printcolumn:name="RSListSrc",type="string",JSONPath=".status.rsListSource"
//+kubebuilder:printcolumn:name="ServiceIP",type="string",JSONPath=".status.serviceIP",priority=1
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// OBProxy is the Schema for the obproxies API
type OBProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBProxySpec   `json:"spec,omitempty"`
	Status OBProxyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OBProxyList contains a list of OBProxy
type OBProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBProxy{}, &OBProxyList{})
}
