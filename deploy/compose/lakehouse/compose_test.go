/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND.
*/

package lakehouse_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type service struct {
	Image       string            `json:"image"`
	Build       any               `json:"build"`
	Environment map[string]string `json:"environment"`
	DependsOn   map[string]struct {
		Condition string `json:"condition"`
	} `json:"depends_on"`
	Volumes []struct {
		Type     string `json:"type"`
		Source   string `json:"source"`
		Target   string `json:"target"`
		ReadOnly bool   `json:"read_only"`
	} `json:"volumes"`
	MemoryLimit json.Number `json:"mem_limit"`
	Ports       []struct {
		HostIP    string `json:"host_ip"`
		Published string `json:"published"`
	} `json:"ports"`
}

func renderCompose(t *testing.T) map[string]service {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker Compose CLI is required for configuration validation")
	}
	// Copy only the public configuration/scripts, proving it does not depend
	// on .env, source code, local secrets or preparation outputs.
	dir := t.TempDir()
	body, err := os.ReadFile("compose.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "scripts"), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("docker", "compose", "--env-file", "/dev/null", "-f", filepath.Join(dir, "compose.yaml"), "config", "--format", "json")
	// Defaults must not accidentally inherit a developer's Compose overrides.
	for _, kv := range os.Environ() {
		name := strings.SplitN(kv, "=", 2)[0]
		if name == "PATH" || name == "HOME" || name == "DOCKER_CONFIG" || name == "DOCKER_CLI_PLUGIN_EXTRA_DIRS" {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("compose config: %v", err)
	}
	var cfg struct {
		Services map[string]service `json:"services"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg.Services
}

func TestQuickstartWithoutPreparation(t *testing.T) {
	services := renderCompose(t)
	if len(services) != 9 {
		t.Fatalf("expected five services and four initialization containers, got %d", len(services))
	}
	for name, s := range services {
		if s.Image == "" || s.Build != nil {
			t.Fatalf("%s requires a missing image or local build", name)
		}
		if strings.HasSuffix(s.Image, ":latest") {
			t.Fatalf("%s uses floating image", name)
		}
		for _, mount := range s.Volumes {
			if mount.Type == "bind" && !strings.HasSuffix(mount.Source, "/scripts") {
				t.Fatalf("%s needs host-prepared files: %+v", name, mount)
			}
		}
	}
	for _, tc := range []struct{ service, dependency, condition string }{
		{"minio", "credentials-init", "service_completed_successfully"},
		{"minio-init", "minio", "service_healthy"},
		{"logservice-bootstrap", "minio-init", "service_completed_successfully"},
		{"logservice-bootstrap", "logservice-1", "service_healthy"},
		{"logservice-bootstrap", "logservice-2", "service_healthy"},
		{"logservice-bootstrap", "logservice-3", "service_healthy"},
		{"observer", "logservice-bootstrap", "service_completed_successfully"},
		{"observer-bootstrap", "observer", "service_healthy"},
	} {
		if services[tc.service].DependsOn[tc.dependency].Condition != tc.condition {
			t.Fatalf("%s must wait for %s (%s)", tc.service, tc.dependency, tc.condition)
		}
	}
}

func TestPrivateCredentialsAndBoundedResources(t *testing.T) {
	services := renderCompose(t)
	for _, name := range []string{"credentials-init", "minio", "minio-init", "logservice-bootstrap", "observer-bootstrap"} {
		found := false
		for _, mount := range services[name].Volumes {
			if mount.Target != "/run/secrets" {
				continue
			}
			found = true
			if mount.Type != "volume" || mount.Source != "credentials" {
				t.Fatalf("%s credentials are not persistent named storage", name)
			}
			if name != "credentials-init" && name != "observer-bootstrap" && !mount.ReadOnly {
				t.Fatalf("%s can change credentials", name)
			}
		}
		if !found {
			t.Fatalf("%s cannot read credentials", name)
		}
	}
	for _, name := range []string{"logservice-1", "logservice-2", "logservice-3"} {
		s := services[name]
		memoryLimit, err := s.MemoryLimit.Int64()
		if err != nil || memoryLimit != 4*1024*1024*1024 || s.Environment["MEMORY_LIMIT"] != "3G" {
			t.Fatalf("%s resource defaults lost", name)
		}
	}
	for name, s := range services {
		if name != "observer" && len(s.Ports) > 0 {
			t.Fatalf("%s publishes unnecessary host ports", name)
		}
	}
	ports := services["observer"].Ports
	if len(ports) != 1 || ports[0].HostIP != "127.0.0.1" || ports[0].Published != "32881" {
		t.Fatalf("unexpected SQL exposure %+v", ports)
	}
}
