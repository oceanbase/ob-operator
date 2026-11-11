/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package lakehouse_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type service struct {
	Command     []string          `json:"command"`
	Image       string            `json:"image"`
	Build       any               `json:"build"`
	Environment map[string]string `json:"environment"`
	DependsOn   map[string]struct {
		Condition string `json:"condition"`
	} `json:"depends_on"`
	Volumes []struct {
		Type     string `json:"type"`
		Source   string `json:"source"`
		Target   string `json:"target"`
		ReadOnly bool   `json:"read_only"`
	} `json:"volumes"`
	MemoryLimit json.RawMessage `json:"mem_limit"`
	Ports       []struct {
		HostIP    string `json:"host_ip"`
		Published string `json:"published"`
	} `json:"ports"`
	NetworkMode string `json:"network_mode"`
}

func renderCompose(t *testing.T) map[string]service {
	t.Helper()
	return renderComposeEnvironment(t, nil, "", true)
}

func renderComposeEnvironment(t *testing.T, overrides map[string]string, envFile string, explicitEnvFile bool) map[string]service {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker Compose CLI is required for configuration validation")
	}
	// Copy only the public configuration/scripts, proving it does not depend
	// on .env, source code, local secrets or preparation outputs.
	dir := t.TempDir()
	body, err := os.ReadFile("compose.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte(envFile), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"compose"}
	if explicitEnvFile {
		args = append(args, "--env-file", envPath)
	}
	args = append(args, "-f", filepath.Join(dir, "compose.yaml"), "config", "--format", "json")
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir
	// Defaults must not accidentally inherit a developer's Compose overrides.
	for _, kv := range os.Environ() {
		name := strings.SplitN(kv, "=", 2)[0]
		if name == "PATH" || name == "HOME" || name == "DOCKER_CONFIG" || name == "DOCKER_CLI_PLUGIN_EXTRA_DIRS" {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	for name, value := range overrides {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compose config: %v\n%s", err, out)
	}
	var cfg struct {
		Services map[string]service `json:"services"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg.Services
}

func TestQuickstartWithoutPreparation(t *testing.T) {
	services := renderCompose(t)
	if len(services) != 13 {
		t.Fatalf("expected eight services and five initialization containers, got %d", len(services))
	}
	for name, s := range services {
		if s.Image == "" || s.Build != nil {
			t.Fatalf("%s requires a missing image or local build", name)
		}
		if strings.HasSuffix(s.Image, ":latest") {
			t.Fatalf("%s uses floating image", name)
		}
		for _, mount := range s.Volumes {
			if mount.Type == "bind" && !strings.HasSuffix(mount.Source, "/scripts") &&
				!strings.Contains(mount.Source, "/monitoring/") {
				t.Fatalf("%s needs host-prepared files: %+v", name, mount)
			}
		}
	}
	for _, tc := range []struct{ service, dependency, condition string }{
		{"minio", "credentials-init", "service_completed_successfully"},
		{"host-info-init", "credentials-init", "service_completed_successfully"},
		{"minio-init", "minio", "service_healthy"},
		{"logservice-bootstrap", "minio-init", "service_completed_successfully"},
		{"logservice-bootstrap", "logservice-1", "service_healthy"},
		{"logservice-bootstrap", "logservice-2", "service_healthy"},
		{"logservice-bootstrap", "logservice-3", "service_healthy"},
		{"observer", "logservice-bootstrap", "service_completed_successfully"},
		{"observer-bootstrap", "observer", "service_healthy"},
		{"observer-bootstrap", "host-info-init", "service_completed_successfully"},
		{"obagent", "observer-bootstrap", "service_completed_successfully"},
		{"prometheus", "obagent", "service_healthy"},
		{"prometheus", "alertmanager", "service_healthy"},
	} {
		if services[tc.service].DependsOn[tc.dependency].Condition != tc.condition {
			t.Fatalf("%s must wait for %s (%s)", tc.service, tc.dependency, tc.condition)
		}
	}
	if services["obagent"].NetworkMode != "service:observer" {
		t.Fatalf("OBAgent must share the Observer network namespace, got %q", services["obagent"].NetworkMode)
	}
	if services["host-info-init"].NetworkMode != "host" {
		t.Fatalf("host-info-init must inspect the host network namespace, got %q", services["host-info-init"].NetworkMode)
	}
	ports := services["observer"].Ports
	foundSQL, foundObshell := false, false
	for _, port := range ports {
		switch port.Published {
		case "32881":
			foundSQL = port.HostIP == "0.0.0.0"
		case "32886":
			foundObshell = port.HostIP == "0.0.0.0"
		}
	}
	if !foundSQL || !foundObshell {
		t.Fatalf("expected SQL on 0.0.0.0:32881 and obshell on 0.0.0.0:32886, got %+v", ports)
	}
	if got := services["observer-bootstrap"].Environment["OBSHELL_HOST_PORT"]; got != "32886" {
		t.Fatalf("observer-bootstrap OBSHELL_HOST_PORT = %q, want 32886", got)
	}
	observer := services["observer"]
	if observer.Environment["AUTO_REGISTER_MONITORING"] != "true" ||
		observer.Environment["PROMETHEUS_INTERNAL_URL"] != "http://prometheus:9090/" ||
		observer.Environment["ALERTMANAGER_INTERNAL_URL"] != "http://alertmanager:9093" {
		t.Fatalf("observer monitoring auto-registration defaults are incomplete: %+v", observer.Environment)
	}
}

func TestEntrypointScriptsAreExecutable(t *testing.T) {
	for _, name := range []string{
		"minio-entrypoint.sh", "minio-init.sh", "logservice-entrypoint.sh",
		"logservice-bootstrap.sh", "observer-entrypoint.sh", "observer-bootstrap.sh",
		"obagent-entrypoint.sh", "host-info-init.sh", "obshell-monitoring-register.sh",
	} {
		info, err := os.Stat(filepath.Join("scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0111 == 0 {
			t.Errorf("%s is configured as an entrypoint but is not executable", name)
		}
	}
	bootstrap, err := os.ReadFile(filepath.Join("scripts", "observer-bootstrap.sh"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`host_ip="$(cat /run/secrets/docker_host_ip)"`,
		"${host_ip}:${OBSERVER_SQL_HOST_PORT:-32881}",
		"http://${host_ip}:${OBSHELL_HOST_PORT:-32886}/",
		"http://127.0.0.1:${PROMETHEUS_HOST_PORT:-39090}/",
		"http://127.0.0.1:${ALERTMANAGER_HOST_PORT:-39093}/",
		"docker compose run --rm --no-deps observer-bootstrap --mode show-password",
	} {
		if !strings.Contains(string(bootstrap), want) {
			t.Fatalf("observer bootstrap completion output is missing %q", want)
		}
	}
}

func TestObshellUsesTheRetainedDatabaseCredentialOnEveryStart(t *testing.T) {
	entrypoint, err := os.ReadFile(filepath.Join("scripts", "observer-entrypoint.sh"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(entrypoint)
	for _, required := range []string{
		`root_password="$(cat /run/secrets/ob_root_password)"`,
		`obshell_args=(agent start --ip "${LOCAL_IP}" --password "${root_password}")`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("Observer entrypoint does not always seed obshell with the retained root credential: missing %q", required)
		}
	}
	if strings.Contains(body, `if [[ "${existing_installation}" == true ]]`) {
		t.Fatal("obshell credential must also be supplied during the first installation")
	}

	acceptance, err := os.ReadFile(filepath.Join("scripts", "acceptance.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(acceptance), "obshell cluster show") ||
		!strings.Contains(string(acceptance), "grep -Eq '[[:space:]]ACTIVE[[:space:]]'") {
		t.Fatal("acceptance must verify an obshell operation that reads the authenticated cluster state")
	}
	bootstrap, err := os.ReadFile(filepath.Join("scripts", "observer-bootstrap.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bootstrap), "/run/secrets/observer_bootstrap_complete") ||
		!strings.Contains(body, "/run/secrets/observer_bootstrap_complete") {
		t.Fatal("bootstrap completion must trigger the first authenticated obshell takeover")
	}
	if !strings.HasPrefix(string(bootstrap), "#!/bin/sh\n") || strings.Contains(string(bootstrap), "bash") {
		t.Fatal("bootstrap wrapper must run in the bootstrap image, which only guarantees /bin/sh")
	}
	if !strings.Contains(body, `kill -TERM "${observer_pid}"`) {
		t.Fatal("failed obshell takeover must terminate Observer so the container restart policy can recover")
	}
}

func TestObshellMonitoringRegistrationUsesLocalSocket(t *testing.T) {
	register, err := os.ReadFile(filepath.Join("scripts", "obshell-monitoring-register.sh"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(register)
	for _, want := range []string{
		`--unix-socket "${socket}"`,
		`/api/v1/system/external/${name}`,
		`wait_ready prometheus`,
		`wait_ready alertmanager`,
		`register prometheus`,
		`register alertmanager`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("monitoring registration is missing %q", want)
		}
	}
	entrypoint, err := os.ReadFile(filepath.Join("scripts", "observer-entrypoint.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(entrypoint), `/opt/lakehouse/obshell-monitoring-register.sh &`) {
		t.Fatal("Observer must start automatic monitoring registration in the background")
	}
}

func TestPrivateCredentialsAndBoundedResources(t *testing.T) {
	services := renderCompose(t)
	for _, name := range []string{"credentials-init", "host-info-init", "minio", "minio-init", "logservice-bootstrap", "observer", "observer-bootstrap", "obagent"} {
		found := false
		for _, mount := range services[name].Volumes {
			if mount.Target != "/run/secrets" {
				continue
			}
			found = true
			if mount.Type != "volume" || mount.Source != "credentials" {
				t.Fatalf("%s credentials are not persistent named storage", name)
			}
			if name != "credentials-init" && name != "host-info-init" && name != "observer-bootstrap" && !mount.ReadOnly {
				t.Fatalf("%s can change credentials", name)
			}
		}
		if !found {
			t.Fatalf("%s cannot read credentials", name)
		}
	}
	for _, name := range []string{"logservice-1", "logservice-2", "logservice-3"} {
		s := services[name]
		// Compose versions serialize the same 4 GiB limit as a number,
		// a numeric string, or the original unit string.
		memoryLimit := strings.ToLower(strings.Trim(string(s.MemoryLimit), `"`))
		if (memoryLimit != "4294967296" && memoryLimit != "4g") || s.Environment["MEMORY_LIMIT"] != "3G" {
			t.Fatalf("%s resource defaults lost: memory limit=%s, MEMORY_LIMIT=%q", name, s.MemoryLimit, s.Environment["MEMORY_LIMIT"])
		}
	}
	for name, s := range services {
		if name != "observer" && name != "prometheus" && name != "alertmanager" && len(s.Ports) > 0 {
			t.Fatalf("%s publishes unnecessary host ports", name)
		}
	}
	ports := services["observer"].Ports
	if len(ports) != 2 || ports[0].HostIP != "0.0.0.0" || ports[0].Published != "32881" ||
		ports[1].HostIP != "0.0.0.0" || ports[1].Published != "32886" {
		t.Fatalf("unexpected SQL/obshell exposure %+v", ports)
	}
	if ports := services["prometheus"].Ports; len(ports) != 0 {
		t.Fatalf("unexpected Prometheus exposure %+v", ports)
	}
	if ports := services["alertmanager"].Ports; len(ports) != 0 {
		t.Fatalf("unexpected Alertmanager exposure %+v", ports)
	}
}

// Exercise real Compose interpolation: the same registry selection must reach
// every consumer, work with the shipped .env, and preserve pinned release tags.
func TestImageRegistrySelection(t *testing.T) {
	example, err := os.ReadFile(".env.example")
	if err != nil {
		t.Fatal(err)
	}
	const registry = "registry.example.invalid:5000/team/lakehouse"
	const digest = "private.example.invalid/custom/observer@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	releaseEnv := strings.Replace(string(example),
		"${IMAGE_REGISTRY:-oceanbase}/lakehouse-bootstrap:20261103-monitoring-r1",
		"${IMAGE_REGISTRY:-oceanbase}/lakehouse-bootstrap:native-e43fab5cc687", 1)
	if releaseEnv == string(example) {
		t.Fatal("release fixture must replace the bootstrap reference")
	}
	for _, tc := range []struct {
		name      string
		env       map[string]string
		file      string
		explicit  bool
		prefix    string
		bootstrap string
		override  bool
	}{
		{name: "original defaults"},
		{name: "empty prefix", env: map[string]string{"IMAGE_REGISTRY": ""}},
		{name: "inline or exported registry", env: map[string]string{"IMAGE_REGISTRY": registry}, prefix: registry},
		{name: "copied example", file: string(example)},
		{name: "environment overrides example", file: string(example), env: map[string]string{"IMAGE_REGISTRY": registry}, prefix: registry},
		{name: "registry in dot env", file: strings.Replace(string(example), "IMAGE_REGISTRY=", "IMAGE_REGISTRY="+registry, 1), prefix: registry},
		{name: "explicit env file", file: strings.Replace(string(example), "IMAGE_REGISTRY=", "IMAGE_REGISTRY="+registry, 1), explicit: true, prefix: registry},
		{name: "shell overrides file registry", file: strings.Replace(string(example), "IMAGE_REGISTRY=", "IMAGE_REGISTRY=ignored.example.invalid/old", 1), env: map[string]string{"IMAGE_REGISTRY": registry}, explicit: true, prefix: registry},
		{name: "full digest override", file: string(example), env: map[string]string{"IMAGE_REGISTRY": registry, "OBSERVER_IMAGE": digest}, prefix: registry, override: true},
		{name: "release defaults", file: releaseEnv, bootstrap: "oceanbase/lakehouse-bootstrap:native-e43fab5cc687"},
		{name: "release registry", file: releaseEnv, env: map[string]string{"IMAGE_REGISTRY": registry}, prefix: registry, bootstrap: registry + "/lakehouse-bootstrap:native-e43fab5cc687"},
		{name: "release custom env file", file: strings.Replace(releaseEnv, "IMAGE_REGISTRY=", "IMAGE_REGISTRY="+registry, 1), explicit: true, prefix: registry, bootstrap: registry + "/lakehouse-bootstrap:native-e43fab5cc687"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			services := renderComposeEnvironment(t, tc.env, tc.file, tc.explicit)
			expected := map[string]string{
				"observer":         "oceanbase/oceanbase-ai-cloud-native:4.6.3.0-82026110121.al8-r3",
				"logservice-1":     "oceanbase/oblogservice:1.3.3-42026110118.al8-r1",
				"minio":            "minio/minio:RELEASE.2025-09-07T16-13-09Z",
				"minio-init":       "minio/mc:RELEASE.2025-08-13T08-35-41Z",
				"credentials-init": "oceanbase/lakehouse-bootstrap:20261103-monitoring-r1",
				"obagent":          "oceanbase/obagent:4.2.4-200000022025090416",
				"prometheus":       "quay.io/prometheus/prometheus:v2.51.1",
				"alertmanager":     "quay.io/prometheus/alertmanager:v0.27.0",
			}
			if tc.prefix != "" {
				for name, ref := range expected {
					expected[name] = tc.prefix + "/" + ref[strings.LastIndex(ref, "/")+1:]
				}
			}
			if tc.bootstrap != "" {
				expected["credentials-init"] = tc.bootstrap
			}
			if tc.override {
				expected["observer"] = digest
			}
			for _, name := range []string{"logservice-2", "logservice-3", "logservice-bootstrap"} {
				expected[name] = expected["logservice-1"]
			}
			expected["observer-bootstrap"] = expected["credentials-init"]
			for name, want := range expected {
				if got := services[name].Image; got != want {
					t.Errorf("%s image = %q, want %q", name, got, want)
				}
			}
		})
	}
}

func TestMonitoringIsPrivateByDefault(t *testing.T) {
	services := renderCompose(t)
	for _, name := range []string{"prometheus", "alertmanager"} {
		if len(services[name].Ports) != 0 {
			t.Fatalf("%s must not publish host ports by default", name)
		}
		for _, arg := range services[name].Command {
			if strings.HasPrefix(arg, "--web.enable-lifecycle") {
				t.Fatalf("%s enables lifecycle API", name)
			}
		}
	}
	out, err := exec.Command("docker", "compose", "--env-file", "/dev/null", "-f", "compose.yaml", "-f", "compose.monitoring.yaml", "config", "--format", "json").CombinedOutput()
	if err != nil {
		t.Fatalf("monitoring overlay: %v\n%s", err, out)
	}
	var config struct {
		Services map[string]service `json:"services"`
	}
	if err := json.Unmarshal(out, &config); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"prometheus", "alertmanager"} {
		ports := config.Services[name].Ports
		if len(ports) != 1 || ports[0].HostIP != "127.0.0.1" {
			t.Fatalf("%s monitoring overlay must bind loopback: %+v", name, ports)
		}
	}
}

func TestObserverEntrypointReplacesInheritedPIDBeforeObshell(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		target := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	write("home/run/observer.pid", "99999999\n")
	write("home/store/etc/observer.conf.bin", "retained configuration")
	write("secrets/ob_root_password", "test-fixture-password")
	write("secrets/observer_bootstrap_complete", "done")
	write("home/bin/observer", "#!/bin/bash\nprintf '%s' \"$$\" > \"$QA_ROOT/actual-pid\"\nsleep 2\n")
	write("home/bin/obshell", `#!/bin/bash
if [[ "$2" == start ]]; then
  actual=$(cat "$QA_ROOT/actual-pid")
  recorded=$(cat "$QA_ROOT/home/run/observer.pid")
  [[ "$actual" == "$recorded" ]] || exit 42
  kill -0 "$recorded" || exit 43
  touch "$QA_ROOT/obshell-saw-running-observer"
fi
`)
	// Stub only the SQL TCP probe; all process/PID handling runs in real Bash.
	write("bin/bash", "#!/bin/sh\nexit 0\n")
	entrypoint, err := os.ReadFile(filepath.Join("scripts", "observer-entrypoint.sh"))
	if err != nil {
		t.Fatal(err)
	}
	body := strings.NewReplacer("/home/admin/oceanbase", root+"/home", "/home/admin/data-file", root+"/data", "/home/admin/log", root+"/log", "/run/secrets", root+"/secrets").Replace(string(entrypoint))
	write("entrypoint.sh", body)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/bash", filepath.Join(root, "entrypoint.sh"))
	cmd.Env = append(os.Environ(), "QA_ROOT="+root, "PATH="+root+"/bin:"+os.Getenv("PATH"), "LOCAL_IP=127.0.0.1", "AUTO_REGISTER_MONITORING=false")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("entrypoint failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(root, "obshell-saw-running-observer")); err != nil {
		t.Fatal("obshell could not identify the running Observer:", err)
	}
}
