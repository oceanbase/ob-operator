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
	"reflect"
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestLogServiceUpdateContractReview(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*OBLogServiceCluster)
	}{
		{"cluster ID", func(c *OBLogServiceCluster) { c.Spec.ClusterId++ }},
		{"bucket URL", func(c *OBLogServiceCluster) { c.Spec.ObjectStoreURL.BucketURL += "-changed" }},
		{"credentials reference", func(c *OBLogServiceCluster) { c.Spec.ObjectStoreURL.SecretRef.Name += "-changed" }},
		{"image", func(c *OBLogServiceCluster) { c.Spec.LogService.Image += "-changed" }},
		{"memory", func(c *OBLogServiceCluster) { c.Spec.LogService.Resource.Memory = resource.MustParse("8Gi") }},
		{"log volume", func(c *OBLogServiceCluster) { c.Spec.LogService.Storage.LogStorage.Size = resource.MustParse("40Gi") }},
		{"add Zone", func(c *OBLogServiceCluster) {
			c.Spec.Topology = append(c.Spec.Topology, apitypes.LogServiceZoneTopology{Zone: "new-zone", Replica: 1})
		}},
		{"remove Zone", func(c *OBLogServiceCluster) { c.Spec.Topology = c.Spec.Topology[:1] }},
		{"Zone resource", func(c *OBLogServiceCluster) {
			c.Spec.Topology[0].Resource = &apitypes.ResourceSpec{Memory: resource.MustParse("8Gi")}
		}},
		{"Zone selector", func(c *OBLogServiceCluster) { c.Spec.Topology[0].NodeSelector = map[string]string{"pool": "other"} }},
		{"Zone affinity", func(c *OBLogServiceCluster) {
			c.Spec.Topology[0].Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}
		}},
		{"Zone tolerations", func(c *OBLogServiceCluster) {
			c.Spec.Topology[0].Tolerations = []corev1.Toleration{{Key: "pool", Operator: corev1.TolerationOpExists}}
		}},
		{"Zone RPC port", func(c *OBLogServiceCluster) { c.Spec.Topology[0].RpcPort = 51000 }},
		{"Zone HTTP port", func(c *OBLogServiceCluster) { c.Spec.Topology[0].HttpPort = 51001 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := newOBLogServiceCluster("update-contract", "secret-need-not-exist")
			old.Spec.Topology = append(old.Spec.Topology, apitypes.LogServiceZoneTopology{Zone: "zone2", Replica: 1})
			next := old.DeepCopy()
			tc.change(next)
			if _, err := next.ValidateUpdate(old); err == nil {
				t.Fatal("unsupported update accepted; controller only reconciles Zone replica changes")
			}
		})
	}
	for _, legacyReplica := range []int{0, -1} {
		t.Run("repair legacy replica "+resource.NewQuantity(int64(legacyReplica), resource.DecimalSI).String(), func(t *testing.T) {
			old := newOBLogServiceCluster("legacy", "secret-need-not-exist")
			old.Spec.Topology[0].Replica = legacyReplica
			next := old.DeepCopy()
			next.Spec.Topology[0].Replica = 1
			if _, err := next.ValidateUpdate(old); err != nil {
				t.Fatalf("repair rejected: %v", err)
			}
		})
	}
	t.Run("repair legacy duplicate Zone", func(t *testing.T) {
		old := newOBLogServiceCluster("legacy-duplicate", "secret-need-not-exist")
		old.Spec.Topology = append(old.Spec.Topology, old.Spec.Topology[0])
		next := old.DeepCopy()
		next.Spec.Topology = next.Spec.Topology[:1]
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("repair rejected: %v", err)
		}
	})
	t.Run("status update does not require Secret", func(t *testing.T) {
		old := newOBLogServiceCluster("status-update", "secret-need-not-exist")
		next := old.DeepCopy()
		next.Status.Status = "running"
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("status update rejected: %v", err)
		}
	})
	t.Run("scale and reorder preserve per-Zone overrides", func(t *testing.T) {
		old := newOBLogServiceCluster("per-zone-overrides", "secret-need-not-exist")
		old.Spec.Topology = []apitypes.LogServiceZoneTopology{
			{Zone: "zone1", Replica: 1, RpcPort: 50051, HttpPort: 50052, Resource: &apitypes.ResourceSpec{Memory: resource.MustParse("4Gi")}, NodeSelector: map[string]string{"pool": "one"}, Affinity: &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{}}, Tolerations: []corev1.Toleration{{Key: "pool", Operator: corev1.TolerationOpExists}}},
			{Zone: "zone2", Replica: 3, RpcPort: 51051, HttpPort: 51052, NodeSelector: map[string]string{"pool": "two"}},
		}
		next := old.DeepCopy()
		next.Spec.Topology[0].Replica = 2
		next.Spec.Topology[1].Replica = 1
		next.Spec.Topology[0], next.Spec.Topology[1] = next.Spec.Topology[1], next.Spec.Topology[0]
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("valid scale and reorder rejected: %v", err)
		}
	})
	t.Run("equivalent representations of Zone overrides", func(t *testing.T) {
		old := newOBLogServiceCluster("equivalent-overrides", "secret-need-not-exist")
		old.Spec.Topology[0].Resource = &apitypes.ResourceSpec{Memory: resource.MustParse("4Gi")}
		next := old.DeepCopy()
		next.Spec.Topology[0].Replica = 2
		next.Spec.Topology[0].Resource.Memory = resource.MustParse("4096Mi")
		next.Spec.Topology[0].NodeSelector = map[string]string{}
		next.Spec.Topology[0].Tolerations = []corev1.Toleration{}
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("equivalent override representation rejected: %v", err)
		}
	})
}

func TestLogServiceStartupUpdateContractReview(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*OBLogServiceCluster)
	}{
		{"ServiceAccount", func(c *OBLogServiceCluster) { c.Spec.ServiceAccount = "other-workload" }},
		{"startup parameter value", func(c *OBLogServiceCluster) { c.Spec.Parameters[0].Value = "INFO" }},
		{"add startup parameter", func(c *OBLogServiceCluster) {
			c.Spec.Parameters = append(c.Spec.Parameters, apitypes.Parameter{Name: "new_parameter", Value: "new_value"})
		}},
		{"remove startup parameter", func(c *OBLogServiceCluster) { c.Spec.Parameters = c.Spec.Parameters[1:] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := newOBLogServiceCluster("startup-contract", "secret-need-not-exist")
			old.Spec.ServiceAccount = "current-workload"
			old.Spec.Parameters = []apitypes.Parameter{{Name: "syslog_level", Value: "WARN"}, {Name: "log_disk_utilization_threshold", Value: "80"}}
			next := old.DeepCopy()
			tc.change(next)
			if _, err := next.ValidateUpdate(old); err == nil {
				t.Fatal("unsupported update accepted; startup parameters and ServiceAccount are not reconciled")
			}
		})
	}
	t.Run("parameter reordering while scaling", func(t *testing.T) {
		old := newOBLogServiceCluster("parameter-order", "secret-need-not-exist")
		old.Spec.ServiceAccount = "current-workload"
		old.Spec.Parameters = []apitypes.Parameter{{Name: "syslog_level", Value: "WARN"}, {Name: "log_disk_utilization_threshold", Value: "80"}}
		next := old.DeepCopy()
		next.Spec.Topology[0].Replica = 2
		next.Spec.Parameters[0], next.Spec.Parameters[1] = next.Spec.Parameters[1], next.Spec.Parameters[0]
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("valid scale with reordered parameters rejected: %v", err)
		}
	})
	t.Run("legacy spec may add derived memory_limit through defaulting", func(t *testing.T) {
		old := newOBLogServiceCluster("legacy-default", "secret-need-not-exist")
		old.Spec.Parameters = []apitypes.Parameter{{Name: "syslog_level", Value: "WARN"}}
		next := old.DeepCopy()
		next.Default()
		next.Spec.Topology[0].Replica = 2
		beforeOld, beforeNext := old.DeepCopy(), next.DeepCopy()
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("defaulted scale rejected: %v", err)
		}
		if !reflect.DeepEqual(old, beforeOld) || !reflect.DeepEqual(next, beforeNext) {
			t.Fatal("validation mutated admission objects while normalizing parameters")
		}
	})
	t.Run("removing explicit derived memory_limit preserves effective default", func(t *testing.T) {
		old := newOBLogServiceCluster("derived-default", "secret-need-not-exist")
		old.Default()
		next := old.DeepCopy()
		next.Spec.Parameters = nil
		next.Spec.Topology[0].Replica = 2
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatalf("equivalent effective default rejected: %v", err)
		}
	})
}
