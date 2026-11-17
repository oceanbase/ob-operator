package lakehouse_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"
)

func TestManualHelmNotesReflectStage(t *testing.T) {
	source, err := os.ReadFile("templates/NOTES.txt")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := notesTemplate(t).Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, native := range []bool{false, true} {
		for _, stage := range []string{"storage", "logservice", "cluster"} {
			var out bytes.Buffer
			err := notes.Execute(&out, map[string]interface{}{"Release": map[string]string{"Name": "qa", "Namespace": "qa-ns"}, "Values": map[string]interface{}{"stage": stage, "access": map[string]string{"host": "192.0.2.10"}, "sqlService": map[string]interface{}{"type": "NodePort", "nodePort": 30881}, "native": map[string]bool{"enabled": native}, "objectStorage": map[string]string{"endpoint": "https://s3.example.com"}, "minio": map[string]bool{"enabled": false}, "dashboard": map[string]interface{}{"enabled": false}, "observer": map[string]string{"rootExistingSecret": ""}}})
			if err != nil {
				t.Fatal(err)
			}
			output := out.String()
			if strings.Contains(output, "install.sh") {
				t.Fatal("native Helm distribution must not ask users to invoke an installer")
			}
			if native && stage == "cluster" {
				for _, want := range []string{
					"Installation verification",
					"logs -f job/qa-readiness",
					"A successful readiness Job is removed automatically",
					"NotFound is expected",
				} {
					if !strings.Contains(output, want) {
						t.Fatalf("native readiness guidance missing %q:\n%s", want, output)
					}
				}
			} else {
				if !strings.Contains(output, "Verify runtime readiness") {
					t.Fatal("legacy staged mode must not assert readiness", output)
				}
			}
			if stage == "cluster" && (!strings.Contains(output, "mysql -h 192.0.2.10 -P 30881") || strings.Contains(output, "127.0.0.1")) {
				t.Fatal("SQL must print the real node address", output)
			}

			for _, want := range []string{
				"a Helm retry does not",
				"follow the Operator recovery procedure",
				"External object storage is configured; no bundled MinIO server is deployed",
				"The storage initialization Job uses mc (AGPL-3.0) to verify buckets",
			} {
				if !strings.Contains(output, want) {
					t.Fatalf("recovery or storage guidance missing %q:\n%s", want, output)
				}
			}
		}
	}
}

func TestDashboardNotesMatchAccessType(t *testing.T) {
	source, err := os.ReadFile("templates/NOTES.txt")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := notesTemplate(t).Parse(string(source))
	if err != nil {
		t.Fatal(err)
	}
	defaults, err := os.ReadFile("values.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, serviceType, expected, absent string
		enabled                             bool
	}{
		{"clusterIP", "ClusterIP", "8080:8088", "<reachable-node-IP>", true},
		{"nodePort", "NodePort", "http://192.0.2.10:31183/", "port-forward svc/", true},
		{"loadBalancer", "LoadBalancer", "http://<EXTERNAL-IP-or-hostname>:8088", "port-forward svc/", true},
		{"disabled", "ClusterIP", "Dashboard is disabled", "get secret qa-dashboard-bootstrap", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var values map[string]interface{}
			if err := yaml.Unmarshal(defaults, &values); err != nil {
				t.Fatal(err)
			}
			dashboard := values["dashboard"].(map[string]interface{})
			dashboard["enabled"] = tc.enabled
			dashboard["service"] = map[string]interface{}{"type": tc.serviceType, "port": 8088, "nodePort": 31183}
			values["observer"].(map[string]interface{})["rootExistingSecret"] = "existing-root-secret"
			var out bytes.Buffer
			if err := notes.Execute(&out, map[string]interface{}{
				"Release": map[string]string{"Name": "qa", "Namespace": "qa-ns"}, "Values": values,
			}); err != nil {
				t.Fatal(err)
			}
			output := out.String()
			if !strings.Contains(output, tc.expected) || strings.Contains(output, tc.absent) {
				t.Fatalf("wrong guidance for %s:\n%s", tc.name, output)
			}
			if tc.enabled {
				if !strings.HasPrefix(strings.TrimSpace(output), "OceanBase Operator Dashboard") ||
					!strings.Contains(output, "Username: admin") ||
					!strings.Contains(output, "separate from the database root@sys account") ||
					!strings.Contains(output, "kubectl -n qa-ns get secret qa-dashboard-bootstrap") {
					t.Fatal("access and account guidance must come before diagnostics")
				}
			}
			for _, want := range []string{"kubectl -n qa-ns get secret existing-root-secret", "logs -f job/qa-readiness", "Namespace deletion", "SAME release, namespace and values", "Bundled object storage uses MinIO and mc"} {
				if !strings.Contains(output, want) {
					t.Fatalf("missing diagnostic/retention guidance %q", want)
				}
			}
		})
	}
}

// Evaluate the actual address helper with a small Node fixture, including NOTES.
func notesTemplate(t *testing.T) *template.Template {
	t.Helper()
	helpers, err := os.ReadFile("templates/_helpers.tpl")
	if err != nil {
		t.Fatal(err)
	}
	source := string(helpers)
	source = source[strings.Index(source, "{{/* Presentation address only:"):]
	var tpl *template.Template
	tpl = template.New("NOTES").Funcs(template.FuncMap{
		"contains": func(sub, text string) bool { return strings.Contains(text, sub) },
		"replace":  func(old, new, text string) string { return strings.ReplaceAll(text, old, new) },
		"int":      func(value interface{}) int { n, _ := strconv.Atoi(fmt.Sprint(value)); return n },
		"default": func(fallback, value interface{}) interface{} {
			if value == nil || reflect.ValueOf(value).IsZero() {
				return fallback
			}
			return value
		},
		"lookup": func(_, _, _, _ string) map[string]interface{} {
			return map[string]interface{}{"items": []interface{}{
				map[string]interface{}{"status": map[string]interface{}{"conditions": []interface{}{map[string]string{"type": "Ready", "status": "True"}}, "addresses": []interface{}{map[string]string{"type": "InternalIP", "address": "192.0.2.10"}}}},
			}}
		},
		"include": func(name string, data interface{}) (string, error) {
			var out bytes.Buffer
			err := tpl.ExecuteTemplate(&out, name, data)
			return out.String(), err
		},
	})
	if _, err := tpl.Parse(source); err != nil {
		t.Fatal(err)
	}
	return tpl
}

func TestAccessHostOverrideAndIPv6(t *testing.T) {
	for _, tc := range []struct{ host, want string }{{"", "192.0.2.10"}, {"db.example.com", "db.example.com"}, {"2001:db8::1", "2001:db8::1"}} {
		tpl := notesTemplate(t)
		var out bytes.Buffer
		if err := tpl.ExecuteTemplate(&out, "lakehouse.accessHost", map[string]interface{}{"Values": map[string]interface{}{"access": map[string]string{"host": tc.host}}}); err != nil {
			t.Fatal(err)
		}
		if out.String() != tc.want {
			t.Fatalf("address = %q, want %q", out.String(), tc.want)
		}
	}
	var out bytes.Buffer
	if err := notesTemplate(t).ExecuteTemplate(&out, "lakehouse.urlHost", "2001:db8::1"); err != nil {
		t.Fatal(err)
	}
	if out.String() != "[2001:db8::1]" {
		t.Fatal(out.String())
	}
}
