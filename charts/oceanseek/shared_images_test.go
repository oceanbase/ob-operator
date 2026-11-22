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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Render both public entry points: matching-looking values files can still be
// overridden differently by Compose interpolation or Helm templates.
func TestSharedDatabaseImagesMatchCompose(t *testing.T) {
	docs, output, err := render(t, nil)
	if err != nil {
		t.Fatalf("helm template: %v\n%s", err, output)
	}
	envFile := filepath.Join(t.TempDir(), "empty.env")
	if err := os.WriteFile(envFile, nil, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("docker", "compose", "--env-file", envFile, "-f", "../../deploy/compose/oceanseek/compose.yaml", "config", "--format", "json")
	for _, entry := range os.Environ() {
		key := strings.SplitN(entry, "=", 2)[0]
		if key == "PATH" || key == "HOME" || key == "DOCKER_CONFIG" || key == "DOCKER_CLI_PLUGIN_EXTRA_DIRS" {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	body, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker compose config: %v\n%s", err, body)
	}
	var compose struct {
		Services map[string]struct{ Image string } `json:"services"`
	}
	if err := json.Unmarshal(body, &compose); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ kind, name, section, service string }{
		{"OBCluster", "qa-ob", "observer", "observer"},
		{"OBCluster", "qa-ob", "monitor", "obagent"},
		{"OBLogServiceCluster", "qa-ls", "logService", "logservice-1"},
	} {
		actual := field(find(t, docs, tc.kind, tc.name), "spec", tc.section, "image")
		want := compose.Services[tc.service].Image
		if want == "" || actual != want {
			t.Errorf("%s differs between Helm (%v) and Compose (%q)", tc.section, actual, want)
		}
	}
}
