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

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OBServerSpec defines the desired state of OBServer
type OBServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterName      string                     `json:"clusterName"`
	ClusterId        int64                      `json:"clusterId,omitempty"`
	Zone             string                     `json:"zone"`
	NodeSelector     map[string]string          `json:"nodeSelector,omitempty"`
	Affinity         *corev1.Affinity           `json:"affinity,omitempty"`
	Tolerations      []corev1.Toleration        `json:"tolerations,omitempty"`
	OBServerTemplate *apitypes.OBServerTemplate `json:"observerTemplate"`
	MonitorTemplate  *apitypes.MonitorTemplate  `json:"monitorTemplate,omitempty"`
	BackupVolume     *apitypes.BackupVolumeSpec `json:"backupVolume,omitempty"`
	//+kubebuilder:default=default
	ServiceAccount string `json:"serviceAccount,omitempty"`
	K8sCluster     string `json:"k8sCluster,omitempty"`
}

// OBServerStatus defines the observed state of OBServer
type OBServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
	Image            string                      `json:"image"`
	Status           string                      `json:"status"`
	PodPhase         corev1.PodPhase             `json:"podPhase"`
	Ready            bool                        `json:"ready"`
	PodIp            string                      `json:"podIp"`
	ServiceIp        string                      `json:"serviceIp,omitempty"`
	NodeIp           string                      `json:"nodeIp"`
	OBStatus         string                      `json:"obStatus,omitempty"`
	StartServiceTime int64                       `json:"startServiceTime,omitempty"`
	CNI              string                      `json:"cni,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="PodIP",type="string",JSONPath=".status.podIp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ClusterName",type="string",JSONPath=".spec.clusterName"
//+kubebuilder:printcolumn:name="ZoneName",type="string",JSONPath=".spec.zone"
//+kubebuilder:printcolumn:name="OBStatus",type="string",JSONPath=".status.obStatus",priority=1
//+kubebuilder:printcolumn:name="Tasks",type="string",JSONPath=".status.operationContext.tasks",priority=1
//+kubebuilder:printcolumn:name="Task",type="string",JSONPath=".status.operationContext.task",priority=1

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

func (ss OBServerStatus) GetConnectAddr() string {
	if ss.ServiceIp != "" {
		return ss.ServiceIp
	}
	return ss.PodIp
}

func (s *OBServer) SupportStaticIP() bool {
	switch s.Status.CNI {
	case oceanbaseconst.CNICalico:
		return true
	default:
		annos := s.GetAnnotations()
		if annos == nil {
			return false
		}
		mode, modeAnnoExist := annos[oceanbaseconst.AnnotationsMode]
		return modeAnnoExist && (mode == oceanbaseconst.ModeStandalone || mode == oceanbaseconst.ModeService)
	}
}

func (s *OBServer) InMasterK8s() bool {
	return s.Spec.K8sCluster == ""
}
