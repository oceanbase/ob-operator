package lakehouse_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestInitializationIdentityIsIndependentAndNonRoot(t *testing.T) {
	for _, mode := range []string{"bundled", "external"} {
		t.Run(mode, func(t *testing.T) {
			storage := map[string]interface{}{"mode": mode}
			if mode == "external" {
				storage["endpoint"] = "http://192.0.2.10:9000"
				storage["existingSecret"] = "external-storage"
			}
			docs, output, err := render(t, map[string]interface{}{
				"objectStorage":  storage,
				"initialization": map[string]interface{}{"runAsUser": 10001, "runAsGroup": 10002, "fsGroup": 10003},
			})
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			var initialize batchv1.Job
			decodeScheduledObject(t, find(t, docs, "Job", "qa-storage-init-1"), &initialize)
			pod := initialize.Spec.Template.Spec
			security := pod.SecurityContext
			if security == nil || security.RunAsUser == nil || *security.RunAsUser != 10001 ||
				security.RunAsGroup == nil || *security.RunAsGroup != 10002 || security.FSGroup == nil || *security.FSGroup != 10003 {
				t.Fatalf("initialization identity was not independently propagated: %+v", security)
			}
			if security.RunAsNonRoot == nil || !*security.RunAsNonRoot || security.SeccompProfile == nil || security.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
				t.Fatal("custom initialization identity must preserve non-root and RuntimeDefault")
			}
			if pod.AutomountServiceAccountToken == nil || *pod.AutomountServiceAccountToken || len(pod.InitContainers) != 0 {
				t.Fatal("custom initialization identity must not introduce token mounts or privileged init containers")
			}
			container := pod.Containers[0].SecurityContext
			if container == nil || container.AllowPrivilegeEscalation == nil || *container.AllowPrivilegeEscalation ||
				container.ReadOnlyRootFilesystem == nil || !*container.ReadOnlyRootFilesystem ||
				(container.Privileged != nil && *container.Privileged) ||
				container.Capabilities == nil || len(container.Capabilities.Add) != 0 ||
				len(container.Capabilities.Drop) != 1 || container.Capabilities.Drop[0] != "ALL" {
				t.Fatal("custom initialization identity must not elevate container privileges")
			}
			emptyDirs := 0
			for _, volume := range pod.Volumes {
				if volume.EmptyDir != nil {
					emptyDirs++
				}
				if volume.HostPath != nil || volume.PersistentVolumeClaim != nil {
					t.Fatal("initialization must not acquire a hostPath or PVC mount")
				}
			}
			if emptyDirs != 2 {
				t.Fatal("initialization must retain its independent work and tmp emptyDirs")
			}
			if mode == "bundled" {
				var minio appsv1.StatefulSet
				decodeScheduledObject(t, find(t, docs, "StatefulSet", "qa-minio"), &minio)
				security := minio.Spec.Template.Spec.SecurityContext
				if security == nil || security.RunAsUser == nil || *security.RunAsUser != 1000 ||
					security.RunAsGroup == nil || *security.RunAsGroup != 1000 || security.FSGroup == nil || *security.FSGroup != 1000 {
					t.Fatal("changing initialization identity must not change MinIO's independent default identity")
				}
			}
		})
	}
}

func TestInitializationRejectsRootOrOverflowIdentity(t *testing.T) {
	for _, key := range []string{"runAsUser", "runAsGroup", "fsGroup"} {
		for _, value := range []interface{}{0, -1, 1.5, "10001", true, int64(2147483648), int64(4294967296)} {
			if _, _, err := render(t, map[string]interface{}{"initialization": map[string]interface{}{key: value}}); err == nil {
				t.Fatalf("invalid initialization %s=%v accepted", key, value)
			}
		}
	}
}

func TestInitializationAcceptsIdentityBoundaries(t *testing.T) {
	for _, id := range []int{1, 2147483647} {
		if _, output, err := render(t, map[string]interface{}{"initialization": map[string]interface{}{"runAsUser": id, "runAsGroup": id, "fsGroup": id}}); err != nil {
			t.Fatalf("valid non-root identity boundary %d rejected: %v %s", id, err, output)
		}
	}
}

func TestInitializationChecksPrivateDirectoriesBeforeMC(t *testing.T) {
	docs, output, err := render(t, map[string]interface{}{})
	if err != nil {
		t.Fatalf("render: %v\n%s", err, output)
	}
	script := field(find(t, docs, "ConfigMap", "qa-storage-init"), "data", "initialize.sh").(string)
	for _, blocked := range []string{"work", "tmp"} {
		t.Run(blocked, func(t *testing.T) {
			dir := t.TempDir()
			work, tmp := filepath.Join(dir, "work"), filepath.Join(dir, "tmp")
			for name, path := range map[string]string{"work": work, "tmp": tmp} {
				// An unavailable directory makes the same -w predicate false even
				// under root CI, without changing permissions on any global path.
				if name != blocked {
					if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			marker := filepath.Join(dir, "mc-called")
			executable(t, dir, "mc", "#!/bin/sh\ntouch \"$MC_MARKER\"\n")
			cmd := exec.Command("sh")
			cmd.Stdin = strings.NewReader(strings.NewReplacer("/work", work, "/tmp", tmp).Replace(script))
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "MC_MARKER="+marker,
				"MODE=external", "CREATE_BUCKETS=false", "ENDPOINT=http://minio:9000", "ACCESS_ID=test-user", "ACCESS_KEY=do-not-print-me",
				"LOG_BUCKET=qa-logs", "DATA_BUCKET=qa-data")
			result, err := cmd.CombinedOutput()
			if exitCode(err) != 1 || !strings.Contains(string(result), "initialization identity cannot write private working directories") ||
				!strings.Contains(string(result), "initialization.runAsUser/runAsGroup/fsGroup") || !strings.Contains(string(result), "volume ACL") {
				t.Fatalf("local volume failure was not diagnosed immediately: %v %s", err, result)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatal("mc was invoked before detecting inaccessible local working directories")
			}
			if strings.Contains(string(result), "do-not-print-me") || strings.Contains(string(result), "credentials were rejected") {
				t.Fatalf("local permissions failure leaked credentials or was misreported as an authentication error: %s", result)
			}
		})
	}
}
