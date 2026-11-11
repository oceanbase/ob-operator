package lakehouse_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Release artifacts must carry the same API contract as controller-gen's
// checked-in sources. This repository-only test is excluded from chart archives
// by .helmignore; installed charts need neither the repository nor Go tooling.
func releaseDocuments(t *testing.T, data []byte) []manifest {
	t.Helper()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var objects []manifest
	for {
		var object map[string]interface{}
		if err := decoder.Decode(&object); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if len(object) != 0 {
			objects = append(objects, manifest(object))
		}
	}
	return objects
}

func releaseSource(t *testing.T, path string) []manifest {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", path))
	if err != nil {
		t.Fatal(err)
	}
	return releaseDocuments(t, data)
}

func releaseWithoutDescriptions(value interface{}) interface{} {
	switch value := value.(type) {
	case map[string]interface{}:
		copy := map[string]interface{}{}
		for key, item := range value {
			if key != "description" {
				copy[key] = releaseWithoutDescriptions(item)
			}
		}
		return copy
	case []interface{}:
		copy := make([]interface{}, len(value))
		for i, item := range value {
			copy[i] = releaseWithoutDescriptions(item)
		}
		return copy
	default:
		return value
	}
}

func releaseWebhook(t *testing.T, objects []manifest, name string) map[string]interface{} {
	t.Helper()
	for _, object := range objects {
		webhooks, _ := object["webhooks"].([]interface{})
		for _, webhook := range webhooks {
			entry := webhook.(map[string]interface{})
			if entry["name"] == name {
				return entry
			}
		}
	}
	t.Fatalf("missing webhook %s", name)
	return nil
}

func TestReleaseManifestSSContract(t *testing.T) {
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		helm = "helm"
	}
	const chartNamespace = "manifest-contract-namespace"
	rendered, err := exec.Command(helm, "template", "contract", "../ob-operator", "-n", chartNamespace).CombinedOutput()
	if err != nil {
		t.Fatalf("Operator chart render: %v\n%s", err, rendered)
	}
	for _, source := range []struct {
		name, namespace string
		objects         []manifest
	}{
		{"deploy", "oceanbase-system", releaseSource(t, "deploy/operator.yaml")},
		{"operator-chart", chartNamespace, releaseDocuments(t, rendered)},
	} {
		t.Run(source.name, func(t *testing.T) {
			for _, name := range []string{"obclusters", "oblogserviceclusters", "oblogservicezones", "oblogservicenodes"} {
				t.Run(name, func(t *testing.T) {
					base := releaseSource(t, "config/crd/bases/oceanbase.oceanbase.com_"+name+".yaml")[0]
					actual := find(t, source.objects, "CustomResourceDefinition", name+".oceanbase.oceanbase.com")
					for _, property := range []string{"group", "names", "scope", "versions"} {
						if !reflect.DeepEqual(releaseWithoutDescriptions(field(base, "spec", property)), releaseWithoutDescriptions(field(actual, "spec", property))) {
							t.Errorf("%s spec.%s differs from canonical generated CRD", name, property)
						}
					}
				})
			}
			for _, name := range []string{"moblogservicecluster.kb.io", "voblogservicecluster.kb.io"} {
				t.Run(name, func(t *testing.T) {
					canonical := releaseWebhook(t, releaseSource(t, "config/webhook/manifests.yaml"), name)
					service := canonical["clientConfig"].(map[string]interface{})["service"].(map[string]interface{})
					service["name"], service["namespace"] = "oceanbase-webhook-service", source.namespace
					actual := releaseWebhook(t, source.objects, name)
					if !reflect.DeepEqual(canonical, actual) {
						t.Errorf("%s differs from canonical webhook including namespace", name)
					}
				})
			}
		})
	}
}
