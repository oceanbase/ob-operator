/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanseek_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"text/template"

	sprig "github.com/go-task/slim-sprig"
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
			wantCheck := !native || stage != "cluster"
			if strings.Contains(output, "Verify runtime readiness") != wantCheck {
				t.Fatalf("unexpected readiness guidance for native=%v stage=%s:\n%s", native, stage, output)
			}
			if stage == "cluster" {
				if !strings.Contains(output, "mysql -h '192.0.2.10' -P 30881") || strings.Contains(output, "127.0.0.1") {
					t.Fatal("SQL must print the real node address", output)
				}
			} else if !strings.Contains(output, "SQL access becomes available after stage=cluster") || strings.Contains(output, "mysql -h") {
				t.Fatal("partial stages must defer SQL access", output)
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
		{"clusterIP", "ClusterIP", "kubectl -n qa-ns port-forward svc/oceanbase-dashboard-qa 8080:8088", "http://192.0.2.10:", true},
		{"nodePort", "NodePort", "Dashboard URL: http://192.0.2.10:31183/#/login", "port-forward svc/oceanbase-dashboard-", true},
		{"loadBalancer", "LoadBalancer", "Dashboard LoadBalancer address", "http://<EXTERNAL-IP", true},
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
				if !strings.Contains(output, "Username: admin") || !strings.Contains(output, "kubectl -n qa-ns get secret qa-dashboard-bootstrap") {
					t.Fatal("missing Dashboard credentials", output)
				}
			}
			if !strings.Contains(output, "kubectl -n qa-ns get secret existing-root-secret") {
				t.Fatal("missing configured SQL credential", output)
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
	start := strings.Index(source, "{{/* Presentation address only:")
	if start < 0 {
		t.Fatal("missing presentation address helpers")
	}
	source = source[start:]
	// The standalone fixture only executes presentation helpers; resource lookup
	// helpers below them require the complete Helm function environment.
	if end := strings.Index(source, "{{/* Preserve the webhook-generated account identity"); end >= 0 {
		source = source[:end]
	}
	var tpl *template.Template
	tpl = template.New("NOTES").Funcs(sprig.TxtFuncMap()).Funcs(template.FuncMap{
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

func TestAccessHostPrefersReadyExternalIP(t *testing.T) {
	node := func(ready, kind, address string) map[string]interface{} {
		return map[string]interface{}{"status": map[string]interface{}{
			"conditions": []map[string]string{{"type": "Ready", "status": ready}},
			"addresses":  []map[string]string{{"type": kind, "address": address}},
		}}
	}
	nodes := []map[string]interface{}{
		node("False", "ExternalIP", "192.0.2.1"),
		node("True", "InternalIP", "10.0.0.2"),
		node("True", "ExternalIP", "2001:db8::3"),
	}
	for _, tc := range []struct {
		count    int
		families []string
		want     string
	}{{1, nil, ""}, {2, nil, "10.0.0.2"}, {3, nil, "2001:db8::3"}, {3, []string{"IPv4"}, "10.0.0.2"}, {2, []string{"IPv6"}, ""}} {
		tpl := notesTemplate(t).Funcs(template.FuncMap{
			"lookup": func(_, kind, _, _ string) map[string]interface{} {
				if kind == "Service" {
					return map[string]interface{}{"spec": map[string]interface{}{"ipFamilies": tc.families}}
				}
				return map[string]interface{}{"items": nodes[:tc.count]}
			},
		})
		var out bytes.Buffer
		data := map[string]string{"namespace": "test", "service": "test-sql", "host": ""}
		if err := tpl.ExecuteTemplate(&out, "oceanseek.accessHost", data); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); got != tc.want {
			t.Errorf("with %d nodes, host = %q, want %q", tc.count, got, tc.want)
		}
	}
}

func TestAccessHostOverrideAndIPv6(t *testing.T) {
	for _, tc := range []struct{ host, want string }{{"", "192.0.2.10"}, {"db.example.com", "db.example.com"}, {"2001:db8::1", "2001:db8::1"}} {
		tpl := notesTemplate(t)
		var out bytes.Buffer
		if err := tpl.ExecuteTemplate(&out, "oceanseek.accessHost", map[string]string{"namespace": "test", "service": "test-sql", "host": tc.host}); err != nil {
			t.Fatal(err)
		}
		if out.String() != tc.want {
			t.Fatalf("address = %q, want %q", out.String(), tc.want)
		}
	}
	var out bytes.Buffer
	if err := notesTemplate(t).ExecuteTemplate(&out, "oceanseek.urlHost", "2001:db8::1"); err != nil {
		t.Fatal(err)
	}
	if out.String() != "[2001:db8::1]" {
		t.Fatal(out.String())
	}
}

// Protect the complete success output, including blank lines and section order.
func TestCompactInstallNotes(t *testing.T) {
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
	var values map[string]interface{}
	if err := yaml.Unmarshal(defaults, &values); err != nil {
		t.Fatal(err)
	}
	values["sqlService"].(map[string]interface{})["nodePort"] = 31783
	values["dashboard"].(map[string]interface{})["service"].(map[string]interface{})["nodePort"] = 32578
	var out bytes.Buffer
	if err := notes.Execute(&out, map[string]interface{}{
		"Release": map[string]string{"Name": "qa", "Namespace": "qa-ns"}, "Values": values,
	}); err != nil {
		t.Fatal(err)
	}
	const want = `Database SQL access
  mysql -h '192.0.2.10' -P 31783 -u 'root@sys' --protocol=TCP -p
Root password (in a trusted terminal):
  kubectl -n qa-ns get secret qa-db-root -o jsonpath='{.data.password}' | base64 --decode; echo

OceanBase Operator Dashboard
  Dashboard URL: http://192.0.2.10:32578/#/login
  Username: admin
Initial Dashboard password (in a trusted terminal):
  kubectl -n qa-ns get secret qa-dashboard-bootstrap -o jsonpath='{.data.password}' | base64 --decode; echo`
	if got := strings.TrimSpace(out.String()); got != want {
		t.Fatalf("expected compact access notes:\n%s\ngot:\n%s", want, got)
	}
}
