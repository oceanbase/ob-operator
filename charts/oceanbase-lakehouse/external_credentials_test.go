package lakehouse_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
)

// External storage has the same independent credential references as the two
// Operator CRs. A shared bucket must not be rejected by a MinIO-only policy.
func TestExternalIndependentCredentials(t *testing.T) {
	for _, sameBucket := range []bool{false, true} {
		storage := map[string]interface{}{
			"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "region": "cn-hangzhou",
			"logBucket": "customer-logs", "dataBucket": "customer-data",
			"logExistingSecret": "log-credentials", "dataExistingSecret": "data-credentials",
		}
		if sameBucket {
			storage["dataBucket"] = "customer-logs"
		}
		docs, output, err := render(t, map[string]interface{}{"stage": "cluster", "objectStorage": storage, "dashboard": map[string]interface{}{"enabled": false}})
		if err != nil {
			t.Fatalf("same bucket=%v: %v\n%s", sameBucket, err, output)
		}
		for _, want := range []struct{ kind, name, property, secret string }{
			{"OBLogServiceCluster", "qa-ls", "objectStoreUrl", "log-credentials"},
			{"OBCluster", "qa-ob", "sharedStorageInfo", "data-credentials"},
		} {
			if field(find(t, docs, want.kind, want.name), "spec", want.property, "secretRef", "name") != want.secret {
				t.Fatalf("%s credential reference is not independent", want.kind)
			}
		}
		var job batchv1.Job
		decodeScheduledObject(t, find(t, docs, "Job", "qa-storage-init-1"), &job)
		refs := map[string]string{}
		for _, env := range job.Spec.Template.Spec.Containers[0].Env {
			if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
				refs[env.Name] = env.ValueFrom.SecretKeyRef.Name + "/" + env.ValueFrom.SecretKeyRef.Key
			}
		}
		for name, want := range map[string]string{"ACCESS_ID": "log-credentials/access_id", "ACCESS_KEY": "log-credentials/access_key", "DATA_ACCESS_ID": "data-credentials/access_id", "DATA_ACCESS_KEY": "data-credentials/access_key"} {
			if refs[name] != want {
				t.Errorf("%s = %s, want %s", name, refs[name], want)
			}
		}
	}
}

func TestExternalSharedBucketWithLegacyCredentials(t *testing.T) {
	storage := map[string]interface{}{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "existingSecret": "shared-credentials", "logBucket": "customer-bucket", "dataBucket": "customer-bucket"}
	if _, output, err := render(t, map[string]interface{}{"stage": "cluster", "objectStorage": storage}); err != nil {
		t.Fatalf("external shared bucket was rejected: %v\n%s", err, output)
	}
}

func TestExternalCredentialFallbackAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, fallback, log, data, wantLog, wantData string
	}{
		{"legacy", "shared", "", "", "shared", "shared"},
		{"log-only-override", "shared", "logs", "", "logs", "shared"},
		{"data-only-override", "shared", "", "data", "shared", "data"},
		{"independent", "", "logs", "data", "logs", "data"},
		{"explicit-shared", "", "shared", "shared", "shared", "shared"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage := map[string]interface{}{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "existingSecret": tc.fallback, "logExistingSecret": tc.log, "dataExistingSecret": tc.data}
			docs, output, err := render(t, map[string]interface{}{"stage": "cluster", "objectStorage": storage, "observer": map[string]interface{}{"rootExistingSecret": "db-root"}, "dashboard": map[string]interface{}{"enabled": false}})
			if err != nil {
				t.Fatal(err, output)
			}
			if field(find(t, docs, "OBLogServiceCluster", "qa-ls"), "spec", "objectStoreUrl", "secretRef", "name") != tc.wantLog ||
				field(find(t, docs, "OBCluster", "qa-ob"), "spec", "sharedStorageInfo", "secretRef", "name") != tc.wantData {
				t.Fatal("incorrect credential fallback")
			}
			for _, doc := range docs {
				if doc["kind"] == "Secret" || doc["kind"] == "PersistentVolumeClaim" || doc["kind"] == "StatefulSet" {
					t.Fatalf("external storage must not own supplied infrastructure: %s", doc["kind"])
				}
			}
		})
	}
	for _, storage := range []map[string]interface{}{
		{"mode": "bundled", "logExistingSecret": "logs"},
		{"mode": "bundled", "dataExistingSecret": "data"},
		{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "logExistingSecret": "logs"},
		{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "dataExistingSecret": "data"},
		{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "logExistingSecret": "../bad", "dataExistingSecret": "data"},
	} {
		if _, _, err := render(t, map[string]interface{}{"objectStorage": storage}); err == nil {
			t.Fatal("invalid or incomplete per-side credentials accepted")
		}
	}
}

func TestExternalCredentialIdentityRetention(t *testing.T) {
	api, config := retainedCredentials(t)
	api.noGuard = true
	storage := map[string]interface{}{"mode": "external", "endpoint": "https://oss-cn-hangzhou.aliyuncs.com", "logBucket": "same-bucket", "dataBucket": "same-bucket", "logExistingSecret": "log-key", "dataExistingSecret": "data-key"}
	values := map[string]interface{}{"objectStorage": storage}
	output, err := renderRetained(t, config, values)
	if err != nil {
		t.Fatal(err, output)
	}
	for _, doc := range releaseDocuments(t, []byte(output)) {
		if doc["kind"] == "ConfigMap" && field(doc, "metadata", "name") == "qa-storage-identity" {
			api.identity = field(doc, "data", "identity").(string)
			api.externalEndpoint = field(doc, "data", "externalDatabaseEndpoint").(string)
		}
	}
	api.noGuard = false
	if output, err = renderRetained(t, config, values); err != nil {
		t.Fatal("same references failed to resume", err, output)
	}
	for _, key := range []string{"logExistingSecret", "dataExistingSecret", "logBucket", "dataBucket"} {
		old := storage[key]
		storage[key] = "changed"
		output, err = renderRetained(t, config, values)
		storage[key] = old
		if err == nil || !strings.Contains(output, "storage identity is immutable") {
			t.Fatalf("retained %s could redirect data: %v %s", key, err, output)
		}
	}
}

func TestExternalPreflightUsesEachIdentityWithoutMinIOAdministration(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatal(err, output)
	}
	script := field(find(t, docs, "ConfigMap", "qa-storage-init"), "data", "initialize.sh").(string)
	dir := t.TempDir()
	executable(t, dir, "seq", "#!/bin/sh\necho 1\n")
	executable(t, dir, "sleep", "#!/bin/sh\nexit 0\n")
	executable(t, dir, "mc", `#!/bin/sh
set -eu
case "$1" in
  alias)
    [ "$2" = set ] && [ "$3" = storage ] || exit 11
    case "$5:$6" in log-id:log-secret|data-id:data-secret) ;; *) exit 12 ;; esac
    printf '%s' "$5" > "$PREFLIGHT_STATE"
    ;;
  stat)
    identity=$(cat "$PREFLIGHT_STATE")
    case "$identity:$2" in log-id:storage/logs|data-id:storage/data) ;; *) exit 13 ;; esac
    printf '%s\n' "$identity:$2" >> "$PREFLIGHT_LOG"
    ;;
  *) echo 'No MinIO admin/readiness/global-list operation allowed' >&2; exit 14 ;;
esac
`)
	cmd := exec.Command("sh")
	cmd.Stdin = strings.NewReader(strings.ReplaceAll(script, "/work", dir))
	cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"MODE=external", "CREATE_BUCKETS=false", "ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com", "MC_PATH=off",
		"ACCESS_ID=log-id", "ACCESS_KEY=log-secret", "DATA_ACCESS_ID=data-id", "DATA_ACCESS_KEY=data-secret",
		"LOG_BUCKET=logs", "DATA_BUCKET=data", "PREFLIGHT_STATE="+filepath.Join(dir, "state"), "PREFLIGHT_LOG="+filepath.Join(dir, "calls"))
	result, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("external preflight needs wrong identity or MinIO-specific API: %v %s", err, result)
	}
	calls, err := os.ReadFile(filepath.Join(dir, "calls"))
	if err != nil || string(calls) != "log-id:storage/logs\ndata-id:storage/data\n" {
		t.Fatalf("per-identity bucket checks missing: %v %s", err, calls)
	}
	if strings.Contains(string(result), "log-secret") || strings.Contains(string(result), "data-secret") {
		t.Fatal("preflight leaked credentials")
	}
}
