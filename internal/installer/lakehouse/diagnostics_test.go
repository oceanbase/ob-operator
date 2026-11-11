package lakehouse

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func diagnosticValues() map[string]interface{} {
	return map[string]interface{}{
		"observer":       map[string]interface{}{"image": "qa/observer:fixed", "topology": []interface{}{map[string]interface{}{"replica": 1}}},
		"logService":     map[string]interface{}{"image": "qa/logservice:fixed"},
		"initialization": map[string]interface{}{"image": "qa/mc:fixed"},
		"dashboard": map[string]interface{}{"enabled": true, "image": "qa/dashboard:fixed", "inspection": map[string]interface{}{
			"obdiag": map[string]interface{}{"image": "qa/obdiag:fixed"}, "oceanbaseHelper": map[string]interface{}{"image": "qa/helper:fixed"},
		}},
	}
}

func diagnosticPod(name, image string) corev1.Pod {
	return corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:   corev1.PodSpec{Containers: []corev1.Container{{Name: "observer", Image: image}}},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Name: "observer", Image: image, Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}}}}}
}

func TestDiagnosticImageConfigurationMustMatchLock(t *testing.T) {
	dir := t.TempDir()
	images := []lockedImage{{Role: "observer", Reference: "qa/observer:fixed"}, {Role: "logservice", Reference: "qa/logservice:fixed"}, {Role: "mc", Reference: "qa/mc:fixed"}, {Role: "dashboard", Reference: "qa/dashboard:fixed"}, {Role: "obdiag", Reference: "qa/obdiag:fixed"}, {Role: "oceanbase-helper", Reference: "qa/helper:fixed"}}
	data, _ := json.Marshal(imageLock{Images: images})
	if err := os.WriteFile(filepath.Join(dir, "image-lock.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateConfiguredImages(dir, diagnosticValues()); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"obdiag", "oceanbaseHelper"} {
		v := diagnosticValues()
		setValue(v, "unlocked:latest", "dashboard", "inspection", field, "image")
		if err := validateConfiguredImages(dir, v); err == nil {
			t.Fatalf("accepted unlocked diagnostic image %s", field)
		}
		setValue(v, false, "dashboard", "enabled")
		if err := validateConfiguredImages(dir, v); err != nil {
			t.Fatalf("disabled Dashboard should not require its tools: %v", err)
		}
	}
}

func TestDiagnosticReadinessWaitsForRunningImage(t *testing.T) {
	reads, commands := 0, 0
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), " exec ") {
			commands++
			if reads < 3 {
				t.Fatal("checked diagnostic Python in old/unready Observer")
			}
			return nil, nil
		}
		reads++
		pod := diagnosticPod("ob-1", "qa/observer:fixed")
		if reads == 1 {
			pod.Status.ContainerStatuses[0].Image = "qa/observer:old"
		}
		if reads == 2 {
			pod.Status.ContainerStatuses[0].Ready = false
		}
		if reads == 3 {
			pod.Status.ContainerStatuses[0].Image = "docker.io/qa/observer:fixed"
		}
		return json.Marshal(corev1.PodList{Items: []corev1.Pod{pod}})
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := verifyObserverDiagnostics(ctx, run, "ns", "demo", diagnosticValues(), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if reads != 3 || commands != 1 {
		t.Fatalf("reads=%d commands=%d", reads, commands)
	}
}

func TestDiagnosticReadinessChecksEveryObserverAndFailsOnMissingPython(t *testing.T) {
	v := diagnosticValues()
	setValue(v, []interface{}{map[string]interface{}{"replica": 2}}, "observer", "topology")
	commands := 0
	run := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), " exec ") {
			commands++
			if strings.Contains(strings.Join(args, " "), "exec ob-2") {
				return nil, errors.New("exit status 127")
			}
			return nil, nil
		}
		return json.Marshal(corev1.PodList{Items: []corev1.Pod{diagnosticPod("ob-1", "qa/observer:fixed"), diagnosticPod("ob-2", "qa/observer:fixed")}})
	}
	err := verifyObserverDiagnostics(context.Background(), run, "ns", "demo", v, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "ob-2") || !strings.Contains(err.Error(), "Python") || strings.Contains(err.Error(), "use a verified diagnostics-capable Observer image") || commands != 2 {
		t.Fatalf("missing runtime dependency accepted: %v (%d checks)", err, commands)
	}
}

func diagnosticProbeCommand(t *testing.T) []string {
	t.Helper()
	var command []string
	run := func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), " exec ") {
			for index, arg := range args {
				if arg == "--" {
					command = append([]string(nil), args[index+1:]...)
					break
				}
			}
			return nil, nil
		}
		return json.Marshal(corev1.PodList{Items: []corev1.Pod{diagnosticPod("ob-1", "qa/observer:fixed")}})
	}
	if err := verifyObserverDiagnostics(context.Background(), run, "ns", "demo", diagnosticValues(), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	return command
}

func TestDiagnosticProbeMatchesPinnedTransport(t *testing.T) {
	command := diagnosticProbeCommand(t)
	if len(command) != 7 || command[0] != "/bin/sh" || command[1] != "-c" || command[5] != "/usr/libexec/platform-python" || command[6] != "python3" {
		t.Fatalf("diagnostic probe must try platform-python then python3 with separate arguments: %q", command)
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "lakehouse", "images", "kubernetes_log_transport.py"))
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(source), "PROBE_CODE = '''")
	if !ok {
		t.Fatal("pinned transport has no PROBE_CODE")
	}
	want, _, ok := strings.Cut(rest, "'''")
	if !ok || command[4] != want {
		t.Fatal("installer probe differs from the actual diagnostic transport")
	}
}

func TestDiagnosticProbeExercisesPythonWithoutTarAndRejectsImportOnly(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("local Python 3 is required for the isolated diagnostic probe test")
	}
	command := diagnosticProbeCommand(t)
	if len(command) != 7 {
		t.Fatalf("expected Python probe command, got %q", command)
	}
	dir := t.TempDir()
	badPython := filepath.Join(dir, "broken-python")
	if err := os.WriteFile(badPython, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	run := func(candidates []string, directory string) error {
		t.Helper()
		args := append(append([]string(nil), command[1:5]...), candidates...)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, command[0], args...)
		cmd.Dir = directory
		// Only absolute Python candidates exist; tar and gzip are absent.
		cmd.Env = []string{"PATH=" + dir}
		_, err := cmd.Output()
		return err
	}
	if err := run([]string{badPython, python}, dir); err != nil {
		t.Fatalf("Python fallback without native tar/gzip failed: %v", err)
	}
	for _, binary := range []string{"tar", "gzip"} {
		if err := os.WriteFile(filepath.Join(dir, binary), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{badPython, filepath.Join(dir, "missing-python")}, dir); err == nil {
		t.Fatal("native tar/gzip incorrectly substituted for the mandatory Python transport")
	}
	// The module imports successfully but cannot create an archive. A probe
	// that only imports the standard libraries would incorrectly pass.
	if err := os.WriteFile(filepath.Join(dir, "tarfile.py"), []byte("def open(*args, **kwargs):\n    raise RuntimeError('archive creation unavailable')\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{badPython, python}, dir); err == nil {
		t.Fatal("Python import-only capability check incorrectly passed")
	}
	if _, err := os.Stat(filepath.Join(dir, "__pycache__")); !os.IsNotExist(err) {
		t.Fatalf("diagnostic probe wrote Python bytecode: %v", err)
	}
}

func TestDiagnosticReadinessMissingPodsIsBounded(t *testing.T) {
	commands := 0
	run := func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), " exec ") {
			commands++
		}
		return []byte(`{"items":[]}`), nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := verifyObserverDiagnostics(ctx, run, "ns", "demo", diagnosticValues(), time.Millisecond); err == nil {
		t.Fatal("empty Pod list passed")
	}
	if commands != 0 {
		t.Fatal("executed in absent Pod")
	}
}
