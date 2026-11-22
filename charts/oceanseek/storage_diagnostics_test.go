package oceanseek_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
)

func TestStorageFailureDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, mode, shared, failSide, operation string
		create                                  bool
	}{
		{"log-access", "external", "", "log", "stat", false},
		{"data-access", "external", "", "data", "stat", false},
		{"log-alias", "external", "", "log", "alias", false},
		{"data-alias", "external", "", "data", "alias", false},
		{"log-create", "external", "", "log", "mb", true},
		{"data-create", "external", "", "data", "mb", true},
		{"legacy-shared", "external", "shared-credentials", "data", "stat", false},
		{"bundled-root", "bundled", "", "data", "stat", false},
		{"success", "external", "", "none", "stat", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage := map[string]interface{}{"mode": tc.mode, "logBucket": "customer-logs", "dataBucket": "customer-data", "createBuckets": tc.create}
			if tc.mode == "external" {
				storage["endpoint"] = "http://127.0.0.1:9000"
				if tc.shared != "" {
					storage["existingSecret"] = tc.shared
				} else {
					storage["logExistingSecret"] = "log-credentials"
					storage["dataExistingSecret"] = "data-credentials"
				}
			}
			docs, output, err := render(t, map[string]interface{}{"stage": "storage", "objectStorage": storage, "dashboard": map[string]interface{}{"enabled": false}})
			if err != nil {
				t.Fatal(err, output)
			}
			var job batchv1.Job
			decodeScheduledObject(t, find(t, docs, "Job", "qa-storage-init-1"), &job)
			env := map[string]string{}
			refs := map[string]string{}
			for _, e := range job.Spec.Template.Spec.Containers[0].Env {
				env[e.Name] = e.Value
				if e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil {
					refs[e.Name] = e.ValueFrom.SecretKeyRef.Name
					env[e.Name] = "sensitive-" + e.Name
				}
			}
			script := field(find(t, docs, "ConfigMap", "qa-storage-init"), "data", "initialize.sh").(string)
			dir := t.TempDir()
			executable(t, dir, "seq", "#!/bin/sh\necho 1\n")
			executable(t, dir, "sleep", "#!/bin/sh\nexit 0\n")
			executable(t, dir, "mc", `#!/bin/sh
set -eu
side=log
if [ "$1" = alias ]; then
 [ "$5" != "$DATA_ACCESS_ID" ] || side=data
else
 for arg in "$@"; do
  case "$arg" in storage/customer-data) side=data ;; esac
 done
fi
if [ "$1" = "$FAIL_OPERATION" ] && [ "$side" = "$FAIL_SIDE" ]; then
 echo "sensitive-ACCESS_ID sensitive-ACCESS_KEY sensitive-DATA_ACCESS_ID sensitive-DATA_ACCESS_KEY $ENDPOINT" >&2
 exit 1
fi
`)
			cmd := exec.Command("sh")
			cmd.Stdin = strings.NewReader(strings.ReplaceAll(script, "/work", dir))
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "FAIL_SIDE="+tc.failSide, "FAIL_OPERATION="+tc.operation, "DATA_ACCESS_ID=unused")
			for k, v := range env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			out, runErr := cmd.CombinedOutput()
			if tc.failSide == "none" {
				if runErr != nil {
					t.Fatal(runErr, string(out))
				}
				return
			}
			if runErr == nil {
				t.Fatalf("expected failure: %s", out)
			}
			secret := refs["ACCESS_ID"]
			bucket := "customer-logs"
			if tc.failSide == "data" {
				bucket = "customer-data"
				if tc.mode == "external" {
					secret = refs["DATA_ACCESS_ID"]
				}
			}
			failureLine := ""
			for _, line := range strings.Split(string(out), "\n") {
				if strings.Contains(line, "failed") {
					failureLine = line
				}
			}
			for _, want := range []string{"Storage initialization", "side=" + tc.failSide, "bucket=" + bucket, "Secret=qa-ns/" + secret} {
				if !strings.Contains(failureLine, want) {
					t.Errorf("missing %q in %s", want, out)
				}
			}
			for _, forbidden := range []string{"sensitive-", env["ENDPOINT"]} {
				if forbidden != "" && strings.Contains(string(out), forbidden) {
					t.Errorf("diagnostics leaked %q", forbidden)
				}
			}
		})
	}
}
