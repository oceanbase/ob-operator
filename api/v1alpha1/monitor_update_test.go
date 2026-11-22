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

package v1alpha1

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	apitypes "github.com/oceanbase/ob-operator/api/types"
)

func TestMonitorRolloutAdmissionOnDynamicIP(t *testing.T) {
	old := &OBCluster{}
	old.Spec.OBServerTemplate = &apitypes.OBServerTemplate{Resource: &apitypes.ResourceSpec{}}
	old.Spec.MonitorTemplate = &apitypes.MonitorTemplate{Image: "agent:v1", Resource: &apitypes.ResourceSpec{Memory: resource.MustParse("1Gi")}}
	old.Spec.Topology = []apitypes.OBZoneTopology{{Zone: "a"}}
	for _, mutate := range []func(*OBCluster){
		func(c *OBCluster) { c.Spec.MonitorTemplate.Image = "agent:v2" },
		func(c *OBCluster) { c.Spec.MonitorTemplate.Resource.Cpu = resource.MustParse("1") },
		func(c *OBCluster) {
			c.Spec.Topology[0].MonitorTemplate = old.Spec.MonitorTemplate.DeepCopy()
			c.Spec.Topology[0].MonitorTemplate.Resource.Memory = resource.MustParse("2Gi")
		},
	} {
		newCluster := old.DeepCopy()
		mutate(newCluster)
		if _, err := newCluster.ValidateUpdate(old); err == nil || !strings.Contains(err.Error(), "monitor settings") {
			t.Fatalf("expected dynamic-IP monitor update rejection, got %v", err)
		}
	}
	equivalent := old.DeepCopy()
	equivalent.Spec.Topology[0].MonitorTemplate = old.Spec.MonitorTemplate.DeepCopy()
	equivalent.Spec.Topology[0].MonitorTemplate.Resource.Memory = resource.MustParse("1024Mi")
	equivalent.Spec.Topology = append(equivalent.Spec.Topology, apitypes.OBZoneTopology{Zone: "new", MonitorTemplate: &apitypes.MonitorTemplate{Image: "other"}})
	if changedMonitorSettings(old, equivalent) {
		t.Fatal("equivalent quantities, inherited settings, or a new zone rejected")
	}
	old.Spec.Topology[0].MonitorTemplate = old.Spec.MonitorTemplate.DeepCopy()
	next := old.DeepCopy()
	next.Spec.MonitorTemplate.Image = "unused-default"
	if changedMonitorSettings(old, next) {
		t.Fatal("default edit with every existing zone overridden does not roll Pods")
	}
	next.Spec.MonitorTemplate = nil
	if changedMonitorSettings(old, next) {
		t.Fatal("an unused default can be removed while all existing zones retain their monitors")
	}
}
