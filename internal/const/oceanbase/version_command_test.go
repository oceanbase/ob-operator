package oceanbase

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCommandUsesObserverOwner(t *testing.T) {
	for _, version := range []string{"4.6.3.1-282026111420.al8", "4.6.3.0-82026110121.al8", "4.2.1.0-100000.el8"} {
		t.Run(version, func(t *testing.T) {
			dir := t.TempDir()
			// Package names are deliberately unavailable: only file ownership works.
			script := "#!/bin/sh\n[ \"$1\" = -qf ] && [ \"$4\" = /home/admin/oceanbase/bin/observer ] || exit 9\nprintf '%s' \"$RPM_TEST_VERSION\"\n"
			if err := os.WriteFile(filepath.Join(dir, "rpm"), []byte(script), 0755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("/bin/sh", "-c", CmdVersion)
			cmd.Env = append(os.Environ(), "PATH="+dir, "RPM_TEST_VERSION="+version)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			want := version[:strings.LastIndex(version, ".")]
			if strings.TrimSpace(string(out)) != want {
				t.Fatalf("got %q, want %q", out, want)
			}
		})
	}
}

func TestVersionCommandPreservesRPMFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rpm"), []byte("#!/bin/sh\necho 'file is not owned by any package' >&2\nexit 7\n"), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", "-c", CmdVersion)
	cmd.Env = append(os.Environ(), "PATH="+dir)
	if out, err := cmd.Output(); err == nil || len(out) != 0 {
		t.Fatalf("RPM failure hidden: err=%v stdout=%q", err, out)
	}
}
