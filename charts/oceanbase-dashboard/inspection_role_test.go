/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package dashboard_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Render the configuration directly as well, so the release-name contract is
// covered even when Helm is not installed in the Go test environment.
func TestInspectionRoleConfiguration(t *testing.T) {
	data, err := os.ReadFile("values.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	for _, release := range []string{"oceanbase-dashboard", "dashboard-qa"} {
		t.Run(release, func(t *testing.T) {
			source, err := os.ReadFile("templates/configmap.yaml")
			if err != nil {
				t.Fatal(err)
			}
			tmpl, err := template.New("configmap").Parse(string(source))
			if err != nil {
				t.Fatal(err)
			}
			var rendered bytes.Buffer
			if err := tmpl.Execute(&rendered, map[string]interface{}{"Release": map[string]string{"Name": release, "Namespace": "dashboard-system"}, "Values": values}); err != nil {
				t.Fatal(err)
			}
			var cm struct {
				Data map[string]string `yaml:"data"`
			}
			if err := yaml.Unmarshal(rendered.Bytes(), &cm); err != nil {
				t.Fatal(err)
			}
			var cfg struct {
				Inspection struct {
					ClusterRoleName string `yaml:"clusterRoleName"`
				} `yaml:"inspection"`
			}
			if err := yaml.Unmarshal([]byte(cm.Data["config.yaml"]), &cfg); err != nil {
				t.Fatal(err)
			}
			if want := release + "-cluster-role"; cfg.Inspection.ClusterRoleName != want {
				t.Fatalf("inspection role = %q, want release-scoped role %q", cfg.Inspection.ClusterRoleName, want)
			}
		})
	}
}

func TestInspectionRoleMatchesHelmRelease(t *testing.T) {
	helm := os.Getenv("HELM_BINARY")
	if helm == "" {
		var err error
		helm, err = exec.LookPath("helm")
		if err != nil {
			t.Skip("helm is required to render the Dashboard chart")
		}
	}
	for _, release := range []string{"oceanbase-dashboard", "dashboard-qa"} {
		t.Run(release, func(t *testing.T) {
			output, err := exec.Command(helm, "template", release, ".", "--namespace", "dashboard-system").CombinedOutput()
			if err != nil {
				t.Fatalf("helm template: %v\n%s", err, output)
			}
			var roleName, inspectionRole string
			decoder := yaml.NewDecoder(strings.NewReader(string(output)))
			for {
				var doc struct {
					Kind     string `yaml:"kind"`
					Metadata struct {
						Name string `yaml:"name"`
					} `yaml:"metadata"`
					Data map[string]string `yaml:"data"`
				}
				if err := decoder.Decode(&doc); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				if doc.Kind == "ClusterRole" {
					roleName = doc.Metadata.Name
				}
				if config, ok := doc.Data["config.yaml"]; ok {
					var cfg struct {
						Inspection struct {
							ClusterRoleName string `yaml:"clusterRoleName"`
						} `yaml:"inspection"`
					}
					if err := yaml.Unmarshal([]byte(config), &cfg); err != nil {
						t.Fatal(err)
					}
					inspectionRole = cfg.Inspection.ClusterRoleName
				}
			}
			if roleName != release+"-cluster-role" {
				t.Fatalf("rendered role = %q", roleName)
			}
			if inspectionRole != roleName {
				t.Fatalf("inspection role = %q, rendered ClusterRole = %q; custom release inspection cannot bind its role", inspectionRole, roleName)
			}
		})
	}
}
