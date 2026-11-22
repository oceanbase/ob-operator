/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
)

func TestMonitorRolloutDistinguishesEditsFromAdmissionDefaults(t *testing.T) {
	m := &OBServerManager{OBServer: &v1alpha1.OBServer{}}
	m.OBServer.Spec.MonitorTemplate = &apitypes.MonitorTemplate{Image: "agent", Resource: &apitypes.ResourceSpec{Memory: resource.MustParse("1Gi")}}
	pod := &corev1.Pod{}
	pod.Spec.Containers = []corev1.Container{{Name: obagentconst.ContainerName, Image: "agent", Resources: corev1.ResourceRequirements{
		Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("1Gi")},
	}}}
	if m.checkIfMonitorMutated(pod) {
		t.Fatal("admission-defaulted CPU restarted a legacy Pod")
	}
	const revision = "oceanbase.oceanbase.com/monitor-revision"
	m.OBServer.Annotations = map[string]string{revision: "1234"}
	if !m.checkIfMonitorMutated(pod) {
		t.Fatal("updated template did not restart the old Pod")
	}
	pod.Annotations = map[string]string{revision: "1234"}
	if m.checkIfMonitorMutated(pod) {
		t.Fatal("replacement Pod restarted despite applying the current revision")
	}
	pod.Spec.Containers = nil
	if !m.checkIfMonitorMutated(pod) {
		t.Fatal("missing monitor container was not detected")
	}
}
