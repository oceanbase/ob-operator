package lakehouse_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"
)

func TestManualHelmNotesReflectStage(t *testing.T) {
	source, err := os.ReadFile("templates/NOTES.txt")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := template.New("NOTES").Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, native := range []bool{false, true} {
		for _, stage := range []string{"storage", "logservice", "cluster"} {
			var out bytes.Buffer
			err := notes.Execute(&out, map[string]interface{}{"Release": map[string]string{"Name": "qa", "Namespace": "qa-ns"}, "Values": map[string]interface{}{"stage": stage, "native": map[string]bool{"enabled": native}, "objectStorage": map[string]string{"mode": "external"}, "dashboard": map[string]interface{}{"enabled": false}, "observer": map[string]string{"rootExistingSecret": ""}}})
			if err != nil {
				t.Fatal(err)
			}
			output := out.String()
			if strings.Contains(output, "install.sh") {
				t.Fatal("native Helm distribution must not ask users to invoke an installer")
			}
			if native && stage == "cluster" {
				if !strings.Contains(output, "hooks enabled") || !strings.Contains(output, "post-install/post-upgrade") {
					t.Fatal("native readiness contract missing", output)
				}
			} else {
				if !strings.Contains(output, "Verify runtime readiness") {
					t.Fatal("legacy staged mode must not assert readiness", output)
				}
			}
			if !strings.Contains(output, "port-forward pod/<observer-pod-name>") || !strings.Contains(output, "root@sys") {
				t.Fatal("SQL login guidance missing")
			}
		}
	}
}
