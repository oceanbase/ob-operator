package lakehouse_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Exercise the shipped chart without any -f or --set arguments, as in the
// public quickstart. Defaults must include the full stack and wait for it.
func TestPackagedQuickstartNeedsNoValuesOverrides(t *testing.T) {
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
	output, err := exec.Command(helm, "template", "qa", archives[0], "-n", "qa-ns").CombinedOutput()
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
	// Leave room on the documented 8 CPU / 32Gi experience worker for MinIO,
	// Dashboard, monitoring, initialization and Kubernetes itself.
	if cpu > 5000 || memory > resource.NewQuantity(22<<30, resource.BinarySI).Value() {
		t.Fatalf("database defaults exceed the single-node budget: %dm CPU, %d bytes memory", cpu, memory)
	}
}
