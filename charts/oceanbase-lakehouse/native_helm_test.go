package lakehouse_test

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeHelmDefaultDeclaresCompleteGraph(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatal(err, output)
	}
	ls := find(t, docs, "OBLogServiceCluster", "qa-ls")
	ob := find(t, docs, "OBCluster", "qa-ob")
	if field(ls, "metadata", "annotations", "oceanbase.oceanbase.com/storage-init-job") != "qa-storage-init-1" {
		t.Fatal("LS bootstrap is not gated on storage initialization")
	}
	for _, cr := range []manifest{ls, ob} {
		if field(cr, "metadata", "annotations", "oceanbase.oceanbase.com/resolve-storage-service") != "true" {
			t.Fatal("missing opt-in runtime service resolution")
		}
	}
	for _, tc := range []struct {
		cr  manifest
		key string
	}{{ls, "objectStoreUrl"}, {ob, "sharedStorageInfo"}} {
		if !strings.Contains(field(tc.cr, "spec", tc.key, "bucketURL").(string), "host=http://qa-minio.qa-ns.svc:9000") {
			t.Fatal("first install must declare a stable Service name before ClusterIP exists")
		}
	}
	init := find(t, docs, "Job", "qa-storage-init-1")
	if field(init, "spec", "ttlSecondsAfterFinished") != nil {
		t.Fatal("Operator gate must survive delayed bootstrap and restarts")
	}
	if field(init, "spec", "backoffLimit") != 3 {
		t.Fatal("initialization must retry transient errors")
	}
	find(t, docs, "Deployment", "oceanbase-dashboard-qa")
}
func TestNativeHelmReadinessIsReadOnlyAndRunsAfterInstallAndUpgrade(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatal(err, output)
	}
	job := find(t, docs, "Job", "qa-readiness")
	if field(job, "metadata", "annotations", "helm.sh/hook") != "post-install,post-upgrade" {
		t.Fatal("readiness must run for install and upgrade")
	}
	if field(job, "metadata", "annotations", "helm.sh/hook-delete-policy") != "before-hook-creation,hook-succeeded" {
		t.Fatal("failed readiness evidence must survive")
	}
	if field(job, "spec", "template", "spec", "serviceAccountName") != "qa-readiness" {
		t.Fatal("readiness must not use a privileged application identity")
	}
	role := find(t, docs, "Role", "qa-readiness")
	for _, rule := range field(role, "rules").([]interface{}) {
		r := rule.(manifest)
		for _, verb := range r["verbs"].([]interface{}) {
			if verb != "get" && verb != "list" {
				t.Fatalf("readiness may not write resources: %v", r)
			}
		}
		for _, resource := range r["resources"].([]interface{}) {
			if resource == "secrets" || resource == "*" {
				t.Fatal("readiness must not read credentials")
			}
		}
	}
	if strings.Contains(output, "command: [helm") || strings.Contains(output, "kubectl apply") || strings.Contains(output, "kubectl patch") {
		t.Fatal("Jobs must not mutate Helm-owned resources")
	}
}
func TestNativeHelmUpgradeKeepsOriginalCRURLs(t *testing.T) {
	api, config := retainedCredentials(t)
	lsURL := "s3://qa-logs?host=http://10.43.0.77:9000&s3_region=us-east-1"
	obURL := "s3://qa-data?host=http://10.43.0.77:9000&s3_region=us-east-1"
	api.crs = map[string]interface{}{
		"/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa-ns/oblogserviceclusters/qa-ls": map[string]interface{}{"apiVersion": "oceanbase.oceanbase.com/v1alpha1", "kind": "OBLogServiceCluster", "metadata": map[string]string{"name": "qa-ls", "namespace": "qa-ns"}, "spec": map[string]interface{}{"objectStoreUrl": map[string]string{"bucketURL": lsURL}}},
		"/apis/oceanbase.oceanbase.com/v1alpha1/namespaces/qa-ns/obclusters/qa-ob":           map[string]interface{}{"apiVersion": "oceanbase.oceanbase.com/v1alpha1", "kind": "OBCluster", "metadata": map[string]string{"name": "qa-ob", "namespace": "qa-ns"}, "spec": map[string]interface{}{"sharedStorageInfo": map[string]string{"bucketURL": obURL}}},
	}
	output, err := renderRetained(t, config, map[string]interface{}{})
	if err != nil {
		t.Fatal(err, output)
	}
	if !strings.Contains(output, lsURL) || !strings.Contains(output, obURL) {
		t.Fatal("upgrading a staged release must preserve immutable storage URLs")
	}
}
func TestNativeHelmNewServiceAllocationDoesNotChangeCRURL(t *testing.T) {
	api, config := retainedCredentials(t)
	for _, ip := range []string{"", "10.43.0.77"} {
		api.serviceIP = ip
		output, err := renderRetained(t, config, map[string]interface{}{})
		if err != nil {
			t.Fatal(err, output)
		}
		if strings.Count(output, "?host=http://qa-minio.qa-ns.svc:9000&s3_region=us-east-1") != 2 {
			t.Fatal("native URL must remain declarative before/after Service allocation")
		}
	}
}
func TestHelmPackageContainsNoInstallScriptOrTests(t *testing.T) {
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	dir := t.TempDir()
	out, err := exec.Command(helm, "package", ".", "--destination", dir).CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	archives, err := filepath.Glob(filepath.Join(dir, "*.tgz"))
	if err != nil || len(archives) != 1 {
		t.Fatal(err, archives)
	}
	f, err := os.Open(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	foundReadiness := false
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(h.Name, "/install.sh") || strings.HasSuffix(h.Name, "_test.go") {
			t.Fatalf("unwanted file shipped: %s", h.Name)
		}
		if strings.HasSuffix(h.Name, "/templates/readiness.yaml") {
			foundReadiness = true
		}
	}
	if !foundReadiness {
		t.Fatal("native readiness hook missing from packaged Chart")
	}
}

func TestNativeExternalStorageKeepsExistingEndpointContract(t *testing.T) {
	docs, output, err := render(t, externalValues("https://s3.example.com"))
	if err != nil {
		t.Fatal(err, output)
	}
	for _, entry := range []struct{ kind, name string }{{"OBLogServiceCluster", "qa-ls"}, {"OBCluster", "qa-ob"}} {
		cr := find(t, docs, entry.kind, entry.name)
		if field(cr, "metadata", "annotations", "oceanbase.oceanbase.com/resolve-storage-service") != nil {
			t.Fatal("external endpoints must retain their existing guarded Helm resolution contract")
		}
	}
	if field(find(t, docs, "OBLogServiceCluster", "qa-ls"), "metadata", "annotations", "oceanbase.oceanbase.com/storage-init-job") != "qa-storage-init-1" {
		t.Fatal("external storage still requires credential/bucket preflight")
	}
}
