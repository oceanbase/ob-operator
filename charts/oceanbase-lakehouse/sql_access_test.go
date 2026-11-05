package lakehouse_test

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestSQLNodePortAndInternalAccess(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		port, node int32
	}{
		{"default", "NodePort", 2881, 30881},
		{"custom", "NodePort", 2882, 30882},
		{"internal", "ClusterIP", 2881, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var values map[string]interface{}
			if tc.name != "default" {
				service := map[string]interface{}{"type": tc.kind, "port": tc.port}
				if tc.node != 0 {
					service["nodePort"] = tc.node
				}
				values = map[string]interface{}{"sqlService": service}
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
