package lakehouse

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// Regression found during independent whole-entry review: existing controller
// Deployments can remain healthy while only some cert-manager CRDs survive.
func TestReviewPartialCertManagerBlocksOperatorInstallation(t *testing.T) {
	mutated := false
	run := func(_ context.Context, tool string, args ...string) ([]byte, error) {
		joined := " " + strings.Join(args, " ") + " "
		if tool == "kubectl" && strings.Contains(joined, " get crds ") {
			return []byte(`{"items":[{"metadata":{"name":"certificates.cert-manager.io"}}]}`), nil
		}
		if tool == "kubectl" && strings.Contains(joined, " rollout status ") {
			return nil, nil
		}
		if tool == "helm" || strings.Contains(joined, " apply ") {
			mutated = true
			return nil, nil
		}
		return nil, errors.New("unexpected mock command")
	}
	err := prepareOperator(context.Background(), run, options{bundle: t.TempDir()}, PreflightResult{NodeCount: 1}, func(string) {})
	if err == nil || !(strings.Contains(strings.ToLower(err.Error()), "partial") || strings.Contains(strings.ToLower(err.Error()), "incomplete")) || mutated {
		t.Fatalf("partial existing cert-manager must be rejected before mutation: mutated=%v error=%v", mutated, err)
	}
}

func TestReviewOldDefaultsRemainEffectiveOnUpgrade(t *testing.T) {
	effective := map[string]interface{}{
		"stage":         "cluster",
		"storageClass":  "old-local",
		"observer":      map[string]interface{}{"clusterId": 123, "image": "old-ob", "resource": map[string]interface{}{"cpu": "2", "memory": "10Gi"}},
		"logService":    map[string]interface{}{"clusterId": 122, "image": "old-ls"},
		"objectStorage": map[string]interface{}{"endpoint": "https://old.example", "existingSecret": "retained"},
		"dashboard":     map[string]interface{}{"service": map[string]interface{}{"type": "LoadBalancer"}, "image": "old-dashboard"},
	}
	user := map[string]interface{}{"observer": map[string]interface{}{"resource": map[string]interface{}{"cpu": "3"}}}
	nextDefaults := map[string]interface{}{"storageClass": "new-default", "observer": map[string]interface{}{"clusterId": 999, "image": "new-ob"}, "dashboard": map[string]interface{}{"service": map[string]interface{}{"type": "NodePort"}, "newOption": true}}
	resolved := mergeValues(nextDefaults, retainConfiguration(effective, user))
	for _, check := range []struct {
		path []string
		want string
	}{
		{[]string{"storageClass"}, "old-local"},
		{[]string{"observer", "clusterId"}, "123"},
		{[]string{"observer", "image"}, "old-ob"},
		{[]string{"observer", "resource", "cpu"}, "3"},
		{[]string{"observer", "resource", "memory"}, "10Gi"},
		{[]string{"objectStorage", "existingSecret"}, "retained"},
		{[]string{"dashboard", "service", "type"}, "LoadBalancer"},
		{[]string{"dashboard", "newOption"}, "true"},
	} {
		if got := pfString(pfGet(resolved, check.path...)); got != check.want {
			t.Fatalf("%v changed: got %s want %s", check.path, got, check.want)
		}
	}
	if _, exists := resolved["stage"]; exists {
		t.Fatal("old stage froze new staged installer")
	}
}

func TestReviewAccessFlagDistinguishesDefaultFromExplicitOverride(t *testing.T) {
	implicit, err := parseOptions(nil, io.Discard)
	if err != nil || implicit.accessExplicit {
		t.Fatalf("default access became a user override: %+v %v", implicit, err)
	}
	for _, args := range [][]string{{"--access", "nodeport"}, {"--access=clusterip"}} {
		explicit, err := parseOptions(args, io.Discard)
		if err != nil || !explicit.accessExplicit {
			t.Fatalf("explicit access flag lost: %+v %v", explicit, err)
		}
	}
}

func reviewCommandFixtures(t *testing.T) (string, string, string) {
	t.Helper()
	bundle := t.TempDir()
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"kubectl": "#!/bin/sh\nprintf '%s\\n' '{\"metadata\":{\"uid\":\"test-cluster\"}}'\n",
		"helm":    "#!/bin/sh\nexit 7\n",
		"jq":      "#!/bin/sh\nexit 7\n",
	} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	key := fmt.Sprintf("%x", sha256.Sum256([]byte("test-cluster/ob-helm-demo/ob-demo")))[:16]
	return bundle, filepath.Join(bundle, ".lakehouse-install", key), filepath.Join(t.TempDir(), "cache")
}

func TestReviewInstallSharesLockAcrossBundlesBeforeReadingReleaseOrState(t *testing.T) {
	bundle, state, cache := reviewCommandFixtures(t)
	lockDir := filepath.Join(cache, "oceanbase-lakehouse", "locks")
	if err := privateDirectory(lockDir); err != nil {
		t.Fatal(err)
	}
	lock, err := os.OpenFile(filepath.Join(lockDir, filepath.Base(state)), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	cacheDirectory := func() (string, error) { return cache, nil }
	for _, directory := range []string{bundle, t.TempDir()} {
		err = runWithCacheDirectory(context.Background(), []string{"--bundle", directory}, io.Discard, io.Discard, cacheDirectory)
		if err == nil || !strings.Contains(err.Error(), "another installation") {
			t.Fatalf("bundle %s bypassed shared lock or read Helm/config first: %v", directory, err)
		}
		if _, err := os.Stat(filepath.Join(directory, ".lakehouse-install")); !os.IsNotExist(err) {
			t.Fatalf("locked installation created bundle state: %v", err)
		}
	}
	// A different release in the same cluster is independent of the held lock.
	err = runWithCacheDirectory(context.Background(), []string{"--bundle", bundle, "--release", "another-demo"}, io.Discard, io.Discard, cacheDirectory)
	if err == nil || !strings.Contains(err.Error(), "Helm release") {
		t.Fatalf("unrelated release incorrectly shares lock: %v", err)
	}
}

func TestReviewCheckDoesNotCreateLocalState(t *testing.T) {
	bundle, state, cache := reviewCommandFixtures(t)
	err := runWithCacheDirectory(context.Background(), []string{"--bundle", bundle, "--check"}, io.Discard, io.Discard, func() (string, error) {
		t.Error("--check looked up installation cache")
		return cache, nil
	})
	if err == nil || !strings.Contains(err.Error(), "Helm release") {
		t.Fatalf("expected read failure from mock Helm: %v", err)
	}
	if _, err = os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("--check created local installation state: %v", err)
	}
	if _, err = os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("--check created installation cache: %v", err)
	}
}

func TestReviewInstallationLockRemainsPrivateAndReleasedAfterFailure(t *testing.T) {
	bundle, state, cache := reviewCommandFixtures(t)
	cacheDirectory := func() (string, error) { return cache, nil }
	lockDir := filepath.Join(cache, "oceanbase-lakehouse", "locks")
	lockPath := filepath.Join(lockDir, filepath.Base(state))
	for attempt := 0; attempt < 2; attempt++ {
		err := runWithCacheDirectory(context.Background(), []string{"--bundle", bundle}, io.Discard, io.Discard, cacheDirectory)
		if err == nil || !strings.Contains(err.Error(), "Helm release") {
			t.Fatalf("attempt %d did not release lock after mock failure: %v", attempt, err)
		}
		for _, path := range []string{filepath.Dir(lockDir), lockDir} {
			if info, err := os.Stat(path); err != nil || info.Mode().Perm() != 0700 {
				t.Fatalf("lock directory mode: %v %v", info, err)
			}
		}
		if info, err := os.Stat(lockPath); err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("lock file mode: %v %v", info, err)
		}
		if attempt == 0 {
			if err := os.Chmod(lockPath, 0666); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(lockDir, 0777); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestReviewInstallationLockRejectsFileSymlink(t *testing.T) {
	bundle, state, cache := reviewCommandFixtures(t)
	lockDir := filepath.Join(cache, "oceanbase-lakehouse", "locks")
	if err := privateDirectory(lockDir); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte("unchanged"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(lockDir, filepath.Base(state))); err != nil {
		t.Fatal(err)
	}
	err := runWithCacheDirectory(context.Background(), []string{"--bundle", bundle}, io.Discard, io.Discard, func() (string, error) { return cache, nil })
	if err == nil || strings.Contains(err.Error(), "Helm release") {
		t.Fatalf("followed lock symlink: %v", err)
	}
	if info, err := os.Stat(target); err != nil || info.Mode().Perm() != 0644 {
		t.Fatalf("symlink target was changed: %v %v", info, err)
	}
}

func TestReviewPrivateStateRemainsPrivateAndRejectsDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	state := filepath.Join(root, "state", "values.yaml")
	if err := writePrivate(state, []byte("observer:\n  clusterId: 123\n")); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(state); err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("state mode: %v %v", info, err)
	}
	if info, err := os.Stat(filepath.Dir(state)); err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("directory mode: %v %v", info, err)
	}
	if err := writePrivate(state, []byte("observer:\n  clusterId: 456\n")); err != nil {
		t.Fatal(err)
	}
	values, err := readValues(state)
	if err != nil || pfString(pfGet(values, "observer", "clusterId")) != "456" {
		t.Fatalf("atomic state replacement unreadable: %v %v", values, err)
	}
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := writePrivate(filepath.Join(link, "values.yaml"), []byte("x: y\n")); err == nil {
		t.Fatal("symlink state directory accepted")
	}
	if _, err := os.Stat(filepath.Join(target, "values.yaml")); !os.IsNotExist(err) {
		t.Fatal("state escaped into symlink target")
	}
}

func TestReviewConfigRejectsDuplicateKeysAndPreservesExplicitIdentity(t *testing.T) {
	file := filepath.Join(t.TempDir(), "values.yaml")
	if err := os.WriteFile(file, []byte("observer:\n  clusterId: 123\n  clusterId: 456\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readValues(file); err == nil {
		t.Fatal("duplicate identity keys accepted")
	}
	base := map[string]interface{}{"observer": map[string]interface{}{"clusterId": 123, "resource": map[string]interface{}{"cpu": "2", "memory": "10Gi"}}}
	overlay := map[string]interface{}{"observer": map[string]interface{}{"resource": map[string]interface{}{"cpu": "3"}}}
	merged := mergeValues(base, overlay)
	if pfString(pfGet(merged, "observer", "clusterId")) != "123" || pfGet(merged, "observer", "resource", "cpu") != "3" || pfGet(merged, "observer", "resource", "memory") != "10Gi" {
		t.Fatalf("incomplete override replaced existing identity: %v", merged)
	}
	if pfGet(base, "observer", "resource", "cpu") != "2" {
		t.Fatal("merge changed caller's state")
	}
}
