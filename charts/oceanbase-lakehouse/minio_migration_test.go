package lakehouse_test

import (
	"reflect"
	"strings"
	"testing"
)

func TestMigratedMinIOValuesPreserveCurrentStorage(t *testing.T) {
	original := map[string]interface{}{
		"storageClass":     "local-path",
		"scheduling":       map[string]interface{}{"nodeSelector": map[string]interface{}{"disk": "ssd"}, "tolerations": []interface{}{}},
		"imagePullSecrets": []interface{}{map[string]interface{}{"name": "registry"}},
	}
	before, output, err := render(t, original)
	if err != nil {
		t.Fatalf("original values: %v\n%s", err, output)
	}
	original["minio"] = map[string]interface{}{"enabled": true, "storageClass": "local-path", "region": "us-east-1", "nodeSelector": map[string]interface{}{"disk": "ssd"}, "tolerations": []interface{}{}, "imagePullSecrets": []interface{}{map[string]interface{}{"name": "registry"}}}
	after, output, err := render(t, original)
	if err != nil {
		t.Fatalf("migrated values: %v\n%s", err, output)
	}
	for _, resource := range [][2]string{{"StatefulSet", "qa-minio"}, {"Service", "qa-minio"}, {"PersistentVolumeClaim", "qa-minio-data"}, {"ConfigMap", "qa-storage-identity"}} {
		if !reflect.DeepEqual(find(t, before, resource[0], resource[1]), find(t, after, resource[0], resource[1])) {
			t.Fatalf("migration changed %v", resource)
		}
	}
}

func TestMigratedMinIOValuesRejectConflictingOverrides(t *testing.T) {
	for _, name := range []string{"enabled", "storageClass", "region", "nodeSelector", "tolerations", "imagePullSecrets"} {
		values := map[string]interface{}{"enabled": false, "storageClass": "another-class", "region": "another-region", "nodeSelector": map[string]interface{}{"disk": "ssd"}, "tolerations": []interface{}{map[string]interface{}{"operator": "Exists"}}, "imagePullSecrets": []interface{}{map[string]interface{}{"name": "another-registry"}}}
		_, output, err := render(t, map[string]interface{}{"minio": map[string]interface{}{name: values[name]}})
		if err == nil || !strings.Contains(output, "migrated minio."+name+" must match") {
			t.Fatalf("conflicting %s was not clearly rejected: %v\n%s", name, err, output)
		}
	}
}
