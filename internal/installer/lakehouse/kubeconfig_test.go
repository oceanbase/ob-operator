package lakehouse

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallerKubeconfigPreservesSelectionAndFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	file := filepath.Join(t.TempDir(), "k3s.yaml")
	if err := os.WriteFile(file, []byte("selected fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := installerKubeconfig("", "", file); err != nil || got != file {
		t.Fatalf("K3s fallback: %q %v", got, err)
	}
	if got, err := installerKubeconfig("", "/missing/intentionally-selected", file); err != nil || got != "" {
		t.Fatal("environment selection changed")
	}
	if got, err := installerKubeconfig(file, "/other", file); err != nil || got != file {
		t.Fatal("explicit selection lost")
	}
	if _, err := installerKubeconfig(file+"missing", "", file); err == nil {
		t.Fatal("invalid explicit configuration silently fell back")
	}
	if err := os.MkdirAll(filepath.Join(home, ".kube"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".kube/config"), []byte("default"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := installerKubeconfig("", "", file); err != nil || got != "" {
		t.Fatal("user default replaced")
	}
	if data, _ := os.ReadFile(filepath.Join(home, ".kube/config")); string(data) != "default" {
		t.Fatal("user configuration modified")
	}
}

func TestInstallerKubeconfigReachesHelmAndChildShell(t *testing.T) {
	t.Setenv("KUBECONFIG", "original")
	ctx := context.WithValue(context.Background(), kubeconfigContextKey{}, "/selected/config")
	for _, name := range []string{"helm", "kubectl", "bash"} {
		cmd := command(ctx, name)
		found := false
		for _, v := range cmd.Environ() {
			if v == "KUBECONFIG=/selected/config" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s lost selected config", name)
		}
	}
	if os.Getenv("KUBECONFIG") != "original" {
		t.Fatal("configuration leaked into global process environment")
	}
}
