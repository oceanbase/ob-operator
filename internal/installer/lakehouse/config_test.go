package lakehouse

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigurationOverlayRetainsIdentityAndReplacesLists(t *testing.T) {
	base := map[string]interface{}{"observer": map[string]interface{}{"clusterId": float64(1910202), "image": "qa/observer", "topology": []interface{}{map[string]interface{}{"zone": "old"}}}, "dashboard": map[string]interface{}{"adminPassword": "private"}}
	changed := mergeValues(base, map[string]interface{}{"observer": map[string]interface{}{"topology": []interface{}{map[string]interface{}{"zone": "new"}}}})
	if pfGet(changed, "observer", "clusterId") != float64(1910202) || pfGet(changed, "observer", "image") != "qa/observer" || pfGet(changed, "dashboard", "adminPassword") != "private" {
		t.Fatal("existing identity/image/credential lost")
	}
	if pfGet(base, "observer", "topology").([]interface{})[0].(map[string]interface{})["zone"] != "old" {
		t.Fatal("overlay mutated input")
	}
	if pfGet(changed, "observer", "topology").([]interface{})[0].(map[string]interface{})["zone"] != "new" {
		t.Fatal("topology list not replaced")
	}
}

func TestPrivateStateAtomicAndRejectsSymlinkDirectory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "state", "values.yaml")
	if err := writePrivate(path, []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := writePrivate(path, []byte("new")); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "new" {
		t.Fatal("not replaced")
	}
	for _, p := range []string{path, filepath.Dir(path)} {
		st, _ := os.Stat(p)
		if st.Mode().Perm()&0077 != 0 {
			t.Fatalf("insecure mode: %s", p)
		}
	}
	link := filepath.Join(root, "symlink")
	if err := os.Symlink(filepath.Dir(path), link); err != nil {
		t.Fatal(err)
	}
	if err := writePrivate(filepath.Join(link, "values.yaml"), []byte("bad")); err == nil {
		t.Fatal("accepted symlink")
	}
	b, _ = os.ReadFile(path)
	if string(b) != "new" {
		t.Fatal("changed file through symlink")
	}
}

func TestOptionsRejectAmbiguousAndUnsafeInput(t *testing.T) {
	for _, args := range [][]string{{"positional"}, {"--release", "../x"}, {"--namespace", "x/y"}, {"--access", "public"}, {"--dashboard-host", "http://example.com"}, {"--dashboard-host", "a;echo"}} {
		if _, err := parseOptions(args, &strings.Builder{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	o, err := parseOptions([]string{"--dashboard-host", "8.147.71.250", "-f", "a", "-f", "b"}, &strings.Builder{})
	if err != nil {
		t.Fatal(err)
	}
	if len(o.files) != 2 || o.access != "nodeport" {
		t.Fatal("wrong defaults")
	}
}

func TestCommandCancellationBoundsChildProcesses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	begin := time.Now()
	_, err := execute(ctx, "bash", "-c", "sleep 30 & wait")
	if err == nil || time.Since(begin) > 3*time.Second {
		t.Fatalf("command descendants were not cancelled promptly: %v", err)
	}
}

func TestFreshIdentitiesPersistAndDoNotUseChartDefaults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "values"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "values", "quickstart.yaml"), []byte("logService:\n  resource: {cpu: '1'}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := initialValues(root)
	if err != nil {
		t.Fatal(err)
	}
	id := pfGet(v, "logService", "clusterId").(int64)
	other := pfGet(v, "observer", "clusterId").(int64)
	if id < 100000000 || other != id+1 {
		t.Fatal("invalid independent IDs")
	}
	if pfGet(v, "logService", "resource", "cpu") != "1" {
		t.Fatal("lost profile")
	}
}
