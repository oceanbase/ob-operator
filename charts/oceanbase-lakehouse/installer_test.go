package lakehouse_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
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
		{"storage-slow", "storage,logservice,cluster", 0, []string{"MOCK_SLOW_UPGRADE=storage", "LAKEHOUSE_PROGRESS_INTERVAL=1"}},
		{"packaged-entry", "storage,logservice,cluster", 0, []string{"LAKEHOUSE_MANAGED_ACCESS=1"}},
		{"dashboard-packaged-entry-failure", "storage,logservice,cluster", 36, []string{"LAKEHOUSE_MANAGED_ACCESS=1", "MOCK_FAIL_DASHBOARD=true"}},
		{"dashboard-disabled", "storage,logservice,cluster", 0, []string{"MOCK_DASHBOARD=false"}},
		{"dashboard-failure", "storage,logservice,cluster", 36, []string{"MOCK_FAIL_DASHBOARD=true"}},
		{"dashboard-missing-values", "storage,logservice,cluster", 1, []string{"MOCK_DASHBOARD=null"}},
		{"dashboard-custom-port", "storage,logservice,cluster", 0, []string{"MOCK_DASHBOARD_PORT=8888"}},
		{"storage-failure", "storage", 21, []string{"MOCK_FAIL_UPGRADE=storage"}},
		{"logservice-api-failure", "storage,logservice", 32, []string{"MOCK_FAIL_GET=logservice"}},
		{"observer-api-failure", "storage,logservice,cluster", 33, []string{"MOCK_FAIL_GET=cluster"}},
		{"logservice-failed-status", "storage,logservice", 1, []string{"MOCK_FAILED_STATUS=logservice"}},
		{"observer-failed-status", "storage,logservice,cluster", 1, []string{"MOCK_FAILED_STATUS=cluster"}},
		{"logservice-timeout", "storage,logservice", 1, []string{"MOCK_PENDING_STATUS=logservice"}},
		{"transient-new-to-running", "storage,logservice,cluster", 0, []string{"MOCK_TRANSIENT_STATUS=logservice"}},
		{"resume-logservice", "logservice,cluster", 0, []string{"MOCK_RELEASE_STAGE=logservice"}},
		{"resume-cluster", "cluster", 0, []string{"MOCK_RELEASE_STAGE=cluster"}},
		{"retained-cluster", "cluster", 0, []string{"MOCK_RETAINED=cluster"}},
		{"retained-logservice", "logservice,cluster", 0, []string{"MOCK_RETAINED=logservice"}},
		{"retained-cluster-overrides-old-stage", "cluster", 0, []string{"MOCK_RELEASE_STAGE=storage", "MOCK_RETAINED=cluster"}},
		{"missing-crd", "", 31, []string{"MOCK_MISSING_CRD=true"}},
		{"namespace-forbidden", "", 39, []string{"MOCK_NAMESPACE=forbidden"}},
		{"retained-read-failure", "", 40, []string{"MOCK_RETAINED=unavailable"}},
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
  'get values')
    values=$(printf '{"stage":"%s","dashboard":{"enabled":%s,"service":{"port":%s}},"logService":{"resource":{"cpu":"1"}},"objectStorage":{"accessKey":"SAVED_CREDENTIAL_SENTINEL"}}\n' "${MOCK_RELEASE_STAGE:-cluster}" "${MOCK_DASHBOARD:-true}" "${MOCK_DASHBOARD_PORT:-80}")
    if [[ " $* " == *' --all '* ]]; then jq '.computedDefault="must-not-freeze"' <<< "$values"; else echo "$values"; fi ;;

  'upgrade --install')
    stage=''
    for arg in "$@"; do [[ $arg != stage=* ]] || stage=${arg#stage=}; done
    if [[ -n ${MOCK_RELEASE_STAGE:-} ]]; then
      for ((i=1; i<=$#; i++)); do
        if [[ ${!i} == -f ]]; then
          j=$((i+1)); first_values=${!j}
          [[ -f $first_values ]] || exit 41
          [[ -n $(find "$first_values" -perm 600) ]] || exit 42
          cp "$first_values" "$CASE_LOG.base-values"
          printf '%s\n' "$first_values" > "$CASE_LOG.base-path"
          break
        fi
      done
    fi
    echo 'HELM_NOTES_PRIVATE_SENTINEL'
    if [[ ${MOCK_SLOW_UPGRADE:-} == "$stage" ]]; then /bin/sleep 2; fi
    [[ ${MOCK_FAIL_UPGRADE:-} != "$stage" ]] || exit 21 ;;
  *) exit 22 ;;
esac
`)
			executable(t, dir, "kubectl", `#!/usr/bin/env bash
set -eu
printf 'kubectl %s\n' "$*" >> "$CASE_LOG"
args=()
for arg in "$@"; do case "$arg" in --request-timeout=*) ;; *) args+=("$arg");; esac; done
set -- "${args[@]}"
case "$*" in
  'get crd/'*) [[ ${MOCK_MISSING_CRD:-false} != true ]] || exit 31 ;;
  'wait crd/'*) ;;
  'get namespace qa-ns --ignore-not-found -o name')
    [[ ${MOCK_NAMESPACE:-exists} != forbidden ]] || exit 39
    if [[ ${MOCK_NAMESPACE:-exists} == exists ]]; then echo namespace/qa-ns; fi ;;
  'create namespace qa-ns') ;;
  '-n qa-ns get obcluster/qa-ob --ignore-not-found -o name')
    [[ ${MOCK_RETAINED:-none} != unavailable ]] || exit 40
    if [[ ${MOCK_RETAINED:-none} == cluster ]]; then echo obcluster/qa-ob; fi ;;
  '-n qa-ns get oblogservicecluster/qa-ls --ignore-not-found -o name')
    if [[ ${MOCK_RETAINED:-none} == logservice ]]; then echo oblogservicecluster/qa-ls; fi ;;
  '-n qa-ns get oblogservicecluster/qa-ls -o jsonpath='*|'-n qa-ns get obcluster/qa-ob -o jsonpath='*)
    kind=logservice
    code=32
    if [[ $4 == obcluster/* ]]; then kind=cluster; code=33; fi
    [[ ${MOCK_FAIL_GET:-} != "$kind" ]] || exit "$code"
    if [[ ${MOCK_FAILED_STATUS:-} == "$kind" ]]; then
      echo failed
    elif [[ ${MOCK_PENDING_STATUS:-} == "$kind" ]]; then
      echo new
    elif [[ ${MOCK_TRANSIENT_STATUS:-} == "$kind" && ! -f $CASE_LOG.$kind.seen ]]; then
      touch "$CASE_LOG.$kind.seen"
      echo new
    else
      echo running
    fi ;;
  '-n qa-ns wait pods '*) ;;
  '-n qa-ns rollout status deployment/oceanbase-dashboard-qa --timeout='*) [[ ${MOCK_FAIL_DASHBOARD:-false} != true ]] || exit 36 ;;
  '-n qa-ns get pods,jobs,obclusters,oblogserviceclusters,pvc -o json')
    echo '{"items":[{"kind":"PersistentVolumeClaim","metadata":{"name":"qa-data"},"status":{"phase":"Pending"},"spec":{"private":"DIAGNOSTIC_PRIVATE_SENTINEL"}}]}' ;;
  '-n qa-ns get events --field-selector=type=Warning -o json')
    echo '{"items":[{"type":"Warning","reason":"FailedScheduling","involvedObject":{"kind":"Pod","name":"qa-pod"},"message":"DIAGNOSTIC_PRIVATE_SENTINEL"}]}' ;;
  *) exit 35 ;;
esac
`)
			// Transient status tests need no real five-second delay. The timeout
			// case retains the actual sleep so Bash's deadline expires naturally.
			if tc.name == "transient-new-to-running" || strings.HasSuffix(tc.name, "failed-status") {
				executable(t, dir, "sleep", "#!/bin/sh\nprintf 'sleep %s\\n' \"$*\" >> \"$CASE_LOG\"\n")
			}
			timeout := "9"
			if tc.name == "logservice-timeout" {
				timeout = "1"
			}
			cmd := exec.Command("bash", "install.sh", "qa", "qa-ns", "-f", "/tmp/values with spaces.yaml", "--set-string", "objectStorage.region=us-east-1")
			cmd.Env = append(os.Environ(), "LC_ALL=C", "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "CASE_LOG="+log, "LAKEHOUSE_TIMEOUT="+timeout, "TMPDIR="+dir)
			cmd.Env = append(cmd.Env, tc.env...)
			output, err := cmd.CombinedOutput()
			if got := exitCode(err); got != tc.code {
				t.Fatalf("exit %d, want %d: %s", got, tc.code, output)
			}
			calls, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "fresh" || tc.name == "resume-cluster" || tc.name == "packaged-entry" {
				t.Logf("installer progress transcript:\n%s", output)
			}
			if strings.Contains(string(output), "HELM_NOTES_PRIVATE_SENTINEL") || strings.Contains(string(output), "DIAGNOSTIC_PRIVATE_SENTINEL") || strings.Contains(string(output), "SAVED_CREDENTIAL_SENTINEL") {
				t.Fatalf("installer leaked private raw output: %s", output)
			}
			logs, logErr := filepath.Glob(filepath.Join(dir, "oceanbase-lakehouse-qa.*", "install.log"))
			if logErr != nil || len(logs) != 1 {
				t.Fatalf("expected one private installation log: %v %v", logs, logErr)
			}
			if info, err := os.Stat(logs[0]); err != nil || info.Mode().Perm() != 0600 {
				t.Fatal("installation log must be mode0600", err)
			}
			if info, err := os.Stat(filepath.Dir(logs[0])); err != nil || info.Mode().Perm() != 0700 {
				t.Fatal("log directory must be mode0700", err)
			}
			if tc.code == 0 && !strings.Contains(string(output), "[stage]") {
				t.Fatalf("installer must announce stages: %s", output)
			}
			// Executable progress contract: distinguish submitted Helm resources
			// from completed runtime gates and do not repeat success on fresh runs.
			if tc.code == 0 {
				managed := tc.name == "packaged-entry"
				if managed {
					if strings.Contains(string(output), "/6]") {
						t.Fatalf("managed entry must not start a second progress count: %s", output)
					}
				}
				for _, step := range []string{"[1/6] Preflight", "[4/6] OceanBase", "[5/6] Observer Pods"} {
					if managed {
						step = strings.SplitN(step, "] ", 2)[1]
					}
					if strings.Count(string(output), "[done] "+step) != 1 {
						t.Fatalf("missing or duplicate completion for %s: %s", step, output)
					}
				}
				for _, kind := range []string{"oblogservicecluster/qa-ls", "obcluster/qa-ob"} {
					if strings.Count(string(output), kind+" is running") != 1 {
						t.Fatalf("runtime success must be reported once for %s: %s", kind, output)
					}
				}
				if tc.name == "fresh" || tc.name == "packaged-entry" || tc.name == "storage-slow" {
					for _, step := range []string{"[2/6] Storage", "[3/6] LogService", "[6/6] Dashboard"} {
						if managed {
							step = strings.SplitN(step, "] ", 2)[1]
						}
						if strings.Count(string(output), "[done] "+step) != 1 {
							t.Fatalf("missing or duplicate completion for %s: %s", step, output)
						}
					}
				}
				if strings.Count(string(output), "Lakehouse ready:") != 1 || strings.Contains(string(output), "For a complete, ordered installation run:") {
					t.Fatalf("installer repeated a final result or install instruction: %s", output)
				}
			}
			if tc.name == "storage-failure" && strings.Contains(string(output), "[done] [2/6]") {
				t.Fatal("failed storage must not be marked complete")
			}
			if tc.name == "storage-slow" && !strings.Contains(string(output), "[wait] Helm storage stage:") {
				t.Fatalf("quiet Helm wait must retain a bounded progress heartbeat: %s", output)
			}
			if tc.name == "fresh" && strings.Count(string(calls), "get oblogservicecluster/qa-ls -o jsonpath=") != 2 {
				t.Fatal("quiet final LogService verification must not be skipped")
			}
			if strings.HasPrefix(tc.name, "resume-") || tc.name == "retained-cluster-overrides-old-stage" {
				basePath, readErr := os.ReadFile(log + ".base-path")
				if readErr != nil {
					t.Fatal("upgrade did not load existing user values:", readErr)
				}
				baseData, err := os.ReadFile(log + ".base-values")
				if err != nil {
					t.Fatal(err)
				}
				var base map[string]interface{}
				if err := json.Unmarshal(baseData, &base); err != nil {
					t.Fatal(err)
				}
				if field(base, "logService", "resource", "cpu") != "1" || field(base, "objectStorage", "accessKey") != "SAVED_CREDENTIAL_SENTINEL" || base["computedDefault"] != nil {
					t.Fatal("must preserve all user-supplied values and omit computed defaults")
				}
				if _, statErr := os.Stat(strings.TrimSpace(string(basePath))); !os.IsNotExist(statErr) {
					t.Fatal("temporary credentials-containing values file must be cleaned")
				}
				for _, line := range strings.Split(string(calls), "\n") {
					if strings.HasPrefix(line, "helm upgrade ") && strings.Index(line, strings.TrimSpace(string(basePath))) > strings.Index(line, "/tmp/values with spaces.yaml") {
						t.Fatal("user values must override saved values")
					}
				}
				if strings.Contains(string(calls), "--reuse-values") {
					t.Fatal("do not freeze old defaults")
				}
				if !strings.Contains(string(calls), "helm get values qa -n qa-ns -o json") {
					t.Fatal("saved base must use user-supplied values, not --all")
				}
			}

			var stages []string
			for _, line := range strings.Split(string(calls), "\n") {
				if strings.HasPrefix(line, "helm upgrade ") {
					parts := strings.SplitN(line, "--set-string stage=", 2)
					if len(parts) != 2 || !strings.Contains(line, "--wait --wait-for-jobs --timeout "+timeout+"s") {
						t.Fatalf("missing bounded readiness gate: %s", line)
					}
					stages = append(stages, strings.Fields(parts[1])[0])
				}
				if strings.Contains(line, " -o jsonpath=") {
					parts := strings.Split(line, "--request-timeout=")
					if len(parts) != 2 {
						t.Fatalf("status request is not bounded: %s", line)
					}
					seconds, err := strconv.Atoi(strings.TrimSuffix(parts[1], "s"))
					maximum, _ := strconv.Atoi(timeout)
					if err != nil || seconds < 1 || seconds > maximum {
						t.Fatalf("status request exceeds remaining installation timeout: %s", line)
					}
				}
			}
			if got := strings.Join(stages, ","); got != tc.stages {
				t.Fatalf("stages %q, want %q\n%s", got, tc.stages, calls)
			}
			if tc.name == "namespace-forbidden" && strings.Contains(string(calls), "create namespace") {
				t.Fatal("API failure must not be treated as NotFound")
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
			dashboardWait := "rollout status deployment/oceanbase-dashboard-qa --timeout=9s"
			if tc.name == "packaged-entry" {
				for _, fragment := range []string{"port-forward", "get secret", "Database root password", "no public", "Dashboard ready:"} {
					if strings.Contains(string(output), fragment) {
						t.Fatalf("managed entry must leave access instructions to the parent: %s", output)
					}
				}
				for _, gate := range []string{dashboardWait, "get oblogservicecluster/qa-ls -o jsonpath=", "get obcluster/qa-ob -o jsonpath=", "wait pods -l ref-obcluster=qa-ob"} {
					if !strings.Contains(string(calls), gate) {
						t.Fatalf("managed entry skipped readiness gate %q: %s", gate, calls)
					}
				}
			}
			if tc.code == 0 && tc.name != "dashboard-disabled" && tc.name != "packaged-entry" {
				if !strings.Contains(string(calls), dashboardWait) || !strings.Contains(string(output), "Dashboard ready:") ||
					!strings.Contains(string(output), "get secret qa-dashboard-bootstrap") {
					t.Fatalf("success requires Dashboard readiness and access instructions: %s\n%s", output, calls)
				}
				port := "80"
				if tc.name == "dashboard-custom-port" {
					port = "8888"
				}
				if !strings.Contains(string(output), "port-forward svc/oceanbase-dashboard-qa 8080:"+port) {
					t.Fatalf("incorrect access command: %s", output)
				}
			}
			if tc.name == "dashboard-disabled" && (strings.Contains(string(calls), dashboardWait) ||
				strings.Contains(string(output), "Dashboard ready:") ||
				!strings.Contains(string(output), "Dashboard is disabled")) {
				t.Fatalf("disabled Dashboard must not be awaited or advertised: %s\n%s", output, calls)
			}
			if tc.code != 0 && (!strings.Contains(string(output), "qa-data") || !strings.Contains(string(output), "FailedScheduling")) {
				t.Fatalf("failure must summarize PVC state and warning reasons: %s", output)
			}
			if tc.code != 0 && strings.Contains(string(output), "Lakehouse ready:") {
				t.Fatalf("failed installation must not claim readiness: %s", output)
			}
			if strings.HasSuffix(tc.name, "failed-status") && (!strings.Contains(string(output), "entered failed status") || strings.Contains(string(calls), "sleep ")) {
				t.Fatalf("failed CR status must stop immediately with diagnostics: %s\n%s", output, calls)
			}
			if tc.name == "logservice-timeout" && !strings.Contains(string(output), "Timed out waiting for oblogservicecluster/qa-ls") {
				t.Fatalf("pending CR must fail with a bounded timeout: %s", output)
			}
			if tc.name == "transient-new-to-running" && !strings.Contains(string(calls), "sleep ") {
				t.Fatalf("transient pending status must be polled before success: %s", calls)
			}
			if tc.code != 0 && !strings.HasPrefix(tc.name, "dashboard-") && strings.Contains(string(calls), "wait pods -l ref-obcluster=") {
				t.Fatalf("installer waited for Pods after a failed prerequisite: %s", calls)
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
		{"unreachable", "external", "false", "alias", false, false, false},
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
  alias|ready|mb|stat|admin)
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
				"DATA_ACCESS_ID=data-user", "DATA_ACCESS_KEY=data-secret",
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

// Hanging client commands must be bounded even when the executable ignores
// client-side timeout flags. These fake tools never reach a cluster.
func TestInstallerBoundsReadsAndDiagnostics(t *testing.T) {
	for _, scenario := range []string{"crd", "helm-list", "diagnostics"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			executable(t, dir, "kubectl", `#!/usr/bin/env bash
set -eu
case "$*" in
  'get crd/'*) [[ $SCENARIO != crd ]] || exec sleep 20 ;;
  'wait crd/'*) ;;
  'get namespace '*) echo namespace/qa-ns ;;
  '-n qa-ns get obcluster/'*|'-n qa-ns get oblogservicecluster/'*) ;;
  '-n qa-ns get pods,'*|'-n qa-ns get events '*) exec sleep 20 ;;
  *) exit 44 ;;
esac
`)
			executable(t, dir, "helm", `#!/usr/bin/env bash
set -eu
case "$1" in
 list) [[ $SCENARIO != helm-list ]] || exec sleep 20; echo '[]' ;;
 upgrade) echo 'PRIVATE_FAILURE_SENTINEL'; exit 47 ;;
 *) exit 48 ;;
esac
`)
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", "install.sh", "qa", "qa-ns")
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "TMPDIR="+dir,
				"SCENARIO="+scenario, "LAKEHOUSE_TIMEOUT=2", "LAKEHOUSE_REQUEST_TIMEOUT=1", "LAKEHOUSE_DIAGNOSTIC_TIMEOUT=1", "LAKEHOUSE_PROGRESS_INTERVAL=1")
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal("unbounded command or diagnostic", string(out))
			}
			want := 124
			if scenario == "diagnostics" {
				want = 47
			}
			if exitCode(err) != want {
				t.Fatalf("exit=%d want=%d: %s", exitCode(err), want, out)
			}
			if strings.Contains(string(out), "PRIVATE_FAILURE_SENTINEL") {
				t.Fatal("leaked Helm error")
			}
			if !strings.Contains(string(out), "data has NOT been automatically removed") {
				t.Fatal("missing recovery advice")
			}
		})
	}
}
