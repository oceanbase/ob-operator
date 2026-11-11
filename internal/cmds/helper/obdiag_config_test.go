package helper

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOBDiagTargetsContainOnlyNodeIdentity(t *testing.T) {
	directory := t.TempDir()
	if err := writeOBDiagTargets(directory, "test", []NodeConfig{{PodName: "ob-pod", IP: "10.1.2.3", KubernetesConfigFile: "private.yaml"}}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "nodes.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var actual []map[string]string
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatal(err)
	}
	expected := []map[string]string{{"namespace": "test", "pod_name": "ob-pod"}}
	if !reflect.DeepEqual(actual, expected) || strings.Contains(string(data), "private") {
		t.Fatalf("unexpected target manifest: %v", actual)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("target manifest must have mode 0600")
	}
}
