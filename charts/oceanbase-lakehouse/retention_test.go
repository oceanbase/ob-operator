package lakehouse_test

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
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// retentionAPI supplies only the discovery and read-only lookup responses Helm
// needs. These tests never connect to the developer's current Kubernetes context.
type retentionAPI struct {
	mu               sync.RWMutex
	identity         string
	secrets          map[string]map[string]string
	configmaps       map[string]map[string]string
	serviceIP        string
	noGuard          bool
	externalService  map[string]interface{}
	externalEndpoint string
	crs              map[string]interface{}
}

func (api *retentionAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "test API permits no mutations", http.StatusMethodNotAllowed)
		return
	}
	api.mu.RLock()
	defer api.mu.RUnlock()
	switch r.URL.Path {
	case "/version":
		fmt.Fprint(w, `{"gitVersion":"v1.30.0","major":"1","minor":"30"}`)
	case "/api":
		fmt.Fprint(w, `{"kind":"APIVersions","apiVersion":"v1","versions":["v1"]}`)
	case "/apis":
		fmt.Fprint(w, `{"kind":"APIGroupList","apiVersion":"v1","groups":[{"name":"oceanbase.oceanbase.com","versions":[{"groupVersion":"oceanbase.oceanbase.com/v1alpha1","version":"v1alpha1"}],"preferredVersion":{"groupVersion":"oceanbase.oceanbase.com/v1alpha1","version":"v1alpha1"}}]}`)
	case "/api/v1":
		fmt.Fprint(w, `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"v1","resources":[{"name":"configmaps","kind":"ConfigMap","namespaced":true,"verbs":["get","list"]},{"name":"secrets","kind":"Secret","namespaced":true,"verbs":["get","list"]},{"name":"services","kind":"Service","namespaced":true,"verbs":["get","list"]}]}`)
	case "/apis/oceanbase.oceanbase.com/v1alpha1":
		fmt.Fprint(w, `{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"oceanbase.oceanbase.com/v1alpha1","resources":[{"name":"obclusters","kind":"OBCluster","namespaced":true,"verbs":["get","list"]},{"name":"oblogserviceclusters","kind":"OBLogServiceCluster","namespaced":true,"verbs":["get","list"]}]}`)
	case "/api/v1/namespaces/qa-ns/configmaps/qa-storage-identity":
		if api.noGuard {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`)
			return
		}
		data := map[string]string{"identity": api.identity}
		if api.externalEndpoint != "" {
			data["externalDatabaseEndpoint"] = api.externalEndpoint
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"apiVersion": "v1", "kind": "ConfigMap",
			"metadata": map[string]string{"name": "qa-storage-identity", "namespace": "qa-ns"},
			"data":     data,
		})
	default:
		if cr, ok := api.crs[r.URL.Path]; ok {
			_ = json.NewEncoder(w).Encode(cr)
			return
		}
		const configmapPrefix = "/api/v1/namespaces/qa-ns/configmaps/"
		if strings.HasPrefix(r.URL.Path, configmapPrefix) {
			name := strings.TrimPrefix(r.URL.Path, configmapPrefix)
			if data, ok := api.configmaps[name]; ok {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"apiVersion": "v1", "kind": "ConfigMap",
					"metadata": map[string]string{"name": name, "namespace": "qa-ns"},
					"data":     data,
				})
				return
			}
		}
		if r.URL.Path == "/api/v1/namespaces/storage/services/minio" && api.externalService != nil {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"apiVersion": "v1", "kind": "Service", "metadata": map[string]string{"name": "minio", "namespace": "storage"}, "spec": api.externalService})
			return
		}
		if r.URL.Path == "/api/v1/namespaces/qa-ns/services/qa-minio" && api.serviceIP != "" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"apiVersion": "v1", "kind": "Service",
				"metadata": map[string]string{"name": "qa-minio", "namespace": "qa-ns"},
				"spec":     map[string]string{"clusterIP": api.serviceIP},
			})
			return
		}
		const secretPrefix = "/api/v1/namespaces/qa-ns/secrets/"
		if strings.HasPrefix(r.URL.Path, secretPrefix) {
			name := strings.TrimPrefix(r.URL.Path, secretPrefix)
			if data, ok := api.secrets[name]; ok {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"apiVersion": "v1", "kind": "Secret",
					"metadata": map[string]string{"name": name, "namespace": "qa-ns"},
					"data":     data,
				})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`)
	}
}

// LogService 1.3.0 uses virtual-host addressing for hostnames, which cannot
// resolve bucket.service.namespace.svc in ordinary Kubernetes DNS. The installer
// creates storage first; subsequent stages must use its stable Service IP.
func TestBundledDatabaseUsesRetainedServiceIP(t *testing.T) {
	for _, tc := range []struct{ ip, host string }{{"10.43.0.99", "10.43.0.99"}, {"fd00::99", "[fd00::99]"}} {
		t.Run(tc.ip, func(t *testing.T) {
			api, kubeconfig := retainedCredentials(t)
			api.mu.Lock()
			api.serviceIP = tc.ip
			api.mu.Unlock()
			output, err := renderRetained(t, kubeconfig, map[string]interface{}{"native": map[string]interface{}{"enabled": false}})
			if err != nil {
				t.Fatalf("retained Service render: %v\n%s", err, output)
			}
			if strings.Count(output, "?host=http://"+tc.host+":9000&s3_region=us-east-1") != 2 {
				t.Fatal("both database CRs must use stable Service IP, not wildcard bucket DNS")
			}
			if !strings.Contains(output, "http://qa-minio.qa-ns.svc:9000") {
				t.Fatal("initialization should retain its DNS endpoint")
			}
		})
	}
}

func retainedCredentials(t *testing.T) (*retentionAPI, string) {
	t.Helper()
	docs, output, err := render(t, map[string]interface{}{"stage": "cluster"})
	if err != nil {
		t.Fatalf("render identity fixture: %v\n%s", err, output)
	}
	identity, ok := field(find(t, docs, "ConfigMap", "qa-storage-identity"), "data", "identity").(string)
	if !ok || identity == "" {
		t.Fatal("missing storage identity")
	}
	encoded := func(value string) string { return base64.StdEncoding.EncodeToString([]byte(value)) }
	api := &retentionAPI{
		identity: identity,
		secrets: map[string]map[string]string{
			"qa-db-root":    {"password": encoded("retained-db-password")},
			"qa-minio-root": {"rootUser": encoded("retained-root-id"), "rootPassword": encoded("retained-root-password")},
			"qa-storage":    {"access_id": encoded("retained-business-id"), "access_key": encoded("retained-business-password")},
		},
	}
	server := httptest.NewServer(api)
	t.Cleanup(server.Close)
	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	config := fmt.Sprintf("apiVersion: v1\nkind: Config\nclusters:\n- name: retention-test\n  cluster:\n    server: %s\ncontexts:\n- name: retention-test\n  context:\n    cluster: retention-test\n    user: retention-test\ncurrent-context: retention-test\nusers:\n- name: retention-test\n  user: {}\n", server.URL)
	if err := os.WriteFile(kubeconfig, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	return api, kubeconfig
}

func renderRetained(t *testing.T, kubeconfig string, values map[string]interface{}) (string, error) {
	t.Helper()
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "values.json")
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(helm, "template", "qa", ".", "-n", "qa-ns", "-f", file,
		"--set", "stage=cluster", "--dry-run=server", "--kubeconfig", kubeconfig).CombinedOutput()
	return string(output), err
}

func TestRetainedGeneratedCredentialsRemainStable(t *testing.T) {
	api, kubeconfig := retainedCredentials(t)
	for attempt := 0; attempt < 2; attempt++ {
		output, err := renderRetained(t, kubeconfig, map[string]interface{}{"dashboard": map[string]interface{}{"enabled": false}})
		if err != nil {
			t.Fatalf("render retained resources: %v\n%s", err, output)
		}
		decoder := yaml.NewDecoder(bytes.NewBufferString(output))
		seen := map[string]bool{}
		for {
			var doc manifest
			if err := decoder.Decode(&doc); err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("invalid rendered YAML: %v", err)
			}
			if doc["kind"] != "Secret" {
				continue
			}
			name := field(doc, "metadata", "name").(string)
			api.mu.RLock()
			data, ok := api.secrets[name]
			if !ok {
				api.mu.RUnlock()
				t.Fatalf("unexpected generated Secret %s", name)
			}
			for key, original := range data {
				if field(doc, "data", key) != original {
					api.mu.RUnlock()
					t.Fatalf("retained Secret %s key %s changed on render %d", name, key, attempt)
				}
			}
			api.mu.RUnlock()
			seen[name] = true
		}
		if len(seen) != 3 {
			t.Fatalf("expected all three retained Secrets, got %v", seen)
		}
	}
}

func TestRetainedStorageIdentityRejectsDrift(t *testing.T) {
	_, kubeconfig := retainedCredentials(t)
	cases := map[string]map[string]interface{}{
		"data-bucket":   {"objectStorage": map[string]interface{}{"dataBucket": "other-data"}},
		"log-bucket":    {"objectStorage": map[string]interface{}{"logBucket": "other-logs"}},
		"region":        {"objectStorage": map[string]interface{}{"region": "eu-west-1"}},
		"storage-mode":  {"objectStorage": map[string]interface{}{"mode": "external", "endpoint": "https://storage.example.com", "existingSecret": "external-storage"}},
		"minio-claim":   {"minio": map[string]interface{}{"existingClaim": "other-pvc"}},
		"minio-size":    {"minio": map[string]interface{}{"size": "200Gi"}},
		"minio-secret":  {"minio": map[string]interface{}{"rootExistingSecret": "other-minio-root"}},
		"storage-class": {"storageClass": "other-storage-class"},
		"database-root": {"observer": map[string]interface{}{"rootExistingSecret": "other-db-root"}},
		"database-id":   {"observer": map[string]interface{}{"clusterId": 991}},
		"logservice-id": {"logService": map[string]interface{}{"clusterId": 992}},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			output, err := renderRetained(t, kubeconfig, values)
			if err == nil || !strings.Contains(output, "storage identity is immutable") {
				t.Fatalf("identity drift was not rejected: %v\n%s", err, output)
			}
		})
	}
}

func TestMissingRetainedSecretFailsClosed(t *testing.T) {
	for _, name := range []string{"qa-db-root", "qa-minio-root", "qa-storage"} {
		t.Run(name, func(t *testing.T) {
			api, kubeconfig := retainedCredentials(t)
			api.mu.Lock()
			delete(api.secrets, name)
			api.mu.Unlock()
			output, err := renderRetained(t, kubeconfig, map[string]interface{}{})
			if err == nil || !strings.Contains(output, "retained credential Secret "+name+" is missing") ||
				!strings.Contains(output, "restore it instead of generating new credentials") {
				t.Fatalf("missing retained Secret did not fail closed: %v\n%s", err, output)
			}
		})
	}
}
