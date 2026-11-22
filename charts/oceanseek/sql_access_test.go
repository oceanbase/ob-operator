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

package oceanseek_test

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestSQLNodePortAndInternalAccess(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		port, node int32
	}{
		{"default", "NodePort", 2883, 0},
		{"custom", "NodePort", 2882, 30882},
		{"internal", "ClusterIP", 2883, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]interface{}{"obproxy": map[string]interface{}{"enabled": false}}
			if tc.name != "default" {
				service := map[string]interface{}{"type": tc.kind, "port": tc.port}
				if tc.node != 0 {
					service["nodePort"] = tc.node
				}
				values["sqlService"] = service
			}
			docs, output, err := render(t, values)
			if err != nil {
				t.Fatal(err, output)
			}
			var service corev1.Service
			decodeScheduledObject(t, find(t, docs, "Service", "qa-sql"), &service)
			if string(service.Spec.Type) != tc.kind || len(service.Spec.Ports) != 1 {
				t.Fatal(service.Spec)
			}
			p := service.Spec.Ports[0]
			if p.Port != tc.port || p.NodePort != tc.node || p.TargetPort.IntVal != 2881 {
				t.Fatal(p)
			}
			if service.Spec.Selector["ref-obcluster"] != "qa-ob" {
				t.Fatal(service.Spec.Selector)
			}
		})
	}
	for _, stage := range []string{"storage", "logservice"} {
		docs, output, err := render(t, map[string]interface{}{"stage": stage})
		if err != nil {
			t.Fatal(err, output)
		}
		for _, doc := range docs {
			if doc["kind"] == "Service" && field(doc, "metadata", "name") == "qa-sql" {
				t.Fatal("SQL Service exposed before cluster stage")
			}
		}
	}
}
