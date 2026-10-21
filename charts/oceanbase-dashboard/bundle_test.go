package dashboard_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type dashboardManifest = map[string]interface{}

func renderDashboard(t *testing.T, namespace string, args ...string) []dashboardManifest {
	t.Helper()
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	cmd := append([]string{"template", "qa", ".", "-n", namespace}, args...)
	output, err := exec.Command(helm, cmd...).CombinedOutput()
	if err != nil {
		t.Fatalf("helm template: %v\n%s", err, output)
	}
	var docs []dashboardManifest
	decoder := yaml.NewDecoder(bytes.NewReader(output))
	for {
		var doc dashboardManifest
		if err := decoder.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if len(doc) != 0 {
			docs = append(docs, doc)
		}
	}
	return docs
}

func dashboardField(m map[string]interface{}, path ...string) interface{} {
	var result interface{} = m
	for _, name := range path {
		switch value := result.(type) {
		case map[string]interface{}:
			result = value[name]
		default:
			return nil
		}
	}
	return result
}

func dashboardObject(t *testing.T, docs []dashboardManifest, kind, name string) dashboardManifest {
	t.Helper()
	for _, doc := range docs {
		if doc["kind"] == kind && dashboardField(doc, "metadata", "name") == name {
			return doc
		}
	}
	t.Fatalf("missing %s/%s", kind, name)
	return nil
}

func TestDiagnosticImageConfigTriggersDashboardRollout(t *testing.T) {
	checksum := func(args ...string) string {
		docs := renderDashboard(t, "qa-ns", args...)
		deployment := dashboardObject(t, docs, "Deployment", "oceanbase-dashboard-qa")
		value, _ := dashboardField(deployment, "spec", "template", "metadata", "annotations", "checksum/dashboard-config").(string)
		if len(value) != 64 {
			t.Fatal("subPath config changes must trigger a Pod rollout")
		}
		return value
	}
	base := checksum()
	if base != checksum() {
		t.Fatal("unchanged config must not trigger repeated rollouts")
	}
	for _, key := range []string{"inspection.obdiag.image", "inspection.oceanbaseHelper.image"} {
		if checksum("--set-string", key+"=registry.example/diagnostics:fixed") == base {
			t.Fatalf("%s change did not roll out stale subPath config", key)
		}
	}
}

func TestPersistenceOwnershipAndClaimReferences(t *testing.T) {
	for _, tc := range []struct {
		name, prometheusClaim, sharedClaim string
		args                               []string
		pvcCount                           int
	}{
		{"standalone", "qa-prometheus-pvc", "qa-shared-volume-pvc", nil, 2},
		{"parent-owned", "qa-prometheus-pvc", "qa-shared-volume-pvc",
			[]string{"--set", "prometheus.persistence.create=false,sharedVolume.persistence.create=false"}, 0},
		{"user-owned", "user-prometheus", "user-shared",
			[]string{"--set", "prometheus.persistence.existingClaim=user-prometheus,sharedVolume.persistence.existingClaim=user-shared"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			docs := renderDashboard(t, "qa-ns", tc.args...)
			count := 0
			for _, doc := range docs {
				if doc["kind"] == "PersistentVolumeClaim" {
					count++
				}
			}
			if count != tc.pvcCount {
				t.Fatalf("PVC count %d, want %d", count, tc.pvcCount)
			}
			deployment := dashboardObject(t, docs, "Deployment", "oceanbase-dashboard-qa")
			volumes := dashboardField(deployment, "spec", "template", "spec", "volumes").([]interface{})
			found := map[string]interface{}{}
			for _, item := range volumes {
				volume := item.(map[string]interface{})
				found[volume["name"].(string)] = dashboardField(volume, "persistentVolumeClaim", "claimName")
			}
			if found["prometheus-storage-volume"] != tc.prometheusClaim || found["shared-volume"] != tc.sharedClaim {
				t.Fatalf("incorrect mounted claims: %v", found)
			}
			containers := dashboardField(deployment, "spec", "template", "spec", "containers").([]interface{})
			for _, item := range containers[0].(map[string]interface{})["env"].([]interface{}) {
				env := item.(map[string]interface{})
				if env["name"] == "SHARED_VOLUME_PVC_NAME" && env["value"] != tc.sharedClaim {
					t.Fatalf("job shared volume points to %v, want %s", env["value"], tc.sharedClaim)
				}
			}
			cleanup := dashboardObject(t, docs, "CronJob", "qa-cleanup")
			cleanupVolumes := dashboardField(cleanup, "spec", "jobTemplate", "spec", "template", "spec", "volumes").([]interface{})
			if dashboardField(cleanupVolumes[0].(map[string]interface{}), "persistentVolumeClaim", "claimName") != tc.sharedClaim {
				t.Fatal("cleanup does not use the configured shared claim")
			}
		})
	}
}

func TestNamespacedRBACAndImageOverride(t *testing.T) {
	for _, ns := range []string{"first", "second"} {
		docs := renderDashboard(t, ns, "--set", "rbac.includeNamespace=true", "--set-string", "image=example.invalid/dashboard:qa")
		roleName := ns + "-qa-cluster-role"
		dashboardObject(t, docs, "ClusterRole", roleName)
		binding := dashboardObject(t, docs, "ClusterRoleBinding", ns+"-qa-cluster-role-binding")
		if dashboardField(binding, "roleRef", "name") != roleName {
			t.Fatal("binding references a different role")
		}
		cfg := dashboardObject(t, docs, "ConfigMap", "qa-dashboard-config")
		var config struct {
			Inspection struct {
				ClusterRoleName string `yaml:"clusterRoleName"`
			} `yaml:"inspection"`
		}
		if err := yaml.Unmarshal([]byte(dashboardField(cfg, "data", "config.yaml").(string)), &config); err != nil {
			t.Fatal(err)
		}
		if config.Inspection.ClusterRoleName != roleName {
			t.Fatalf("inspection role = %s, want %s", config.Inspection.ClusterRoleName, roleName)
		}
		deployment := dashboardObject(t, docs, "Deployment", "oceanbase-dashboard-qa")
		containers := dashboardField(deployment, "spec", "template", "spec", "containers").([]interface{})
		if containers[0].(map[string]interface{})["image"] != "example.invalid/dashboard:qa" {
			t.Fatal("image override was ignored")
		}
		dashboard := containers[0].(map[string]interface{})
		for _, probe := range []string{"startupProbe", "readinessProbe"} {
			if dashboardField(dashboard, probe, "httpGet", "path") != "/api/v1/info" ||
				dashboardField(dashboard, probe, "httpGet", "port") != 8080 {
				t.Fatalf("%s must check the public backend endpoint", probe)
			}
		}
	}
}

func TestServiceAccountOwnership(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		create bool
	}{
		{"standalone", nil, true},
		{"parent-owned", []string{"--set", "serviceAccount.create=false"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			docs := renderDashboard(t, "qa-ns", tc.args...)
			found := false
			for _, doc := range docs {
				if doc["kind"] == "ServiceAccount" {
					found = true
					if dashboardField(doc, "metadata", "name") != "qa-sa" {
						t.Fatal("service account name changed")
					}
				}
			}
			if found != tc.create {
				t.Fatalf("service account rendered = %t, want %t", found, tc.create)
			}
			deployment := dashboardObject(t, docs, "Deployment", "oceanbase-dashboard-qa")
			if dashboardField(deployment, "spec", "template", "spec", "serviceAccountName") != "qa-sa" {
				t.Fatal("deployment does not reference the expected service account")
			}
		})
	}
}

func TestRetainedAccountsAndRBACSurviveUpgrade(t *testing.T) {
	encode := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
	accounts := map[string]interface{}{
		"admin": encode("password-changed-in-ui <|:SEP:|> admin <|:SEP:|> 123 <|:SEP:|> owner"),
		"alice": encode("additional-account <|:SEP:|> Alice <|:SEP:|> 123 <|:SEP:|> analyst"),
	}
	policy := map[string]interface{}{
		"rbac_policy.csv": "p, admin, *, *, \"Super admin\"\ng, admin, admin\np, reader, obcluster/*, read, \"Reader\"\ng, alice, reader",
		"custom-key":      "keep-all-data",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			t.Errorf("unexpected mutation %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/version":
			fmt.Fprint(w, `{"gitVersion":"v1.30.0","major":"1","minor":"30"}`)
		case "/api":
			fmt.Fprint(w, `{"kind":"APIVersions","apiVersion":"v1","versions":["v1"]}`)
		case "/apis":
			fmt.Fprint(w, `{"kind":"APIGroupList","apiVersion":"v1","groups":[]}`)
		case "/api/v1":
			fmt.Fprint(w, `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"v1","resources":[{"name":"configmaps","kind":"ConfigMap","namespaced":true,"verbs":["get"]},{"name":"secrets","kind":"Secret","namespaced":true,"verbs":["get"]}]}`)
		case "/api/v1/namespaces/qa-ns/secrets/qa-user-credentials":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"apiVersion": "v1", "kind": "Secret", "metadata": map[string]string{"name": "qa-user-credentials", "namespace": "qa-ns"}, "data": accounts})
		case "/api/v1/namespaces/qa-ns/configmaps/qa-rbac-policy":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]string{"name": "qa-rbac-policy", "namespace": "qa-ns"}, "data": policy})
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`)
		}
	}))
	defer server.Close()
	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	config := fmt.Sprintf("apiVersion: v1\nkind: Config\nclusters:\n- name: test\n  cluster:\n    server: %s\ncontexts:\n- name: test\n  context:\n    cluster: test\n    user: test\ncurrent-context: test\nusers:\n- name: test\n  user: {}\n", server.URL)
	if err := os.WriteFile(kubeconfig, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	docs := renderDashboard(t, "qa-ns", "--set", "retainData=true", "--set-string", "adminPassword=must-not-overwrite-ui-password", "--dry-run=server", "--kubeconfig", kubeconfig)
	secret := dashboardObject(t, docs, "Secret", "qa-user-credentials")
	cm := dashboardObject(t, docs, "ConfigMap", "qa-rbac-policy")
	if !reflect.DeepEqual(secret["data"], accounts) {
		t.Fatal("upgrade must preserve changed passwords and additional users")
	}
	if !reflect.DeepEqual(cm["data"], policy) {
		t.Fatal("upgrade must preserve custom roles and all policy data")
	}
	for _, doc := range docs {
		switch doc["kind"] {
		case "Secret", "PersistentVolumeClaim":
			if dashboardField(doc, "metadata", "annotations", "helm.sh/resource-policy") != "keep" {
				t.Fatalf("%v is not retained", doc["kind"])
			}
		}
	}
	if dashboardField(cm, "metadata", "annotations", "helm.sh/resource-policy") != "keep" {
		t.Fatal("authorization state is not retained")
	}
}

func TestMonitoringSidecarsGateReadiness(t *testing.T) {
	deployment := dashboardObject(t, renderDashboard(t, "qa-ns"), "Deployment", "oceanbase-dashboard-qa")
	containers := dashboardField(deployment, "spec", "template", "spec", "containers").([]interface{})
	for _, item := range containers {
		container := item.(map[string]interface{})
		port := 0
		switch container["name"] {
		case "prometheus":
			port = 9090
		case "alertmanager":
			port = 9093
		default:
			continue
		}
		if dashboardField(container, "readinessProbe", "httpGet", "path") != "/-/ready" || dashboardField(container, "readinessProbe", "httpGet", "port") != port {
			t.Fatalf("%s must gate Pod readiness with its actual HTTP health endpoint", container["name"])
		}
		if dashboardField(container, "readinessProbe", "timeoutSeconds") != 3 {
			t.Fatal("sidecar probe must be bounded")
		}
	}
}

func TestDashboardSchedulingAndNodePort(t *testing.T) {
	docs := renderDashboard(t, "qa-ns", "--set-string", "nodeSelector.kubernetes\\.io/hostname=worker-a",
		"--set-string", "tolerations[0].key=dedicated,tolerations[0].operator=Equal,tolerations[0].value=lakehouse,tolerations[0].effect=NoSchedule",
		"--set", "service.type=NodePort,service.nodePort=32101")
	deployment := dashboardObject(t, docs, "Deployment", "oceanbase-dashboard-qa")
	if dashboardField(deployment, "spec", "template", "spec", "nodeSelector", "kubernetes.io/hostname") != "worker-a" {
		t.Fatal("Dashboard must follow the selected worker")
	}
	tolerations := dashboardField(deployment, "spec", "template", "spec", "tolerations").([]interface{})
	want := map[string]interface{}{"key": "dedicated", "operator": "Equal", "value": "lakehouse", "effect": "NoSchedule"}
	if len(tolerations) != 1 || !reflect.DeepEqual(tolerations[0], want) {
		t.Fatalf("unexpected tolerations: %v", tolerations)
	}
	service := dashboardObject(t, docs, "Service", "oceanbase-dashboard-qa")
	ports := dashboardField(service, "spec", "ports").([]interface{})
	if dashboardField(ports[0].(map[string]interface{}), "nodePort") != 32101 {
		t.Fatal("explicit NodePort was ignored")
	}
	for _, args := range [][]string{nil, {"--set", "service.type=ClusterIP,service.nodePort=32101"}} {
		docs := renderDashboard(t, "qa-ns", args...)
		service := dashboardObject(t, docs, "Service", "oceanbase-dashboard-qa")
		ports := dashboardField(service, "spec", "ports").([]interface{})
		if dashboardField(ports[0].(map[string]interface{}), "nodePort") != nil {
			t.Fatal("must omit nodePort when unset or not applicable")
		}
	}
}

func TestNotesDescribeActualCredentialOwnership(t *testing.T) {
	// Render the actual NOTES template inside an isolated ConfigMap so this test
	// remains offline (helm template normally suppresses NOTES from its output).
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "templates"), 0700); err != nil {
		t.Fatal(err)
	}
	notes, err := os.ReadFile("templates/NOTES.txt")
	if err != nil {
		t.Fatal(err)
	}
	values, err := os.ReadFile("values.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{
		"Chart.yaml":           []byte("apiVersion: v2\nname: notes-test\nversion: 0.1.0\n"),
		"values.yaml":          values,
		"notes.tpl":            notes,
		"templates/notes.yaml": []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: notes\ndata:\n  notes: {{ tpl (.Files.Get \"notes.tpl\") . | quote }}\n"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, initialize := range []string{"true", "false"} {
		out, err := exec.Command("helm", "template", "qa", dir, "-n", "qa-ns", "--set", "initCredentials="+initialize+",service.port=8888",
			"--set-string", "adminPassword=NOTES_PASSWORD_SENTINEL,userCredentials=provided-users,userNamespace=identity-ns").CombinedOutput()
		if err != nil {
			t.Fatalf("notes render: %v %s", err, out)
		}
		var document map[string]interface{}
		if err := yaml.Unmarshal(out, &document); err != nil {
			t.Fatal(err)
		}
		text := dashboardField(document, "data", "notes").(string)
		if strings.Contains(text, "NOTES_PASSWORD_SENTINEL") || strings.Contains(text, "--address 0.0.0.0") {
			t.Fatal("notes exposed credentials or public forwarding")
		}
		if !strings.Contains(text, "18081:8888") {
			t.Fatal("access command ignored configured service port")
		}
		if initialize == "true" {
			if !strings.Contains(text, "standalone chart default") || !strings.Contains(text, "use the current password") {
				t.Fatal("standalone password guidance incomplete")
			}
		} else {
			if strings.Contains(text, "standalone chart default") || !strings.Contains(text, "provided-users") || !strings.Contains(text, "identity-ns") || !strings.Contains(text, "qa-dashboard-bootstrap") {
				t.Fatal("parent/external credential guidance incorrect")
			}
		}
	}
}
