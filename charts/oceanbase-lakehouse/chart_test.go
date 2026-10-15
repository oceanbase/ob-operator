package lakehouse_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type manifest map[string]interface{}

func render(t *testing.T, values map[string]interface{}) ([]manifest, string, error) {
	t.Helper()
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	file := filepath.Join(t.TempDir(), "values.json")
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(helm, "template", "qa", ".", "-n", "qa-ns", "-f", file).CombinedOutput()
	if err != nil {
		return nil, string(output), err
	}
	var docs []manifest
	d := yaml.NewDecoder(bytes.NewReader(output))
	for {
		var doc manifest
		if err := d.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("invalid YAML: %v", err)
		}
		if len(doc) != 0 {
			docs = append(docs, doc)
		}
	}
	return docs, string(output), nil
}

func field(m manifest, keys ...string) interface{} {
	var value interface{} = map[string]interface{}(m)
	for _, key := range keys {
		switch object := value.(type) {
		case map[string]interface{}:
			value = object[key]
		case manifest:
			value = object[key]
		default:
			return nil
		}
	}
	return value
}

func find(t *testing.T, docs []manifest, kind, name string) manifest {
	t.Helper()
	for _, doc := range docs {
		if doc["kind"] == kind && field(doc, "metadata", "name") == name {
			return doc
		}
	}
	t.Fatalf("missing %s/%s", kind, name)
	return nil
}

func TestStagesAndRetention(t *testing.T) {
	for _, stage := range []string{"storage", "logservice", "cluster"} {
		t.Run(stage, func(t *testing.T) {
			docs, output, err := render(t, map[string]interface{}{"stage": stage})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			counts := map[string]int{}
			for _, doc := range docs {
				kind := doc["kind"].(string)
				counts[kind]++
				if field(doc, "metadata", "namespace") != "qa-ns" {
					t.Fatalf("wrong namespace: %v", doc["metadata"])
				}
				switch kind {
				case "Secret", "PersistentVolumeClaim", "StatefulSet", "Service", "ServiceAccount", "OBCluster", "OBLogServiceCluster":
					if field(doc, "metadata", "annotations", "helm.sh/resource-policy") != "keep" {
						t.Fatalf("%s lacks retention", kind)
					}
					if field(doc, "metadata", "ownerReferences") != nil {
						t.Fatalf("%s must not have a garbage-collection owner", kind)
					}
				}
			}
			if counts["StatefulSet"] != 1 || counts["PersistentVolumeClaim"] != 1 || counts["Secret"] != 3 || counts["Job"] != 1 {
				t.Fatalf("unexpected infrastructure: %v", counts)
			}
			wantLS, wantOB := 0, 0
			if stage != "storage" {
				wantLS = 1
			}
			if stage == "cluster" {
				wantOB = 1
				ob := find(t, docs, "OBCluster", "qa-ob")
				if field(ob, "spec", "deploymentMode") != "shared_storage" || field(ob, "spec", "logServiceRef", "name") != "qa-ls" {
					t.Fatal("SS/reference wiring is broken")
				}
				if field(ob, "spec", "sharedStorageInfo", "secretRef", "name") != "qa-storage" {
					t.Fatal("Observer must use business credentials")
				}
			}
			if counts["OBLogServiceCluster"] != wantLS || counts["OBCluster"] != wantOB {
				t.Fatalf("incorrect stage ordering: %v", counts)
			}
			sts := find(t, docs, "StatefulSet", "qa-minio")
			if field(sts, "spec", "template", "spec", "automountServiceAccountToken") != false {
				t.Fatal("MinIO must not receive Kubernetes credentials")
			}
			if field(sts, "spec", "updateStrategy", "type") != "OnDelete" {
				t.Fatal("MinIO upgrades must be explicitly planned")
			}
			if field(find(t, docs, "Service", "qa-minio"), "spec", "type") != "ClusterIP" {
				t.Fatal("MinIO must not open public ports")
			}
		})
	}
}

func TestExternalMinIOAndUserOwnedResources(t *testing.T) {
	values := map[string]interface{}{
		"stage":         "cluster",
		"objectStorage": map[string]interface{}{"mode": "external", "endpoint": "https://minio.example.com:9000", "existingSecret": "external-s3"},
		"observer":      map[string]interface{}{"rootExistingSecret": "db-root"},
	}
	docs, output, err := render(t, values)
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	for _, doc := range docs {
		switch doc["kind"] {
		case "StatefulSet", "PersistentVolumeClaim", "Service", "Secret":
			t.Fatalf("external mode must not own storage or supplied Secrets: %v", doc["kind"])
		}
	}
	for _, kindName := range [][3]string{{"OBCluster", "qa-ob", "sharedStorageInfo"}, {"OBLogServiceCluster", "qa-ls", "objectStoreUrl"}} {
		cr := find(t, docs, kindName[0], kindName[1])
		if field(cr, "spec", kindName[2], "secretRef", "name") != "external-s3" {
			t.Fatal("external credential reference is broken")
		}
		url := field(cr, "spec", kindName[2], "bucketURL").(string)
		if !strings.Contains(url, "host=https://minio.example.com:9000&s3_region=us-east-1") {
			t.Fatalf("bad external URL %s", url)
		}
	}
}

func TestExistingClaimAndRootSecretAreNotOwned(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{"minio": map[string]interface{}{"existingClaim": "storage-pvc", "rootExistingSecret": "storage-root"}})
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	for _, doc := range docs {
		if doc["kind"] == "PersistentVolumeClaim" || field(doc, "metadata", "name") == "storage-root" {
			t.Fatal("must not adopt user-owned storage or credentials")
		}
	}
	if !strings.Contains(output, "claimName: storage-pvc") || !strings.Contains(output, "name: storage-root") {
		t.Fatal("missing references to existing storage")
	}
}

func TestRejectInvalidConfiguration(t *testing.T) {
	cases := map[string]map[string]interface{}{
		"external-no-secret":   {"objectStorage": map[string]interface{}{"mode": "external", "endpoint": "http://minio:9000"}},
		"external-no-endpoint": {"objectStorage": map[string]interface{}{"mode": "external", "existingSecret": "creds"}},
		"bad-mode":             {"objectStorage": map[string]interface{}{"mode": "typo"}},
		"url-query":            {"objectStorage": map[string]interface{}{"endpoint": "http://minio:9000?access_key=SECRET"}},
		"url-credentials":      {"objectStorage": map[string]interface{}{"endpoint": "http://user:SECRET@minio:9000"}},
		"url-shell":            {"objectStorage": map[string]interface{}{"endpoint": "http://minio/$(id)"}},
		"bad-bucket":           {"objectStorage": map[string]interface{}{"dataBucket": "bucket\"bad"}},
		"same-bucket":          {"objectStorage": map[string]interface{}{"dataBucket": "same", "logBucket": "same"}},
		"inline-creds":         {"objectStorage": map[string]interface{}{"access_key": "SECRET"}},
		"latest":               {"minio": map[string]interface{}{"image": "minio/minio:latest"}},
		"untagged":             {"minio": map[string]interface{}{"image": "minio/minio"}},
		"empty-topology":       {"logService": map[string]interface{}{"topology": []interface{}{}}},
		"zero-replicas":        {"logService": map[string]interface{}{"topology": []interface{}{map[string]interface{}{"zone": "zone1", "replica": 0}}}},
		"duplicate-zones":      {"logService": map[string]interface{}{"topology": []interface{}{map[string]interface{}{"zone": "z1", "replica": 1}, map[string]interface{}{"zone": "z1", "replica": 1}}}},
		"short-deadline":       {"initialization": map[string]interface{}{"deadlineSeconds": 0}},
		"disable-persistence":  {"minio": map[string]interface{}{"persistence": false}},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := render(t, values); err == nil {
				t.Fatal("unsafe configuration was accepted")
			}
		})
	}
}

func TestStorageInitializationScript(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	script := field(find(t, docs, "ConfigMap", "qa-storage-init"), "data", "initialize.sh").(string)
	for _, want := range []string{"set -eu", "umask 077", "mc ready", "--ignore-existing", "if [ \"$MODE\" = bundled ]", "s3:ListBucketMultipartUploads"} {
		if !strings.Contains(script, want) {
			t.Fatalf("initialization lacks %q", want)
		}
	}
	for _, forbidden := range []string{"set -x", "--insecure", "mc rb", "mc rm", "s3:*", "s3:ListAllMyBuckets", "eval "} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("unsafe initialization: %q", forbidden)
		}
	}
	cmd := exec.Command("sh", "-n")
	cmd.Stdin = strings.NewReader(script)
	if result, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("initialization script syntax: %v %s", err, result)
	}
}

func TestLogServiceBootstrapRequiresThreeNodes(t *testing.T) {
	for _, replicas := range []int{1, 2, 3, 4} {
		values := map[string]interface{}{"logService": map[string]interface{}{"topology": []interface{}{map[string]interface{}{"zone": "zone1", "replica": replicas}}}}
		_, output, err := render(t, values)
		if replicas == 3 && err != nil {
			t.Fatalf("three-node bootstrap rejected: %v %s", err, output)
		}
		if replicas != 3 && (err == nil || !strings.Contains(output, "bootstrap requires exactly 3 nodes")) {
			t.Fatalf("invalid initial node count %d accepted: %v %s", replicas, err, output)
		}
	}
}

func TestInstallScriptArgumentSafety(t *testing.T) {
	for _, args := range [][]string{
		{}, {"UPPER", "qa"}, {"qa", "bad/namespace"}, {"qa", "qa", "--atomic"}, {"qa", "qa", "--force"}, {"qa", "qa", "--cleanup-on-fail"}, {"qa", "qa", "-f"},
	} {
		cmd := exec.Command("bash", append([]string{"install.sh"}, args...)...)
		if result, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("unsafe arguments accepted: %v %s", args, result)
		}
	}
}
