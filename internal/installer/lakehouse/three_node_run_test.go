package lakehouse

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Exercise the real Run path using only local command fixtures. In particular,
// the SSH/SCP sentinels prevent any test from reaching a real host or runtime.
func newThreeNodeRunFixture(t *testing.T) *resourceRunFixture {
	t.Helper()
	f := newResourceRunFixture(t)
	nodes := f.responses["kubectl get nodes -o json"].(corev1.NodeList)
	for i, name := range []string{"worker-1", "worker-2", "worker-3"} {
		var node corev1.Node
		if i == 0 {
			node = *nodes.Items[0].DeepCopy()
		} else {
			node = *nodes.Items[0].DeepCopy()
			node.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "192.0.2." + string(rune('1'+i))}}
		}
		node.Name = name
		// Node object names must never be used as hostname labels implicitly.
		node.Labels["kubernetes.io/hostname"] = "host-" + string(rune('1'+i))
		node.Status.NodeInfo.Architecture = "amd64"
		node.Status.NodeInfo.OperatingSystem = "linux"
		node.Status.NodeInfo.SystemUUID = "fixture-uuid-" + name
		if i == 0 {
			nodes.Items[0] = node
		} else {
			nodes.Items = append(nodes.Items, node)
		}
		f.responses["kubectl get node "+name+" -o json"] = node
	}
	f.responses["kubectl get nodes -o json"] = nodes
	f.responses["kubectl get oblogserviceclusters -n demo-ns -o json"] = map[string]interface{}{"items": []interface{}{}}
	f.syncCommands(t)
	for _, name := range []string{"ssh", "scp"} {
		body := "#!/bin/sh\nprintf '%s\\n' " + resourceTestQuote(name+" ") + "\"$*\" >> " + resourceTestQuote(f.trace) + "\nprintf 'Unexpected remote fixture command\\n' >&2\nexit 97\n"
		if err := os.WriteFile(filepath.Join(f.bin, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

// Permit only read-only SSH identity/runtime/image commands. The only executable
// remote command is Docker, resolved through resourceRunFixture's strict stub;
// neither arbitrary SSH commands nor transfer/import commands can reach a host.
func threeNodeReadOnlySSH(t *testing.T, f *resourceRunFixture, cold bool) {
	t.Helper()
	body := `#!/bin/sh
set -eu
while [ "$#" -gt 0 ] && [ "$1" != -- ]; do shift; done
[ "$#" -eq 3 ] || exit 97
shift
host=$1
remote=$2
printf 'ssh %s %s\n' "$host" "$remote" >> TRACE
case "$host" in
  192.0.2.2|node-two.example) node=worker-2 ;;
  192.0.2.3) node=worker-3 ;;
  *) printf 'Unexpected SSH fixture host\n' >&2; exit 97 ;;
esac
case "$remote" in
  "'sh' '-c' "*)
    case "$remote" in
      *'/sys/class/dmi/id/product_uuid'*) printf 'UUID fixture-uuid-%s\n' "$node" ;;
      *) printf 'Mutating SSH fixture command rejected\n' >&2; exit 97 ;;
    esac ;;
  "'docker' "*)
    case "$remote" in
      *"'image' 'inspect'"*)
        if [ COLD = yes ]; then
          printf 'sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff\n'
          exit 0
        fi ;;
    esac
    /bin/sh -c "$remote" ;;
  *) printf 'Unexpected SSH fixture command\n' >&2; exit 97 ;;
esac
`
	body = strings.ReplaceAll(body, "TRACE", resourceTestQuote(f.trace))
	mode := "no"
	if cold {
		mode = "yes"
	}
	body = strings.ReplaceAll(body, "COLD", mode)
	if err := os.WriteFile(filepath.Join(f.bin, "ssh"), []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
}

func assertThreeNodeRunPlacement(t *testing.T, values map[string]interface{}) {
	t.Helper()
	for _, component := range []string{"observer", "logService"} {
		zones, _ := pfGet(values, component, "topology").([]interface{})
		if len(zones) != 3 {
			t.Fatalf("%s topology contains %d entries, want three", component, len(zones))
		}
		for i, item := range zones {
			zone := pfMap(item)
			if pfString(zone["zone"]) != fmt.Sprintf("zone%d", i+1) || pfString(zone["replica"]) != "1" {
				t.Error("three-node entry must create three separate one-replica zones", component, zone)
			}
			if host := pfString(pfGet(zone, "nodeSelector", "kubernetes.io/hostname")); host != fmt.Sprintf("host-%d", i+1) {
				t.Error("topology must use the actual hostname label for each requested Node", component, host)
			}
		}
	}
	for _, path := range [][]string{{"scheduling", "nodeSelector", "kubernetes.io/hostname"}, {"dashboard", "nodeSelector", "kubernetes.io/hostname"}} {
		if got := pfString(pfGet(values, path...)); got != "host-1" {
			t.Error("storage and Dashboard must stay on the first selected worker", path, got)
		}
	}
}

func TestThreeNodeRunPersistsTopologyAndPerInstanceSizing(t *testing.T) {
	f := newThreeNodeRunFixture(t)
	threeNodeReadOnlySSH(t, f, false)
	out, err := f.run(t, "--nodes", "worker-1,worker-2,worker-3", "--oc", "4", "--om", "16Gi", "--ods", "80Gi", "--lc", "2", "--lm", "6Gi")
	if err != nil {
		t.Fatalf("three-node cached-image fixture installation failed: %v\n%s", err, out)
	}
	values := f.persisted(t)
	assertThreeNodeRunPlacement(t, values)
	for path, expected := range map[string]string{
		"observer.resource.cpu": "4", "observer.resource.memory": "16Gi", "observer.dataSize": "80Gi",
		"logService.resource.cpu": "2", "logService.resource.memory": "6Gi",
	} {
		if got := pfString(pfGet(values, strings.Split(path, ".")...)); got != expected {
			t.Errorf("per-instance sizing %s = %q, want %q", path, got, expected)
		}
	}
	for _, node := range []string{"worker-1", "worker-2", "worker-3"} {
		if !strings.Contains(out, "Node: "+node) {
			t.Error("capacity summary did not include selected worker", node, out)
		}
	}
}

func TestThreeNodeRunCheckVisitsEveryWorkerWithoutWriting(t *testing.T) {
	for _, cold := range []bool{false, true} {
		t.Run(fmt.Sprintf("remote-cold-%t", cold), func(t *testing.T) {
			f := newThreeNodeRunFixture(t)
			threeNodeReadOnlySSH(t, f, cold)
			out, err := f.run(t, "--nodes", "worker-1,worker-2,worker-3", "--ssh-host", "worker-2=node-two.example", "--check")
			if cold {
				if err == nil || !strings.Contains(err.Error(), "worker-2") || !strings.Contains(err.Error(), "worker-3") || !strings.Contains(err.Error(), "images are not ready") {
					t.Fatalf("read-only check did not report both workers' missing image IDs: %v\n%s", err, out)
				}
			} else if err != nil {
				t.Fatalf("all-cached read-only check failed: %v\n%s", err, out)
			}
			trace, _ := os.ReadFile(f.trace)
			for _, host := range []string{"node-two.example", "192.0.2.3"} {
				if !strings.Contains(string(trace), "ssh "+host+" 'docker' '--host' 'unix:///var/run/docker.sock' 'image' 'inspect'") {
					t.Error("read-only check did not inspect locked image IDs on worker", host, string(trace))
				}
			}
			for _, forbidden := range []string{"scp ", "'mktemp'", "'rm'", "'load'", "'import'", " apply ", " upgrade ", " install "} {
				if strings.Contains(string(trace), forbidden) {
					t.Fatal("read-only check attempted a remote or cluster mutation", forbidden, string(trace))
				}
			}
			for _, path := range []string{f.state, f.cache, f.marker} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatal("read-only check created local installation state", path, err)
				}
			}
		})
	}
}

func TestThreeNodeRunResumesSavedAndInstalledTopologyWithoutNodesFlag(t *testing.T) {
	for _, installed := range []bool{false, true} {
		t.Run(fmt.Sprintf("installed-%t", installed), func(t *testing.T) {
			f := newThreeNodeRunFixture(t)
			threeNodeReadOnlySSH(t, f, false)
			if out, err := f.run(t, "--nodes", "worker-1,worker-2,worker-3"); err != nil {
				t.Fatalf("first three-node fixture run failed: %v\n%s", err, out)
			}
			first := f.persisted(t)
			if installed {
				f.responses["helm list -n demo-ns --all --filter ^demo$ -o json"] = []interface{}{map[string]interface{}{"name": "demo"}}
				f.responses["helm get values demo -n demo-ns -o json"] = first
				f.responses["helm get values demo -n demo-ns --all -o json"] = first
				f.syncCommands(t)
			}
			if out, err := f.run(t); err != nil {
				t.Fatalf("retry without --nodes did not infer retained three-node placement: %v\n%s", err, out)
			}
			second := f.persisted(t)
			assertThreeNodeRunPlacement(t, second)
			for _, component := range []string{"observer", "logService"} {
				if pfString(pfGet(first, component, "clusterId")) != pfString(pfGet(second, component, "clusterId")) {
					t.Error("retry regenerated database identity", component)
				}
			}
		})
	}
}

func TestThreeNodeRunRejectsExistingSingleWorkerBeforeImagesOrState(t *testing.T) {
	f := newThreeNodeRunFixture(t)
	installed := mergeValues(f.defaults, map[string]interface{}{})
	setValue(installed, "host-1", "scheduling", "nodeSelector", "kubernetes.io/hostname")
	f.responses["helm list -n demo-ns --all --filter ^demo$ -o json"] = []interface{}{map[string]interface{}{"name": "demo"}}
	f.responses["helm get values demo -n demo-ns -o json"] = installed
	f.responses["helm get values demo -n demo-ns --all -o json"] = installed
	f.syncCommands(t)
	out, err := f.run(t, "--nodes", "worker-1,worker-2,worker-3")
	if err == nil || strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatalf("existing single-worker release was converted by the three-node installer: %v\n%s", err, out)
	}
	trace, _ := os.ReadFile(f.trace)
	for _, forbidden := range []string{"docker ", "ssh ", "scp "} {
		if strings.Contains(string(trace), forbidden) {
			t.Fatal("unsupported migration reached image or remote preparation", string(trace))
		}
	}
	for _, path := range []string{f.state, f.marker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("unsupported migration wrote state or ran the chart", path, err)
		}
	}
}

func threeNodeInstalledValues(f *resourceRunFixture) map[string]interface{} {
	values := mergeValues(f.defaults, map[string]interface{}{})
	for _, component := range []string{"observer", "logService"} {
		var zones []interface{}
		for i := 0; i < 3; i++ {
			zones = append(zones, map[string]interface{}{
				"zone": "zone" + string(rune('1'+i)), "replica": 1,
				"nodeSelector": map[string]interface{}{"kubernetes.io/hostname": "host-" + string(rune('1'+i))},
			})
		}
		setValue(values, zones, component, "topology")
	}
	setValue(values, "host-1", "scheduling", "nodeSelector", "kubernetes.io/hostname")
	setValue(values, "host-1", "dashboard", "nodeSelector", "kubernetes.io/hostname")
	return values
}

func TestThreeNodeRunRejectsPlacementChangesWithoutRelyingOnRunningPods(t *testing.T) {
	for _, mode := range []string{"swapped-node-order", "saved-swapped-node-order", "zone-renamed", "changed-zone-worker", "changed-storage-worker"} {
		t.Run(mode, func(t *testing.T) {
			f := newThreeNodeRunFixture(t)
			installed := threeNodeInstalledValues(f)
			if mode != "saved-swapped-node-order" {
				f.responses["helm list -n demo-ns --all --filter ^demo$ -o json"] = []interface{}{map[string]interface{}{"name": "demo"}}
				f.responses["helm get values demo -n demo-ns -o json"] = installed
				f.responses["helm get values demo -n demo-ns --all -o json"] = installed
			}
			f.syncCommands(t)
			resourceTestWrite(t, filepath.Join(f.state, "values.yaml"), installed)
			before, err := os.ReadFile(filepath.Join(f.state, "values.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--nodes", "worker-1,worker-2,worker-3"}
			if strings.HasSuffix(mode, "swapped-node-order") {
				args[1] = "worker-1,worker-3,worker-2"
			} else {
				overlay := threeNodeInstalledValues(f)
				zones := pfGet(overlay, "observer", "topology").([]interface{})
				switch mode {
				case "zone-renamed":
					zones[0].(map[string]interface{})["zone"] = "moved-zone"
				case "changed-zone-worker":
					zones[0].(map[string]interface{})["nodeSelector"] = map[string]interface{}{"kubernetes.io/hostname": "host-2"}
				case "changed-storage-worker":
					setValue(overlay, "host-2", "scheduling", "nodeSelector", "kubernetes.io/hostname")
				}
				path := filepath.Join(t.TempDir(), "migration.yaml")
				resourceTestWrite(t, path, overlay)
				args = append(args, "-f", path)
			}
			out, runErr := f.run(t, args...)
			if runErr == nil || strings.Contains(runErr.Error(), "flag provided but not defined") {
				t.Fatalf("unsupported retained topology change was accepted: %v\n%s", runErr, out)
			}
			trace, _ := os.ReadFile(f.trace)
			for _, forbidden := range []string{"docker ", "ssh ", "scp "} {
				if strings.Contains(string(trace), forbidden) {
					t.Fatal("unsupported placement reached image preparation", string(trace))
				}
			}
			after, err := os.ReadFile(filepath.Join(f.state, "values.yaml"))
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("rejected placement changed saved values", err)
			}
			if _, err := os.Stat(f.marker); !os.IsNotExist(err) {
				t.Fatal("rejected placement launched the chart", err)
			}
		})
	}
}

func TestThreeNodeOptionsRejectInvalidSelectionBeforeExternalCommands(t *testing.T) {
	for _, args := range [][]string{
		{"--nodes", ""},
		{"--nodes", "worker-1"},
		{"--nodes", "worker-1,worker-2"},
		{"--nodes", "worker-1,worker-2,worker-3,worker-4"},
		{"--nodes", "worker-1,worker-1,worker-3"},
		{"--nodes", "worker-1,,worker-3"},
		{"--nodes", "worker-1,worker-2,worker-3", "--node", "worker-1"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			root := t.TempDir()
			trace := filepath.Join(root, "unexpected-command")
			for _, tool := range []string{"helm", "kubectl", "jq", "bash", "ssh", "scp", "docker", "k3s", "ctr"} {
				body := "#!/bin/sh\nprintf ran > " + resourceTestQuote(trace) + "\nexit 97\n"
				if err := os.WriteFile(filepath.Join(root, tool), []byte(body), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", root)
			err := runWithCacheDirectory(context.Background(), args, io.Discard, io.Discard, func() (string, error) {
				t.Error("invalid node options must not locate installation cache")
				return filepath.Join(root, "cache"), nil
			})
			if err == nil || strings.Contains(err.Error(), "flag provided but not defined") {
				t.Fatal("invalid node selection was not rejected by the supported option parser", err)
			}
			if _, err := os.Stat(trace); !os.IsNotExist(err) {
				t.Fatal("invalid selection executed an external command", err)
			}
		})
	}
}

func TestThreeNodeRunCapacityFailurePrecedesImageAndStateChanges(t *testing.T) {
	f := newThreeNodeRunFixture(t)
	nodes := f.responses["kubectl get nodes -o json"].(corev1.NodeList)
	nodes.Items[2].Status.Allocatable[corev1.ResourceCPU] = resource.MustParse("500m")
	f.responses["kubectl get nodes -o json"] = nodes
	f.responses["kubectl get node worker-3 -o json"] = nodes.Items[2]
	f.syncCommands(t)
	out, err := f.run(t, "--nodes", "worker-1,worker-2,worker-3")
	if err == nil || !strings.Contains(err.Error(), "lacks resources") || !strings.Contains(err.Error(), "worker-3") {
		t.Fatalf("third-worker capacity must fail before installation: %v\n%s", err, out)
	}
	trace, _ := os.ReadFile(f.trace)
	for _, forbidden := range []string{"docker ", "ssh ", "scp "} {
		if strings.Contains(string(trace), forbidden) {
			t.Fatal("images/remote commands ran before all worker capacity checks passed", string(trace))
		}
	}
	for _, path := range []string{f.state, f.marker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("capacity rejection wrote values or launched the chart", path, err)
		}
	}
}
