/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type OBLogServiceNodeSpec struct {
	ClusterName    string                          `json:"clusterName"`
	ClusterId      int64                           `json:"clusterId"`
	Zone           string                          `json:"zone"`
	Image          string                          `json:"image"`
	Resource       *apitypes.ResourceSpec          `json:"resource,omitempty"`
	RpcPort        int32                           `json:"rpcPort,omitempty"`
	HttpPort       int32                           `json:"httpPort,omitempty"`
	NodeSelector   map[string]string               `json:"nodeSelector,omitempty"`
	Affinity       *corev1.Affinity                `json:"affinity,omitempty"`
	Tolerations    []corev1.Toleration             `json:"tolerations,omitempty"`
	ObjectStoreURL apitypes.ObjectStoreConfig      `json:"objectStoreUrl"`
	Storage        *apitypes.LogServiceStorageSpec `json:"storage"`
	Parameters     []apitypes.Parameter            `json:"parameters,omitempty"`
	ServiceAccount string                          `json:"serviceAccount,omitempty"`
}

type OBLogServiceNodeStatus struct {
	Status           string                      `json:"status"`
	OperationContext *tasktypes.OperationContext `json:"operationContext,omitempty"`
	PodName          string                      `json:"podName,omitempty"`
	PodIP            string                      `json:"podIP,omitempty"`
	ServiceIP        string                      `json:"serviceIP,omitempty"`
	CNI              string                      `json:"cni,omitempty"`
	PodPhase         corev1.PodPhase             `json:"podPhase,omitempty"`
	Ready            bool                        `json:"ready"`
}

func (s *OBLogServiceNodeStatus) GetConnectAddr() string {
	if s.ServiceIP != "" {
		return s.ServiceIP
	}
	return s.PodIP
}

func (n *OBLogServiceNode) SupportStaticIP() bool {
	switch n.Status.CNI {
	case oceanbaseconst.CNICalico, oceanbaseconst.CNIKubeOvn:
		return true
	default:
		// In service-IP advertise mode (oblogservice >=1.3.0) the per-pod Service
		// provides a stable identity, so the pod can be recreated for the same
		// node and rejoin at the same ServiceIP — mirroring observer's service
		// mode (SupportStaticIP returns true when mode=service).
		return n.Status.ServiceIP != ""
	}
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Zone",type="string",JSONPath=".spec.zone"
//+kubebuilder:printcolumn:name="PodIP",type="string",JSONPath=".status.podIP"
//+kubebuilder:printcolumn:name="ServiceIP",type="string",JSONPath=".status.serviceIP"

type OBLogServiceNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OBLogServiceNodeSpec   `json:"spec,omitempty"`
	Status OBLogServiceNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type OBLogServiceNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OBLogServiceNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OBLogServiceNode{}, &OBLogServiceNodeList{})
}
