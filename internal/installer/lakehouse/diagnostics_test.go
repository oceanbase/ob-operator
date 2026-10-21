package lakehouse

import (
	"context"
	"encoding/json"
	"errors"
	"os"
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
				t.Fatal("checked tar in old/unready Observer")
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

func TestDiagnosticReadinessChecksEveryObserverAndFailsOnMissingTar(t *testing.T) {
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
	if err == nil || !strings.Contains(err.Error(), "ob-2") || !strings.Contains(err.Error(), "tar/gzip") || commands != 2 {
		t.Fatalf("missing runtime dependency accepted: %v (%d checks)", err, commands)
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
