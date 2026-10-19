package lakehouse_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func externalValues(endpoint string) map[string]interface{} {
	return map[string]interface{}{"objectStorage": map[string]interface{}{"mode": "external", "endpoint": endpoint, "existingSecret": "external-s3"}}
}

func TestExternalServiceEndpointMatchesDatabaseAndPreflight(t *testing.T) {
	for _, host := range []string{"minio.storage.svc", "minio.storage.svc.cluster.local"} {
		api, config := retainedCredentials(t)
		api.noGuard = true
		api.externalService = map[string]interface{}{"clusterIP": "10.43.0.88", "ports": []interface{}{map[string]interface{}{"port": 9000, "protocol": "TCP"}}}
		output, err := renderRetained(t, config, externalValues("http://"+host+":9000"))
		if err != nil {
			t.Fatalf("%v\n%s", err, output)
		}
		if strings.Count(output, "?host=http://10.43.0.88:9000&s3_region=") != 2 {
			t.Fatal("database CRs did not use Service IP")
		}
		if !strings.Contains(output, `name: ENDPOINT, value: "http://10.43.0.88:9000"`) {
			t.Fatal("preflight uses a different endpoint")
		}
		if !strings.Contains(output, `externalDatabaseEndpoint: "http://10.43.0.88:9000"`) {
			t.Fatal("resolved endpoint not protected")
		}
	}
}

func TestExternalServiceRejectsUnsafeResolution(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint string
		service        map[string]interface{}
	}{
		{"missing", "http://minio.storage.svc:9000", nil},
		{"headless", "http://minio.storage.svc:9000", map[string]interface{}{"clusterIP": "None"}},
		{"wrong-port", "http://minio.storage.svc:9001", map[string]interface{}{"clusterIP": "10.43.0.88", "ports": []interface{}{map[string]interface{}{"port": 9000}}}},
		{"tls", "https://minio.storage.svc:9000", map[string]interface{}{"clusterIP": "10.43.0.88"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api, config := retainedCredentials(t)
			api.noGuard = true
			api.externalService = tc.service
			if _, err := renderRetained(t, config, externalValues(tc.endpoint)); err == nil {
				t.Fatal("unsafe Service accepted")
			}
		})
	}
}

func TestExternalPublicEndpointIsNotRewritten(t *testing.T) {
	_, output, err := render(t, externalValues("https://s3.example.com"))
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	if !strings.Contains(output, `name: ENDPOINT, value: "https://s3.example.com"`) {
		t.Fatal("public endpoint was rewritten")
	}
}

func TestExternalEndpointRetentionAndLegacyCompatibility(t *testing.T) {
	api, config := retainedCredentials(t)
	api.noGuard = true
	api.externalService = map[string]interface{}{"clusterIP": "10.43.0.88", "ports": []interface{}{map[string]interface{}{"port": 9000}}}
	values := externalValues("http://minio.storage.svc:9000")
	output, err := renderRetained(t, config, values)
	if err != nil {
		t.Fatal(err, output)
	}
	decoder := yaml.NewDecoder(strings.NewReader(output))
	for {
		var doc manifest
		if decoder.Decode(&doc) != nil {
			break
		}
		if doc["kind"] == "ConfigMap" && field(doc, "metadata", "name") == "qa-storage-identity" {
			api.identity = field(doc, "data", "identity").(string)
		}
	}
	api.noGuard = false
	api.externalEndpoint = "http://10.43.0.88:9000"
	if output, err = renderRetained(t, config, values); err != nil {
		t.Fatal("same endpoint resume rejected", err, output)
	}
	api.externalService["clusterIP"] = "10.43.0.89"
	if output, err = renderRetained(t, config, values); err == nil || !strings.Contains(output, "Service address changed") {
		t.Fatal("Service address drift accepted", err, output)
	}
	// Pre-fix identity has no resolved endpoint. Upgrade must not change the
	// immutable CR URLs or require recreation of a healthy legacy release.
	legacy := map[string]interface{}{"mode": "external", "endpoint": "http://minio.storage.svc:9000", "region": "us-east-1", "logs": "qa-logs", "data": "qa-data", "secret": "external-s3", "minioSecret": "qa-minio-root", "claim": "", "size": "100Gi", "storageClass": "local-path", "dbSecret": "qa-db-root", "lsId": 1001, "obId": 1002}
	data, _ := json.Marshal(legacy)
	api.identity = fmt.Sprintf("%x", sha256.Sum256(data))
	api.externalEndpoint = ""
	output, err = renderRetained(t, config, values)
	if err != nil || strings.Count(output, "?host=http://minio.storage.svc:9000&s3_region=") != 2 {
		t.Fatal("legacy URLs or identity changed", err, output)
	}
	if strings.Contains(output, "externalDatabaseEndpoint:") {
		t.Fatal("legacy storage silently migrated")
	}
}

func TestPreflightAddressingTracksLogService13(t *testing.T) {
	for endpoint, mode := range map[string]string{"http://192.0.2.10:9000": "on", "https://s3.example.com": "off"} {
		_, output, err := render(t, externalValues(endpoint))
		if err != nil {
			t.Fatal(err, output)
		}
		if !strings.Contains(output, `name: MC_PATH, value: "`+mode+`"`) || !strings.Contains(output, `--path "$MC_PATH"`) {
			t.Fatal("preflight addressing mismatch", endpoint)
		}
	}
}
