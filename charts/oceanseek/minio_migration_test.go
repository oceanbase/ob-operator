package lakehouse_test

import (
	"strings"
	"testing"
)

// MinIO is an independent subchart: parent storage and scheduling settings do
// not migrate automatically. This is also the contract of the custom guide.
func TestMinIOConfigurationIsIndependent(t *testing.T) {
	values := map[string]interface{}{
		"storageClass": "database-class",
		"scheduling":   map[string]interface{}{"nodeSelector": map[string]string{"disk": "database"}},
		"minio":        map[string]interface{}{"storageClass": "object-class", "nodeSelector": map[string]string{"disk": "objects"}},
	}
	docs, output, err := render(t, values)
	if err != nil {
		t.Fatal(err, output)
	}
	if field(find(t, docs, "PersistentVolumeClaim", "qa-minio-data"), "spec", "storageClassName") != "object-class" {
		t.Fatal("MinIO storage class must be independently configurable")
	}
	if field(find(t, docs, "StatefulSet", "qa-minio"), "spec", "template", "spec", "nodeSelector", "disk") != "objects" {
		t.Fatal("MinIO scheduling must use its own values")
	}
}

func TestDisabledMinIORequiresExternalEndpoint(t *testing.T) {
	_, output, err := render(t, map[string]interface{}{"minio": map[string]interface{}{"enabled": false}})
	if err == nil || !strings.Contains(output, "configure objectStorage.endpoint or enable minio.enabled") {
		t.Fatal("disabled storage must require an external endpoint", err, output)
	}
}
