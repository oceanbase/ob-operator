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
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Exercise a packaged source chart with the three external release image inputs.
// Native distribution bundles embed these inputs in packaged values.yaml.
func TestPackagedSourceQuickstartWithReleaseImages(t *testing.T) {
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	dir := t.TempDir()
	if output, err := exec.Command(helm, "package", ".", "--destination", dir).CombinedOutput(); err != nil {
		t.Fatalf("package: %v\n%s", err, output)
	}
	archives, err := filepath.Glob(filepath.Join(dir, "*.tgz"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("expected one packaged chart: %v, %v", archives, err)
	}
	output, err := exec.Command(helm, "template", "qa", archives[0], "-n", "qa-ns", "-f", "testdata/release-images.yaml").CombinedOutput()
	if err != nil {
		t.Fatalf("default package render: %v\n%s", err, output)
	}
	docs := releaseDocuments(t, output)
	for _, item := range []struct{ kind, name string }{
		{"StatefulSet", "qa-minio"}, {"Job", "qa-storage-init-1"},
		{"OBLogServiceCluster", "qa-ls"}, {"OBCluster", "qa-ob"},
		{"Deployment", "oceanbase-dashboard-qa"}, {"Secret", "qa-dashboard-bootstrap"},
	} {
		find(t, docs, item.kind, item.name)
	}
	job := find(t, docs, "Job", "qa-readiness")
	if field(job, "metadata", "annotations", "helm.sh/hook") != "post-install,post-upgrade" {
		t.Fatal("quickstart must check readiness on installation and upgrade")
	}
	containers := field(job, "spec", "template", "spec", "containers").([]interface{})
	args := containers[0].(map[string]interface{})["args"].([]interface{})
	foundDashboard := false
	for _, arg := range args {
		if arg == "--dashboard=oceanbase-dashboard-qa" {
			foundDashboard = true
		}
	}
	if !foundDashboard {
		t.Fatal("default installation must wait for Dashboard readiness")
	}
	if field(find(t, docs, "Service", "oceanbase-dashboard-qa"), "spec", "type") != "NodePort" {
		t.Fatal("quickstart must publish Dashboard through NodePort")
	}
}

func TestDefaultSizingMatchesSingleNodeProfile(t *testing.T) {
	body, err := os.ReadFile("values-single-node.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var profile map[string]interface{}
	if err := yaml.Unmarshal(body, &profile); err != nil {
		t.Fatal(err)
	}
	defaults, output, err := render(t, nil)
	if err != nil {
		t.Fatal(err, output)
	}
	explicit, output, err := render(t, profile)
	if err != nil {
		t.Fatal(err, output)
	}
	var cpu, memory int64
	for _, tc := range []struct{ kind, name, section string }{
		{"OBCluster", "qa-ob", "observer"},
		{"OBLogServiceCluster", "qa-ls", "logService"},
	} {
		actual := find(t, defaults, tc.kind, tc.name)
		want := find(t, explicit, tc.kind, tc.name)
		for _, path := range [][]string{{"spec", tc.section, "resource"}, {"spec", tc.section, "storage"}, {"spec", "topology"}} {
			if !reflect.DeepEqual(field(actual, path...), field(want, path...)) {
				t.Fatalf("default %s %v does not match the single-node profile", tc.kind, path)
			}
		}
		perCPU := resource.MustParse(field(actual, "spec", tc.section, "resource", "cpu").(string))
		perMemory := resource.MustParse(field(actual, "spec", tc.section, "resource", "memory").(string))
		for _, zone := range field(actual, "spec", "topology").([]interface{}) {
			replicas := int64(zone.(manifest)["replica"].(int))
			cpu += replicas * perCPU.MilliValue()
			memory += replicas * perMemory.Value()
		}
	}
	// Include OBProxy in the test-profile component budget. This does not
	// include Kubernetes, monitoring or object storage overhead.
	var proxy appsv1.Deployment
	decodeScheduledObject(t, find(t, defaults, "Deployment", "qa-obproxy"), &proxy)
	for _, container := range proxy.Spec.Template.Spec.Containers {
		cpu += int64(*proxy.Spec.Replicas) * container.Resources.Requests.Cpu().MilliValue()
		memory += int64(*proxy.Spec.Replicas) * container.Resources.Requests.Memory().Value()
	}
	if cpu != 10000 || memory != resource.NewQuantity(24<<30, resource.BinarySI).Value() {
		t.Fatalf("test-profile database + proxy budget must be 10 CPU/24Gi: %dm CPU, %d bytes memory", cpu, memory)
	}
	for _, tc := range []struct {
		kind, name, section, data string
		total                     int64
	}{
		{"OBCluster", "qa-ob", "observer", "dataStorage", 50},
		{"OBLogServiceCluster", "qa-ls", "logService", "storeStorage", 20},
	} {
		doc := find(t, defaults, tc.kind, tc.name)
		data := resource.MustParse(field(doc, "spec", tc.section, "storage", tc.data, "size").(string))
		logs := resource.MustParse(field(doc, "spec", tc.section, "storage", "logStorage", "size").(string))
		if data.Value()+logs.Value() != tc.total<<30 {
			t.Fatalf("%s volumes do not fit its %dGi per-instance budget", tc.kind, tc.total)
		}
	}
}

func TestTestEnvironmentUsesFourSeparateWorkers(t *testing.T) {
	body, err := os.ReadFile("values-test-environment.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var profile map[string]interface{}
	if err := yaml.Unmarshal(body, &profile); err != nil {
		t.Fatal(err)
	}
	docs, output, err := render(t, profile)
	if err != nil {
		t.Fatal(err, output)
	}
	cluster := find(t, docs, "OBCluster", "qa-ob")
	computeZones := field(cluster, "spec", "topology").([]interface{})
	if len(computeZones) != 1 {
		t.Fatal("test profile must have one compute zone")
	}
	compute := field(computeZones[0].(manifest), "nodeSelector", "kubernetes.io/hostname")
	var proxy appsv1.Deployment
	decodeScheduledObject(t, find(t, docs, "Deployment", "qa-obproxy"), &proxy)
	if compute != "worker-compute" || proxy.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] != compute {
		t.Fatal("Observer and ODP must share the compute worker")
	}
	workers := map[interface{}]bool{compute: true}
	logZones := field(find(t, docs, "OBLogServiceCluster", "qa-ls"), "spec", "topology").([]interface{})
	if len(logZones) != 3 {
		t.Fatal("test profile must have three LogService zones")
	}
	for _, item := range logZones {
		zone := item.(manifest)
		worker := field(zone, "nodeSelector", "kubernetes.io/hostname")
		if worker == nil || workers[worker] || field(zone, "replica") != 1 {
			t.Fatalf("LogService must use a distinct worker: %v", zone)
		}
		workers[worker] = true
	}
}
