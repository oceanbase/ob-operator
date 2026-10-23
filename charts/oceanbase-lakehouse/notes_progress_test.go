package lakehouse_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"
)

// NOTES uses only standard Go-template actions, exactly as Helm renders them.
// No cluster or release is touched while checking the manual-Helm guidance.
func TestManualHelmNotesReflectStage(t *testing.T) {
	source, err := os.ReadFile("templates/NOTES.txt")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := template.New("NOTES").Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range []string{"storage", "logservice", "cluster"} {
		t.Run(stage, func(t *testing.T) {
			var out bytes.Buffer
			err := notes.Execute(&out, map[string]interface{}{
				"Release": map[string]string{"Name": "qa", "Namespace": "qa-ns"},
				"Values":  map[string]interface{}{"stage": stage, "objectStorage": map[string]string{"mode": "external"}, "dashboard": map[string]interface{}{"enabled": false}},
			})
			if err != nil {
				t.Fatal(err)
			}
			output := out.String()
			if stage == "cluster" {
				if strings.Contains(output, "For a complete, ordered installation run:") || strings.Contains(output, "-f your-values.yaml") {
					t.Fatalf("cluster stage must not tell an active or repeated install to start again: %s", output)
				}
				if !strings.Contains(output, "Verify runtime readiness") || strings.Contains(output, "Lakehouse ready:") {
					t.Fatal("Helm status alone must not assert runtime readiness", output)
				}
			} else if !strings.Contains(output, "manual Helm") || !strings.Contains(output, "install.sh qa qa-ns") {
				t.Fatal("partial manual Helm install needs explicit scoped continuation guidance", output)
			}
		})
	}
}
