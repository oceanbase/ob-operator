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
	"os"
	"path/filepath"
	"strings"
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func TestLogServiceUpdateTopology(t *testing.T) {
	tests := []struct {
		name      string
		change    func(*OBLogServiceCluster)
		wantError string
	}{
		{"zero replica", func(c *OBLogServiceCluster) { c.Spec.Topology[0].Replica = 0 }, "replica must be at least 1"},
		{"negative replica", func(c *OBLogServiceCluster) { c.Spec.Topology[0].Replica = -1 }, "replica must be at least 1"},
		{"duplicate zone", func(c *OBLogServiceCluster) { c.Spec.Topology = append(c.Spec.Topology, c.Spec.Topology[0]) }, "is duplicated"},
		{"empty zone", func(c *OBLogServiceCluster) { c.Spec.Topology[0].Zone = "" }, "zone is required"},
		{"nil topology", func(c *OBLogServiceCluster) { c.Spec.Topology = nil }, "at least 1 zone"},
		{"empty topology", func(c *OBLogServiceCluster) { c.Spec.Topology = []apitypes.LogServiceZoneTopology{} }, "at least 1 zone"},
		{"scale out", func(c *OBLogServiceCluster) { c.Spec.Topology[0].Replica = 2 }, ""},
		{"unchanged spec", func(c *OBLogServiceCluster) {}, ""},
		{"metadata update", func(c *OBLogServiceCluster) { c.Annotations = map[string]string{"qa": "allowed"} }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := newOBLogServiceCluster("topology-test", "unavailable-secret-must-not-block-update")
			next := old.DeepCopy()
			tt.change(next)
			_, err := next.ValidateUpdate(old)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("valid update rejected: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("want error containing %q, got %v", tt.wantError, err)
			}
			if old.Spec.Topology[0].Replica != 1 || len(old.Spec.Topology) != 1 {
				t.Fatal("validation mutated old object")
			}
		})
	}
	t.Run("scale in and reorder", func(t *testing.T) {
		old := newOBLogServiceCluster("topology-test", "unused")
		old.Spec.Topology = []apitypes.LogServiceZoneTopology{{Zone: "zone1", Replica: 2}, {Zone: "zone2", Replica: 1}}
		next := old.DeepCopy()
		next.Spec.Topology = []apitypes.LogServiceZoneTopology{{Zone: "zone2", Replica: 1}, {Zone: "zone1", Replica: 1}}
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("repair legacy invalid replica", func(t *testing.T) {
		old := newOBLogServiceCluster("topology-test", "unused")
		old.Spec.Topology[0].Replica = 0
		next := old.DeepCopy()
		next.Spec.Topology[0].Replica = 1
		if _, err := next.ValidateUpdate(old); err != nil {
			t.Fatal(err)
		}
	})
}

func TestLogServiceTopologySchema(t *testing.T) {
	for _, resource := range []string{"oblogserviceclusters", "oblogservicezones"} {
		t.Run(resource, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "config", "crd", "bases", "oceanbase.oceanbase.com_"+resource+".yaml"))
			if err != nil {
				t.Fatal(err)
			}
			var crd apiextensionsv1.CustomResourceDefinition
			if err := yaml.Unmarshal(data, &crd); err != nil {
				t.Fatal(err)
			}
			for _, v := range crd.Spec.Versions {
				topology := v.Schema.OpenAPIV3Schema.Properties["spec"].Properties["topology"]
				if resource == "oblogserviceclusters" {
					if topology.MinItems == nil || *topology.MinItems != 1 {
						t.Fatal("topology must have minItems=1")
					}
					if topology.Items == nil || topology.Items.Schema == nil {
						t.Fatal("missing topology item schema")
					}
					topology = *topology.Items.Schema
				}
				replica := topology.Properties["replica"]
				if replica.Minimum == nil || *replica.Minimum != 1 {
					t.Fatal("replica must have minimum=1")
				}
				zone := topology.Properties["zone"]
				if zone.MinLength == nil || *zone.MinLength != 1 {
					t.Fatal("zone must have minLength=1")
				}
			}
		})
	}
}
