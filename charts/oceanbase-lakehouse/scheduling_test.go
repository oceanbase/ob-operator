package lakehouse_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
)

type scheduledLocation struct {
	name         string
	nodeSelector map[string]string
	tolerations  []corev1.Toleration
}

// Decode the complete objects into the real Kubernetes/Operator API types,
// rejecting unknown fields rather than silently accepting misplaced scheduling.
func decodeScheduledObject(t *testing.T, doc manifest, target interface{}) {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("invalid typed %s: %v", doc["kind"], err)
	}
}

func schedulingLocations(t *testing.T, docs []manifest) []scheduledLocation {
	t.Helper()
	var locations []scheduledLocation
	for _, doc := range docs {
		switch doc["kind"] {
		case "StatefulSet":
			var object appsv1.StatefulSet
			decodeScheduledObject(t, doc, &object)
			locations = append(locations, scheduledLocation{"minio", object.Spec.Template.Spec.NodeSelector, object.Spec.Template.Spec.Tolerations})
		case "Job":
			var object batchv1.Job
			decodeScheduledObject(t, doc, &object)
			locations = append(locations, scheduledLocation{object.Name, object.Spec.Template.Spec.NodeSelector, object.Spec.Template.Spec.Tolerations})
		case "OBLogServiceCluster":
			var object v1alpha1.OBLogServiceCluster
			decodeScheduledObject(t, doc, &object)
			for _, zone := range object.Spec.Topology {
				locations = append(locations, scheduledLocation{"ls/" + zone.Zone, zone.NodeSelector, zone.Tolerations})
			}
		case "OBCluster":
			var object v1alpha1.OBCluster
			decodeScheduledObject(t, doc, &object)
			for _, zone := range object.Spec.Topology {
				locations = append(locations, scheduledLocation{"ob/" + zone.Zone, zone.NodeSelector, zone.Tolerations})
			}
		}
	}
	return locations
}

func TestSchedulingDefaultDoesNotRestrictPlacement(t *testing.T) {
	for _, mode := range []string{"bundled", "external"} {
		t.Run(mode, func(t *testing.T) {
			storage := map[string]interface{}{"mode": mode}
			want := 7 // MinIO + init + readiness + three LS zones + one OB zone.
			if mode == "external" {
				storage["endpoint"] = "http://192.0.2.10:9000"
				storage["existingSecret"] = "external-storage"
				want--
			}
			docs, output, err := render(t, map[string]interface{}{"stage": "cluster", "objectStorage": storage})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			locations := schedulingLocations(t, docs)
			if len(locations) != want {
				t.Fatalf("unexpected scheduling locations: %d", len(locations))
			}
			for _, location := range locations {
				if len(location.nodeSelector) != 0 || len(location.tolerations) != 0 {
					t.Fatalf("default unexpectedly restricts %s: %+v", location.name, location)
				}
			}
		})
	}
}

func TestSchedulingPropagatesToEveryPodAndZone(t *testing.T) {
	selector := map[string]string{
		"kubernetes.io/hostname":     "seekdb-worker-01",
		"oceanbase.example.com/pool": "lakehouse",
		"example.com/string-label":   "true",
		"example.com/numeric-label":  "01",
	}
	seconds := int64(600)
	tolerations := []corev1.Toleration{
		{Key: "example.com/dedicated", Operator: corev1.TolerationOpEqual, Value: "lakehouse", Effect: corev1.TaintEffectNoSchedule},
		{Key: "node.kubernetes.io/not-ready", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoExecute, TolerationSeconds: &seconds},
		{Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectPreferNoSchedule},
	}
	for _, mode := range []string{"bundled", "external"} {
		t.Run(mode, func(t *testing.T) {
			storage := map[string]interface{}{"mode": mode}
			if mode == "external" {
				storage["endpoint"] = "http://192.0.2.10:9000"
				storage["existingSecret"] = "external-storage"
			}
			values := map[string]interface{}{
				"stage":         "cluster",
				"objectStorage": storage,
				"scheduling":    map[string]interface{}{"nodeSelector": selector, "tolerations": tolerations},
				"observer": map[string]interface{}{"topology": []interface{}{
					map[string]interface{}{"zone": "ob-zone1", "replica": 1},
					map[string]interface{}{"zone": "ob-zone2", "replica": 1},
				}},
			}
			docs, output, err := render(t, values)
			if err != nil {
				t.Fatalf("render scheduled %s: %v\n%s", mode, err, output)
			}
			locations := schedulingLocations(t, docs)
			want := 8 // MinIO + init + readiness + three LS zones + two OB zones.
			if mode == "external" {
				want-- // Existing MinIO is not managed or scheduled by this chart.
			}
			if len(locations) != want {
				t.Fatalf("got %d scheduling locations, want %d", len(locations), want)
			}
			for _, location := range locations {
				if !reflect.DeepEqual(location.nodeSelector, selector) || !reflect.DeepEqual(location.tolerations, tolerations) {
					t.Fatalf("scheduling changed or omitted at %s: %+v", location.name, location)
				}
			}
		})
	}
}

func TestSchedulingRejectsInvalidTolerations(t *testing.T) {
	for name, toleration := range map[string]map[string]interface{}{
		"unknown-operator":     {"key": "dedicated", "operator": "NotEqual", "value": "lakehouse", "effect": "NoSchedule"},
		"unknown-effect":       {"key": "dedicated", "operator": "Exists", "effect": "Never"},
		"exists-with-value":    {"key": "dedicated", "operator": "Exists", "value": "lakehouse", "effect": "NoSchedule"},
		"empty-key-equal":      {"key": "", "operator": "Equal", "value": "lakehouse", "effect": "NoSchedule"},
		"string-seconds":       {"key": "dedicated", "operator": "Exists", "effect": "NoExecute", "tolerationSeconds": "600"},
		"negative-seconds":     {"key": "dedicated", "operator": "Exists", "effect": "NoExecute", "tolerationSeconds": -1},
		"seconds-wrong-effect": {"key": "dedicated", "operator": "Exists", "effect": "NoSchedule", "tolerationSeconds": 600},
		"unknown-field":        {"key": "dedicated", "operator": "Exists", "effect": "NoExecute", "typoSeconds": 600},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := render(t, map[string]interface{}{"scheduling": map[string]interface{}{"tolerations": []interface{}{toleration}}}); err == nil {
				t.Fatal("invalid toleration was accepted")
			}
		})
	}
}

func TestSchedulingRequiresStringSelectorValues(t *testing.T) {
	if _, _, err := render(t, map[string]interface{}{"scheduling": map[string]interface{}{"nodeSelector": map[string]interface{}{"example.com/pool": true}}}); err == nil {
		t.Fatal("boolean selector value must not become an invalid Kubernetes label")
	}
}

func TestZoneNodeSelectorsDistributeWithoutLeakingToOtherZones(t *testing.T) {
	for _, mode := range []string{"bundled", "external"} {
		t.Run(mode, func(t *testing.T) {
			storage := map[string]interface{}{"mode": mode}
			if mode == "external" {
				storage["endpoint"] = "http://192.0.2.10:9000"
				storage["existingSecret"] = "external-storage"
			}
			tolerations := []corev1.Toleration{{Key: "example.com/pool", Operator: corev1.TolerationOpEqual, Value: "qa", Effect: corev1.TaintEffectNoSchedule}}
			topology := []interface{}{
				map[string]interface{}{"zone": "zone1", "replica": 1, "nodeSelector": map[string]string{"kubernetes.io/hostname": "worker-1", "example.com/zone": "01"}},
				map[string]interface{}{"zone": "zone2", "replica": 1, "nodeSelector": map[string]string{"kubernetes.io/hostname": "worker-2"}},
				map[string]interface{}{"zone": "zone3", "replica": 1},
			}
			docs, output, err := render(t, map[string]interface{}{
				"stage": "cluster", "objectStorage": storage,
				"scheduling": map[string]interface{}{"nodeSelector": map[string]string{"example.com/pool": "qa", "kubernetes.io/hostname": "worker-3"}, "tolerations": tolerations},
				"logService": map[string]interface{}{"topology": topology},
				"observer":   map[string]interface{}{"topology": topology},
			})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			locations := schedulingLocations(t, docs)
			wantCount := 9
			if mode == "external" {
				wantCount--
			}
			if len(locations) != wantCount {
				t.Fatalf("got %d locations, want %d", len(locations), wantCount)
			}
			for _, location := range locations {
				want := map[string]string{"example.com/pool": "qa", "kubernetes.io/hostname": "worker-3"}
				switch location.name {
				case "ls/zone1", "ob/zone1":
					want["kubernetes.io/hostname"] = "worker-1"
					want["example.com/zone"] = "01"
				case "ls/zone2", "ob/zone2":
					want["kubernetes.io/hostname"] = "worker-2"
				}
				if !reflect.DeepEqual(location.nodeSelector, want) || !reflect.DeepEqual(location.tolerations, tolerations) {
					t.Fatalf("wrong placement at %s: %+v, want selector %v", location.name, location, want)
				}
			}
		})
	}
}

func TestZoneNodeSelectorRequiresStringValues(t *testing.T) {
	for _, component := range []string{"logService", "observer"} {
		for _, value := range []interface{}{true, 123, []string{"node1"}} {
			if _, _, err := render(t, map[string]interface{}{component: map[string]interface{}{"topology": []interface{}{
				map[string]interface{}{"zone": "zone1", "replica": 3, "nodeSelector": map[string]interface{}{"kubernetes.io/hostname": value}},
			}}}); err == nil {
				t.Fatalf("%s accepted invalid selector value %v", component, value)
			}
		}
	}
}
