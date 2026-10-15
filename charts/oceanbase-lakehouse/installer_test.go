package lakehouse_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func executable(t *testing.T, directory, name, script string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}

// Execute the real installer, replacing only infrastructure clients. No test
// contacts or modifies a Kubernetes cluster or existing Helm installation.
func TestInstallerSequencingResumeAndFailures(t *testing.T) {
	for _, tc := range []struct {
		name, stages string
		code         int
		env          []string
	}{
		{"fresh", "storage,logservice,cluster", 0, []string{"MOCK_NAMESPACE=absent"}},
		{"storage-failure", "storage", 21, []string{"MOCK_FAIL_UPGRADE=storage"}},
		{"logservice-readiness-failure", "storage,logservice", 32, []string{"MOCK_FAIL_WAIT=logservice"}},
		{"observer-readiness-failure", "storage,logservice,cluster", 33, []string{"MOCK_FAIL_WAIT=cluster"}},
		{"resume-logservice", "logservice,cluster", 0, []string{"MOCK_RELEASE_STAGE=logservice"}},
		{"resume-cluster", "cluster", 0, []string{"MOCK_RELEASE_STAGE=cluster"}},
		{"retained-cluster", "cluster", 0, []string{"MOCK_RETAINED=cluster"}},
		{"retained-logservice", "logservice,cluster", 0, []string{"MOCK_RETAINED=logservice"}},
		{"retained-cluster-overrides-old-stage", "cluster", 0, []string{"MOCK_RELEASE_STAGE=storage", "MOCK_RETAINED=cluster"}},
		{"missing-crd", "", 31, []string{"MOCK_MISSING_CRD=true"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			log := filepath.Join(dir, "calls")
			executable(t, dir, "helm", `#!/usr/bin/env bash
set -eu
printf 'helm %s\n' "$*" >> "$CASE_LOG"
case "$1 $2" in
  'list -n')
    if [[ -n ${MOCK_RELEASE_STAGE:-} ]]; then echo '[{"name":"qa"}]'; else echo '[]'; fi ;;
  'get values') printf '{"stage":"%s"}\n' "$MOCK_RELEASE_STAGE" ;;
  'upgrade --install')
    stage=''
    for arg in "$@"; do [[ $arg != stage=* ]] || stage=${arg#stage=}; done
    [[ ${MOCK_FAIL_UPGRADE:-} != "$stage" ]] || exit 21 ;;
  *) exit 22 ;;
esac
`)
			executable(t, dir, "kubectl", `#!/usr/bin/env bash
set -eu
printf 'kubectl %s\n' "$*" >> "$CASE_LOG"
case "$*" in
  'get crd/'*) [[ ${MOCK_MISSING_CRD:-false} != true ]] || exit 31 ;;
  'wait crd/'*) ;;
  'get namespace qa-ns') [[ ${MOCK_NAMESPACE:-exists} == exists ]] ;;
  'create namespace qa-ns') ;;
  '-n qa-ns get obcluster/qa-ob') [[ ${MOCK_RETAINED:-none} == cluster ]] ;;
  '-n qa-ns get oblogservicecluster/qa-ls') [[ ${MOCK_RETAINED:-none} == logservice ]] ;;
  '-n qa-ns wait oblogservicecluster/qa-ls '*) [[ ${MOCK_FAIL_WAIT:-} != logservice ]] || exit 32 ;;
  '-n qa-ns wait obcluster/qa-ob '*) [[ ${MOCK_FAIL_WAIT:-} != cluster ]] || exit 33 ;;
  '-n qa-ns wait pods '*) ;;
  '-n qa-ns get pods,jobs,obclusters,oblogserviceclusters') ;;
  *) exit 35 ;;
esac
`)
			cmd := exec.Command("bash", "install.sh", "qa", "qa-ns", "-f", "/tmp/values with spaces.yaml", "--set-string", "objectStorage.region=us-east-1")
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "CASE_LOG="+log, "LAKEHOUSE_TIMEOUT=9")
			cmd.Env = append(cmd.Env, tc.env...)
			output, err := cmd.CombinedOutput()
			if got := exitCode(err); got != tc.code {
				t.Fatalf("exit %d, want %d: %s", got, tc.code, output)
			}
			calls, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			var stages []string
			for _, line := range strings.Split(string(calls), "\n") {
				if strings.HasPrefix(line, "helm upgrade ") {
					parts := strings.SplitN(line, "--set-string stage=", 2)
					if len(parts) != 2 || !strings.Contains(line, "--wait --wait-for-jobs --timeout 9s") {
						t.Fatalf("missing bounded readiness gate: %s", line)
					}
					stages = append(stages, strings.Fields(parts[1])[0])
				}
			}
			if got := strings.Join(stages, ","); got != tc.stages {
				t.Fatalf("stages %q, want %q\n%s", got, tc.stages, calls)
			}
			if strings.Contains(string(calls), "helm rollback") || strings.Contains(string(calls), "helm uninstall") || strings.Contains(string(calls), " delete ") {
				t.Fatalf("installer issued destructive recovery: %s", calls)
			}
			if tc.code != 0 && tc.name != "missing-crd" && !strings.Contains(string(output), "data has NOT been automatically removed") {
				t.Fatalf("failure in a function must run ERR diagnostics: %s", output)
			}
			if tc.code == 0 && (!strings.Contains(string(output), "Lakehouse ready:") || !strings.Contains(string(calls), "wait pods -l ref-obcluster=qa-ob --for=condition=Ready --timeout=9s")) {
				t.Fatalf("success reported before Observer Pod readiness: %s\n%s", output, calls)
			}
		})
	}
}

// Run the rendered script with a local fake mc. Rewriting its private work
// directory avoids depending on /work or sharing files between test processes.
func TestInitializationExternalSafetyAndFailureRedaction(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	script := field(find(t, docs, "ConfigMap", "qa-storage-init"), "data", "initialize.sh").(string)
	for _, tc := range []struct {
		name, mode, create, fail string
		success, buckets, admin  bool
	}{
		{"bundled", "bundled", "false", "", true, true, true},
		{"external-readonly", "external", "false", "", true, false, false},
		{"external-create", "external", "true", "", true, true, false},
		{"unreachable", "external", "false", "ready", false, false, false},
		{"missing-bucket", "external", "false", "stat", false, false, false},
		{"bucket-creation-failed", "external", "true", "mb", false, true, false},
		{"admin-failed", "bundled", "false", "admin", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			log := filepath.Join(dir, "mc-calls")
			executable(t, dir, "mc", `#!/bin/sh
set -eu
case "$1 $2" in
  'alias set') echo 'mc alias set [credentials redacted]' >> "$INIT_LOG" ;;
  *) printf 'mc %s\n' "$*" >> "$INIT_LOG" ;;
esac
case "$1" in
  ready|mb|stat|admin)
    if [ "$MC_FAIL" = "$1" ]; then echo 'test-secret-value' >&2; exit 1; fi ;;
esac
`)
			executable(t, dir, "seq", "#!/bin/sh\necho 1\n")
			executable(t, dir, "sleep", "#!/bin/sh\nexit 0\n")
			cmd := exec.Command("sh")
			cmd.Stdin = strings.NewReader(strings.ReplaceAll(script, "/work", dir))
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "INIT_LOG="+log,
				"MODE="+tc.mode, "CREATE_BUCKETS="+tc.create, "MC_FAIL="+tc.fail, "ENDPOINT=http://minio:9000",
				"ACCESS_ID=root-user", "ACCESS_KEY=test-secret-value", "BUSINESS_ID=business-user", "BUSINESS_KEY=business-secret",
				"REGION=us-east-1", "LOG_BUCKET=qa-logs", "DATA_BUCKET=qa-data", "POLICY_NAME=qa-buckets")
			result, err := cmd.CombinedOutput()
			if (err == nil) != tc.success {
				t.Fatalf("unexpected success=%v: %v %s", err == nil, err, result)
			}
			calls, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(calls), "mc mb ") != tc.buckets || strings.Contains(string(calls), "mc admin ") != tc.admin {
				t.Fatalf("unexpected storage mutation: %s", calls)
			}
			if strings.Contains(string(result), "test-secret-value") || strings.Contains(string(result), "business-secret") {
				t.Fatalf("client error exposed a credential: %s", result)
			}
			for _, destructive := range []string{"mc rm ", "mc rb ", "mc mirror "} {
				if strings.Contains(string(calls), destructive) {
					t.Fatalf("unexpected data deletion: %s", calls)
				}
			}
		})
	}
}
