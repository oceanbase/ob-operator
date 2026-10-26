package lakehouse_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func runShellProgress(t *testing.T, podSnapshots []string, events string, env ...string) (string, string, int) {
	t.Helper()
	dir := t.TempDir()
	for i, data := range podSnapshots {
		path := filepath.Join(dir, "pods-"+strconv.Itoa(i+1)+".json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if events == "" {
		events = `{"items":[]}`
	}
	if err := os.WriteFile(filepath.Join(dir, "events.json"), []byte(events), 0600); err != nil {
		t.Fatal(err)
	}
	// Advance only the installer's poll clock; deadline watchdogs still use real
	// /bin/sleep. This exercises heartbeat/caching behavior without a 45s test.
	if err := os.WriteFile(filepath.Join(dir, "clock.sh"), []byte("sleep() { SECONDS=$((SECONDS + $1)); }\n"), 0600); err != nil {
		t.Fatal(err)
	}
	executable(t, dir, "helm", `#!/usr/bin/env bash
set -eu
case "$1 $2" in
  'list -n') echo '[]' ;;
  'get values') echo '{"dashboard":{"enabled":false,"service":{"port":80}}}' ;;
  'upgrade --install') ;;
  *) exit 90 ;;
esac
`)
	executable(t, dir, "kubectl", `#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >> "$CASE_DIR/calls"
args=()
for arg in "$@"; do case "$arg" in --request-timeout=*) ;; *) args+=("$arg");; esac; done
set -- "${args[@]}"
case "$*" in
  '-n qa-ns get oblogservicecluster/qa-ls -o jsonpath='*|'-n qa-ns get obcluster/qa-ob -o jsonpath='*)
    if [[ "$4" != "${MOCK_PROGRESS_RESOURCE:-oblogservicecluster/qa-ls}" ]]; then echo running; exit; fi
    count=$(cat "$CASE_DIR/status-count" 2>/dev/null || echo 0)
    count=$((count + 1)); echo "$count" > "$CASE_DIR/status-count"
    if (( count <= ${MOCK_PENDING_READS:-4} )); then echo new; else echo "${MOCK_FINAL_STATUS:-running}"; fi ;;
  '-n qa-ns get pods -l '*'-o json')
    [[ ${MOCK_POD_FAILURE:-0} != 1 ]] || { echo 'PRIVATE_API_FAILURE' >&2; exit 42; }
    count=$(cat "$CASE_DIR/pod-count" 2>/dev/null || echo 0)
    count=$((count + 1)); echo "$count" > "$CASE_DIR/pod-count"
    [[ ${MOCK_POD_FAILURE_AFTER:-0} != "$count" ]] || { echo 'PRIVATE_API_FAILURE' >&2; exit 42; }
    path="$CASE_DIR/pods-$count.json"
    if [[ ! -f $path ]]; then path="$CASE_DIR/pods-1.json"; fi
    cat "$path" ;;
  '-n qa-ns get events --field-selector=type=Warning -o json')
    [[ ${MOCK_EVENT_FAILURE:-0} != 1 ]] || { echo 'PRIVATE_EVENT_FAILURE' >&2; exit 43; }
    cat "$CASE_DIR/events.json" ;;
  '-n qa-ns get pods,jobs,obclusters,oblogserviceclusters,pvc -o json') echo '{"items":[]}' ;;
  'get crd/'*|'wait crd/'*|'get namespace '*|'create namespace '*|'-n qa-ns get '*'-o name'|'-n qa-ns wait pods '*) ;;
  *) exit 91 ;;
esac
`)
	cmd := exec.Command("bash", "install.sh", "qa", "qa-ns")
	cmd.Env = append(os.Environ(), "LC_ALL=C", "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "CASE_DIR="+dir, "TMPDIR="+dir, "BASH_ENV="+filepath.Join(dir, "clock.sh"), "LAKEHOUSE_TIMEOUT=100", "LAKEHOUSE_MANAGED_ACCESS=1")
	cmd.Env = append(cmd.Env, env...)
	output, err := cmd.CombinedOutput()
	calls, readErr := os.ReadFile(filepath.Join(dir, "calls"))
	if readErr != nil {
		t.Fatal(readErr, string(output))
	}
	return string(output), string(calls), exitCode(err)
}

const progressCreatingPods = `{"items":[
  {"metadata":{"name":"pod-a","uid":"uid-a"},"spec":{"nodeName":"worker","volumes":[{"persistentVolumeClaim":{"claimName":"claim-a"}}],"env":"PRIVATE_POD_SPEC"},"status":{"conditions":[{"type":"Ready","status":"False"}],"containerStatuses":[{"state":{"waiting":{"reason":"ContainerCreating","message":"PRIVATE_WAIT_MESSAGE"}}}]}},
  {"metadata":{"name":"pod-b","uid":"uid-b"},"spec":{},"status":{"conditions":[{"type":"PodScheduled","status":"False","reason":"Unschedulable","message":"PRIVATE_SCHEDULER_MESSAGE"}],"containerStatuses":[{"state":{"waiting":{"reason":"PRIVATE_UNKNOWN_REASON"}}}]}}
]}`

const progressReadyPods = `{"items":[
  {"metadata":{"name":"pod-a","uid":"uid-a"},"spec":{"nodeName":"worker"},"status":{"conditions":[{"type":"Ready","status":"True"}]}},
  {"metadata":{"name":"pod-b","uid":"uid-b"},"spec":{"nodeName":"worker"},"status":{"conditions":[{"type":"Ready","status":"True"}]}}
]}`

func TestShellProgressUsesScopedPodSnapshotsAndKnownReasons(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	var events []map[string]interface{}
	for _, event := range []struct{ kind, name, uid, reason, stamp string }{
		{"Pod", "pod-a", "uid-a", "FailedMount", now},
		{"PersistentVolumeClaim", "claim-a", "claim-uid", "FailedBinding", now},
		{"Pod", "pod-a", "old-uid", "FailedAttachVolume", now},
		{"Pod", "pod-b", "uid-b", "FailedScheduling", "2000-01-01T00:00:00Z"},
		{"Pod", "pod-a", "uid-a", "PRIVATE_EVENT_REASON", now},
	} {
		events = append(events, map[string]interface{}{"involvedObject": map[string]string{"kind": event.kind, "name": event.name, "uid": event.uid}, "reason": event.reason, "lastTimestamp": event.stamp, "message": "PRIVATE_EVENT_MESSAGE"})
	}
	eventJSON, _ := json.Marshal(map[string]interface{}{"items": events})
	out, calls, code := runShellProgress(t, []string{progressCreatingPods, progressReadyPods}, string(eventJSON))
	if code != 0 {
		t.Fatalf("installer failed: %d\n%s", code, out)
	}
	for _, want := range []string{"status=new; podsReady=0/2; scheduled=1/2; waiting=ContainerCreating; scheduling=Unschedulable", "recentWarnings=FailedBinding,FailedMount", "podsReady=2/2; scheduled=2/2; waiting=ClusterInitialization", "resource status checked; Pod snapshot age=0s", "phase elapsed=", "total elapsed="} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "PRIVATE_") || strings.Contains(out, "FailedAttachVolume") || strings.Contains(out, "FailedScheduling") {
		t.Fatal("raw, unrelated, or stale event content leaked", out)
	}
	if !strings.Contains(calls, "get pods -l ref-oblogservicecluster=qa-ls -o json --request-timeout=1s") {
		t.Fatal("LogService progress selector or timeout incorrect", calls)
	}
	if strings.Count(out, "[wait] oblogservicecluster/qa-ls:") != 2 {
		t.Fatal("non-TTY should report only the two actual Pod status changes", out)
	}
	if strings.Count(out, "[warning] Inspect Pods: kubectl -n qa-ns describe pods -l ref-oblogservicecluster=qa-ls") != 1 || !strings.Contains(out, "storage is blocked") {
		t.Fatal("the first known issue needs one actionable warning", out)
	}
}

func TestShellProgressPollingAndFailureFallback(t *testing.T) {
	for _, tc := range []struct {
		name  string
		env   []string
		want  string
		waits int
	}{
		{"plain-heartbeat", []string{"MOCK_PENDING_READS=9"}, "Pod snapshot age=0s", 2},
		{"live-checks", []string{"MOCK_PENDING_READS=3", "LAKEHOUSE_LIVE_PROGRESS=1"}, "Pod snapshot age=", 3},
		{"pod-failure", []string{"MOCK_POD_FAILURE=1"}, "Pod details unavailable", 1},
		{"pod-failure-after-snapshot", []string{"MOCK_POD_FAILURE_AFTER=2"}, "Pod details unavailable", 2},
		{"event-failure", []string{"MOCK_EVENT_FAILURE=1"}, "podsReady=0/2", 1},
		{"observer-selector", []string{"MOCK_PROGRESS_RESOURCE=obcluster/qa-ob"}, "[wait] obcluster/qa-ob:", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, calls, code := runShellProgress(t, []string{progressCreatingPods}, "", tc.env...)
			if code != 0 || !strings.Contains(out, tc.want) || strings.Contains(out, "PRIVATE_") {
				t.Fatalf("unexpected fallback/result: %d\n%s", code, out)
			}
			if got := strings.Count(out, "[wait] ob"); got != tc.waits {
				t.Errorf("wait lines=%d want=%d:\n%s", got, tc.waits, out)
			}
			if tc.name == "observer-selector" && !strings.Contains(calls, "get pods -l ref-obcluster=qa-ob") {
				t.Error("Observer progress must use the actual Operator label", calls)
			}
			if tc.name == "live-checks" && strings.Count(calls, "get pods -l ") != 2 {
				t.Error("five-second live checks must reuse a ten-second Pod snapshot", calls)
			}
			if tc.name != "pod-failure" && strings.Count(out, "[warning] Inspect Pods:") != 1 {
				t.Error("a persistent warning must not repeat on heartbeats or detail failures", out)
			}
		})
	}
}

func TestShellProgressWarningsRecurOnlyAfterRecovery(t *testing.T) {
	out, _, code := runShellProgress(t, []string{progressCreatingPods, progressReadyPods, progressCreatingPods}, "", "MOCK_PENDING_READS=6", "LAKEHOUSE_LIVE_PROGRESS=1")
	if code != 0 || strings.Count(out, "[warning] Inspect Pods:") != 2 {
		t.Fatalf("warning must be reissued once when the issue returns after recovery: %d\n%s", code, out)
	}
	for _, tc := range []struct{ reason, hint string }{
		{"ImagePullBackOff", "image startup is blocked"},
		{"CrashLoopBackOff", "container startup is failing"},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			pods := `{"items":[{"metadata":{"uid":"pod-uid"},"spec":{"nodeName":"worker"},"status":{"containerStatuses":[{"state":{"waiting":{"reason":"` + tc.reason + `"}}}]}}]}`
			out, _, code := runShellProgress(t, []string{pods}, "", "MOCK_PENDING_READS=1")
			if code != 0 || !strings.Contains(out, tc.hint) || strings.Count(out, "[warning] Inspect Pods:") != 1 {
				t.Fatalf("missing actionable container warning: %d\n%s", code, out)
			}
		})
	}
}

func TestShellProgressDoesNotHideResourceFailureOrExceedDeadline(t *testing.T) {
	out, _, code := runShellProgress(t, []string{progressCreatingPods}, "", "MOCK_FINAL_STATUS=failed", "MOCK_POD_FAILURE=1")
	if code != 1 || !strings.Contains(out, "entered failed status") || strings.Contains(out, "Lakehouse ready:") {
		t.Fatalf("optional progress hid the real resource failure: %d\n%s", code, out)
	}
	out, calls, code := runShellProgress(t, []string{progressCreatingPods}, "", "LAKEHOUSE_TIMEOUT=1")
	if code != 1 || !strings.Contains(out, "Timed out waiting") || strings.Contains(calls, "get pods -l ") {
		t.Fatalf("optional progress must skip insufficient remaining budget: %d\n%s\n%s", code, out, calls)
	}
}
