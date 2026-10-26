package lakehouse

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// These tests run the real CLI/merge/preflight/state path with command binaries
// restricted to fixture responses. They do not use the developer's kubeconfig,
// Docker socket or any deployment. The final chart script only records which
// persisted values file it would receive.
type resourceRunFixture struct {
	bundle, state, cache, trace, marker, bin string
	responses                                map[string]interface{}
	defaults                                 map[string]interface{}
}

func resourceTestQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func resourceTestWrite(t *testing.T, path string, value interface{}) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	var data []byte
	var err error
	if text, ok := value.(string); ok {
		data = []byte(text)
	} else {
		data, err = json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func newResourceRunFixture(t *testing.T) *resourceRunFixture {
	t.Helper()
	root := t.TempDir()
	f := &resourceRunFixture{
		bundle: filepath.Join(root, "bundle"), cache: filepath.Join(root, "cold-cache"),
		trace: filepath.Join(root, "commands.log"), marker: filepath.Join(root, "chart-values-path"),
		bin: filepath.Join(root, "bin"), responses: pfTestFixture(), defaults: pfTestValues(),
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte("test-cluster/demo-ns/demo")))[:16]
	f.state = filepath.Join(f.bundle, ".lakehouse-install", key)
	node := pfTestNode()
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatal(err)
	}
	localIP := ""
	for _, address := range addresses {
		ip, _, parseErr := net.ParseCIDR(address.String())
		if parseErr == nil && !ip.IsLoopback() && !ip.IsUnspecified() {
			localIP = ip.String()
			break
		}
	}
	if localIP == "" {
		t.Skip("the real local-worker guard needs a non-loopback host interface")
	}
	node.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: localIP}}
	node.Status.Allocatable[corev1.ResourceCPU] = resource.MustParse("32")
	node.Status.Allocatable[corev1.ResourceMemory] = resource.MustParse("128Gi")
	node.Status.NodeInfo.ContainerRuntimeVersion = "docker://24.0.0"
	f.responses["kubectl get nodes -o json"] = corev1.NodeList{Items: []corev1.Node{node}}
	f.responses["kubectl get node worker-1 -o json"] = node
	f.responses["kubectl get namespace kube-system -o json"] = map[string]interface{}{"metadata": map[string]string{"uid": "test-cluster"}}
	f.responses["helm list -n demo-ns --all --filter ^demo$ -o json"] = []interface{}{}
	f.responses["helm get values demo -n demo-ns -o json"] = map[string]interface{}{}
	f.responses["helm get values demo -n demo-ns --all -o json"] = f.defaults
	f.responses["docker --host unix:///var/run/docker.sock info --format {{.ServerVersion}}"] = "24.0.0\n"
	setValue(f.defaults, false, "dashboard", "enabled")
	setValue(f.defaults, "40Gi", "observer", "dataSize")
	setValue(f.defaults, "20Gi", "observer", "logSize")
	setValue(f.defaults, "20Gi", "logService", "storeSize")
	setValue(f.defaults, "10Gi", "logService", "logSize")
	setValue(f.defaults, "100Gi", "minio", "size")
	var images []lockedImage
	var refs []string
	for i, component := range []string{"observer", "logService", "minio", "initialization"} {
		role := map[string]string{"observer": "observer", "logService": "logservice", "minio": "minio", "initialization": "mc"}[component]
		ref := "fixture/" + role + ":1"
		id := "sha256:" + strings.Repeat(fmt.Sprintf("%x", i+1), 64)
		setValue(f.defaults, ref, component, "image")
		images = append(images, lockedImage{Role: role, Reference: ref, ImageID: id})
		refs = append(refs, ref)
		f.responses["docker --host unix:///var/run/docker.sock image inspect --format {{.Id}} "+ref] = id + "\n"
	}
	resourceTestWrite(t, filepath.Join(f.bundle, "charts/oceanbase-lakehouse/values.yaml"), f.defaults)
	resourceTestWrite(t, filepath.Join(f.bundle, "charts/oceanbase-dashboard/values.yaml"), "{}\n")
	resourceTestWrite(t, filepath.Join(f.bundle, "values/quickstart.yaml"), f.defaults)
	resourceTestWrite(t, filepath.Join(f.bundle, "image-lock.json"), imageLock{Images: images})
	resourceTestWrite(t, filepath.Join(f.bundle, "images/archives.json"), archiveManifest{SchemaVersion: 1, Archives: []imageArchive{{Name: "fixture", Base: "bundle", Path: "images/fixture.tar", SHA256: strings.Repeat("a", 64), References: refs}}})
	resourceTestWrite(t, filepath.Join(f.bundle, "charts/oceanbase-lakehouse/install.sh"), "#!/bin/sh\nset -eu\n[ \"$1\" = demo ] && [ \"$2\" = demo-ns ] && [ \"$3\" = -f ]\nprintf '%s\\n' \"$4\" > \"$RESOURCE_OPTIONS_CHART_MARKER\"\n")
	t.Setenv("RESOURCE_OPTIONS_CHART_MARKER", f.marker)
	t.Setenv("PATH", f.bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	f.syncCommands(t)
	return f
}

func (f *resourceRunFixture) syncCommands(t *testing.T) {
	t.Helper()
	if err := os.MkdirAll(f.bin, 0700); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(f.responses))
	for key := range f.responses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, tool := range []string{"kubectl", "helm", "docker", "jq"} {
		var script strings.Builder
		script.WriteString("#!/bin/sh\n")
		if tool == "kubectl" {
			script.WriteString("case \"${1-}\" in --request-timeout=*) shift ;; esac\n")
		}
		fmt.Fprintf(&script, "printf '%%s\\n' %s\"$*\" >> %s\ncase \"$*\" in\n", resourceTestQuote(tool+" "), resourceTestQuote(f.trace))
		for i, key := range keys {
			if !strings.HasPrefix(key, tool+" ") {
				continue
			}
			path := filepath.Join(f.bin, fmt.Sprintf("response-%d.json", i))
			resourceTestWrite(t, path, f.responses[key])
			fmt.Fprintf(&script, "%s) /bin/cat %s ;;\n", resourceTestQuote(strings.TrimPrefix(key, tool+" ")), resourceTestQuote(path))
		}
		script.WriteString("*) printf 'Unexpected fixture command: %s\\n' \"$*\" >&2; exit 97 ;;\nesac\n")
		if err := os.WriteFile(filepath.Join(f.bin, tool), []byte(script.String()), 0700); err != nil {
			t.Fatal(err)
		}
	}
}

func (f *resourceRunFixture) run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out, stderr bytes.Buffer
	base := []string{"--bundle", f.bundle, "--release", "demo", "--namespace", "demo-ns", "--progress", "plain"}
	err := runWithCacheDirectory(context.Background(), append(base, args...), &out, &stderr, func() (string, error) {
		for _, arg := range args {
			if arg == "--check" {
				t.Error("--check must not even locate the local installation lock cache")
			}
		}
		return f.cache, nil
	})
	return out.String() + stderr.String(), err
}

func (f *resourceRunFixture) persisted(t *testing.T) map[string]interface{} {
	t.Helper()
	path := filepath.Join(f.state, "values.yaml")
	value, err := readValues(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("persisted values must remain private", err)
	}
	chartPath, err := os.ReadFile(f.marker)
	if err != nil || strings.TrimSpace(string(chartPath)) != path {
		t.Fatal("chart did not receive the validated persisted configuration", err, string(chartPath))
	}
	return value
}

var resourceIntegrationOptions = []struct {
	flag, value string
	path        []string
}{
	{"--observer-cpu", "4", []string{"observer", "resource", "cpu"}},
	{"--observer-memory", "16Gi", []string{"observer", "resource", "memory"}},
	{"--observer-data-size", "80Gi", []string{"observer", "dataSize"}},
	{"--observer-log-size", "30Gi", []string{"observer", "logSize"}},
	{"--logservice-cpu", "2", []string{"logService", "resource", "cpu"}},
	{"--logservice-memory", "6Gi", []string{"logService", "resource", "memory"}},
	{"--logservice-store-size", "35Gi", []string{"logService", "storeSize"}},
	{"--logservice-log-size", "15Gi", []string{"logService", "logSize"}},
	{"--minio-size", "120Gi", []string{"minio", "size"}},
}

func resourceIntegrationArgs() []string {
	var args []string
	for _, option := range resourceIntegrationOptions {
		args = append(args, option.flag, option.value)
	}
	return args
}

func assertResourceIntegrationValues(t *testing.T, values map[string]interface{}) {
	t.Helper()
	for _, option := range resourceIntegrationOptions {
		if got := pfString(pfGet(values, option.path...)); got != option.value {
			t.Errorf("%s did not reach the persisted chart values: got %q, want %q", option.flag, got, option.value)
		}
	}
}

func TestResourceOptionsReachPreflightAndPersistedChartValues(t *testing.T) {
	f := newResourceRunFixture(t)
	out, err := f.run(t, resourceIntegrationArgs()...)
	if err != nil {
		t.Fatalf("isolated install failed: %v\n%s", err, out)
	}
	assertResourceIntegrationValues(t, f.persisted(t))
	if !strings.Contains(out, "CPU: 10.15 cores required / 32.00 available; memory: 34.6 GiB required / 128.0 available") {
		t.Fatal("preflight did not use overridden Observer and all three LogService replicas", out)
	}
}

func TestResourceOptionsOverrideAllValuesFilesRegardlessOfFlagPosition(t *testing.T) {
	for _, cliFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("cli-first-%t", cliFirst), func(t *testing.T) {
			f := newResourceRunFixture(t)
			first := filepath.Join(t.TempDir(), "first.yaml")
			last := filepath.Join(t.TempDir(), "last.yaml")
			resourceTestWrite(t, first, "observer:\n  resource: {cpu: '1', memory: 8Gi}\n  dataSize: 11Gi\n  logSize: 12Gi\nlogService:\n  resource: {cpu: '1', memory: 3Gi}\n  storeSize: 13Gi\n  logSize: 14Gi\nminio: {size: 15Gi}\n")
			resourceTestWrite(t, last, "observer:\n  resource: {cpu: '3', memory: 12Gi}\n  dataSize: 61Gi\n  logSize: 21Gi\nlogService:\n  resource: {cpu: '1', memory: 5Gi}\n  storeSize: 22Gi\n  logSize: 11Gi\nminio: {size: 101Gi}\n")
			files := []string{"-f", first, "--values", last}
			args := append(files, resourceIntegrationArgs()...)
			if cliFirst {
				args = append(resourceIntegrationArgs(), files...)
			}
			out, err := f.run(t, args...)
			if err != nil {
				t.Fatalf("values/CLI precedence failed: %v\n%s", err, out)
			}
			assertResourceIntegrationValues(t, f.persisted(t))
		})
	}
}

func TestResourceOptionsRetainUnspecifiedSavedValuesOnRetry(t *testing.T) {
	f := newResourceRunFixture(t)
	if out, err := f.run(t, resourceIntegrationArgs()...); err != nil {
		t.Fatalf("initial fixture run failed: %v\n%s", err, out)
	}
	first := f.persisted(t)
	if out, err := f.run(t); err != nil {
		t.Fatalf("retry without resource flags failed: %v\n%s", err, out)
	}
	second := f.persisted(t)
	assertResourceIntegrationValues(t, second)
	for _, component := range []string{"observer", "logService"} {
		if pfString(pfGet(first, component, "clusterId")) != pfString(pfGet(second, component, "clusterId")) {
			t.Fatal("retry regenerated the database identity", component)
		}
	}
}

func TestResourceOptionsRetainInstalledEffectiveValuesAndRejectResize(t *testing.T) {
	f := newResourceRunFixture(t)
	effective := mergeValues(f.defaults, map[string]interface{}{})
	for _, option := range resourceIntegrationOptions {
		setValue(effective, option.value, option.path...)
	}
	f.responses["helm list -n demo-ns --all --filter ^demo$ -o json"] = []interface{}{map[string]interface{}{"name": "demo"}}
	f.responses["helm get values demo -n demo-ns --all -o json"] = effective
	f.syncCommands(t)
	// Equal quantities must retain the installed representation too: MinIO's
	// textual size participates in the retained storage identity.
	out, err := f.run(t, "--observer-cpu", "4000m", "--minio-size", "122880Mi")
	if err != nil {
		t.Fatalf("installed-release update failed: %v\n%s", err, out)
	}
	values := f.persisted(t)
	assertResourceIntegrationValues(t, values)
	before, err := os.ReadFile(filepath.Join(f.state, "values.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(f.marker); err != nil {
		t.Fatal(err)
	}
	resourceTestWrite(t, f.trace, "")
	out, err = f.run(t, "--observer-cpu", "5")
	if err == nil || !strings.Contains(err.Error(), "existing release") {
		t.Fatalf("unsupported resize was not rejected: %v\n%s", err, out)
	}
	after, err := os.ReadFile(filepath.Join(f.state, "values.yaml"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("rejected resize changed saved configuration", err)
	}
	trace, _ := os.ReadFile(f.trace)
	if strings.Contains(string(trace), "docker ") || strings.Contains(string(trace), "kubectl version") {
		t.Fatal("unsupported resize reached image preparation or cluster preflight", string(trace))
	}
	if _, err := os.Stat(f.marker); !os.IsNotExist(err) {
		t.Fatal("rejected resize ran the chart", err)
	}
}

func TestResourceOptionsCapacityFailurePrecedesImageAndStateChanges(t *testing.T) {
	f := newResourceRunFixture(t)
	out, err := f.run(t, "--observer-cpu", "32")
	if err == nil || !strings.Contains(err.Error(), "lacks resources") {
		t.Fatalf("oversized explicit CPU did not reach the capacity gate: %v\n%s", err, out)
	}
	trace, _ := os.ReadFile(f.trace)
	if strings.Contains(string(trace), "docker ") {
		t.Fatal("image work started before resource capacity validation", string(trace))
	}
	for _, path := range []string{f.state, f.marker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("capacity rejection wrote state or launched the chart", path, err)
		}
	}
}

func TestResourceOptionsFractionalCPUReachesChartAsMillicores(t *testing.T) {
	f := newResourceRunFixture(t)
	out, err := f.run(t, "--observer-cpu", "1.5")
	if err != nil {
		t.Fatalf("fractional CPU install failed: %v\n%s", err, out)
	}
	values := f.persisted(t)
	if got := pfString(pfGet(values, "observer", "resource", "cpu")); got != "1500m" {
		t.Fatal("fractional cores must use the chart-compatible millicore representation", got)
	}
	if !strings.Contains(out, "CPU: 4.65 cores required") {
		t.Fatal("capacity check did not use the fractional CPU override", out)
	}
}

func TestResourceOptionsMemoryDiskConflictFailsBeforeClusterAndImageWork(t *testing.T) {
	f := newResourceRunFixture(t)
	out, err := f.run(t, "--observer-memory", "20Gi", "--observer-data-size", "50Gi")
	if err == nil || !strings.Contains(err.Error(), "three times") {
		t.Fatalf("unsupported memory/disk combination reached deployment: %v\n%s", err, out)
	}
	trace, _ := os.ReadFile(f.trace)
	if strings.Contains(string(trace), "docker ") || strings.Contains(string(trace), "kubectl version") {
		t.Fatal("invalid memory/disk sizing reached cluster preflight or image preparation", string(trace))
	}
	for _, path := range []string{f.state, f.marker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("invalid memory/disk sizing wrote values or ran the chart", path, err)
		}
	}
}

func TestResourceOptionsInvalidInputsFailBeforeAnyExternalCommand(t *testing.T) {
	for _, option := range resourceIntegrationOptions {
		for _, mode := range []string{"zero", "show-access"} {
			t.Run(strings.TrimPrefix(option.flag, "--")+"-"+mode, func(t *testing.T) {
				root := t.TempDir()
				bin, trace := filepath.Join(root, "bin"), filepath.Join(root, "command-ran")
				if err := os.Mkdir(bin, 0700); err != nil {
					t.Fatal(err)
				}
				for _, tool := range []string{"kubectl", "helm", "jq", "bash", "docker"} {
					body := "#!/bin/sh\nprintf ran > " + resourceTestQuote(trace) + "\nexit 99\n"
					if err := os.WriteFile(filepath.Join(bin, tool), []byte(body), 0700); err != nil {
						t.Fatal(err)
					}
				}
				t.Setenv("PATH", bin)
				args := []string{"--bundle", filepath.Join(root, "bundle"), option.flag, "0"}
				if mode == "show-access" {
					args = []string{"--bundle", filepath.Join(root, "bundle"), "--show-access", option.flag, option.value}
				}
				err := runWithCacheDirectory(context.Background(), args, io.Discard, io.Discard, func() (string, error) {
					t.Error("invalid options reached the installation lock cache")
					return filepath.Join(root, "cache"), nil
				})
				if err == nil || strings.Contains(err.Error(), "flag provided but not defined") {
					t.Fatal("resource input must be rejected by validation, not an unregistered flag", args, err)
				}
				if _, err := os.Stat(trace); !os.IsNotExist(err) {
					t.Fatal("invalid input invoked an external command", err)
				}
				for _, path := range []string{filepath.Join(root, "bundle"), filepath.Join(root, "cache")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatal("invalid input created installation state", path, err)
					}
				}
			})
		}
	}
}

func TestResourceOptionsColdCacheCheckDoesNotWriteStateOrImportImages(t *testing.T) {
	f := newResourceRunFixture(t)
	for command := range f.responses {
		if strings.Contains(command, " image inspect ") {
			// A wrong/missing cached config ID is a cold runtime; --check must
			// report it and must never import or even create local state.
			f.responses[command] = "sha256:" + strings.Repeat("f", 64)
		}
	}
	f.syncCommands(t)
	args := append(resourceIntegrationArgs(), "--check")
	out, err := f.run(t, args...)
	if err == nil || !strings.Contains(err.Error(), "images are not ready") || !strings.Contains(out, "10.15 cores required") {
		t.Fatalf("cold read-only check did not validate the selected resources and images: %v\n%s", err, out)
	}
	trace, _ := os.ReadFile(f.trace)
	for _, forbidden := range []string{" load ", " import ", " apply ", " install ", " upgrade "} {
		if strings.Contains(string(trace), forbidden) {
			t.Fatal("read-only preflight attempted a mutation", forbidden, string(trace))
		}
	}
	for _, path := range []string{filepath.Join(f.bundle, ".lakehouse-install"), f.cache, f.marker} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("cold-cache --check changed local state", path, err)
		}
	}
}
