package obcluster

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func TestMigrationComparesRPMJobOutputWithObserverBuild(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rpm"), []byte("#!/bin/sh\nprintf '%s' '4.2.1.3-103000032023122818.el8'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", "-c", oceanbaseconst.CmdVersion)
	cmd.Env = append(os.Environ(), "PATH="+dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		build string
		match bool
	}{
		{"4.2.1.3_103000032023122818-8fe69c20", true},
		{"4.2.1.3_103000032023122819-8fe69c20", false},
		{"4.2.1.7_103000032023122818-8fe69c20", false},
	} {
		source, err := model.ParseOBVersion(tc.build)
		if err != nil {
			t.Fatal(err)
		}
		if err := checkMigrationVersion(source, string(output)); (err == nil) != tc.match {
			t.Fatalf("source=%s rpm output=%q match=%v: %v", source, output, tc.match, err)
		}
	}
}
