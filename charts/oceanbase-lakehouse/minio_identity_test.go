package lakehouse_test

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestMinIONonRootIdentityConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name            string
		user, group, fs int64
		values          map[string]interface{}
	}{
		{"default", 1000, 1000, 1000, nil},
		{"platform-assigned", 10001, 10002, 10003, map[string]interface{}{"minio": map[string]interface{}{"runAsUser": 10001, "runAsGroup": 10002, "fsGroup": 10003}}},
		{"existing-pvc", 10001, 10001, 10001, map[string]interface{}{"minio": map[string]interface{}{"runAsUser": 10001, "runAsGroup": 10001, "fsGroup": 10001, "existingClaim": "user-owned-pvc"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			docs, output, err := render(t, tc.values)
			if err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			var minio appsv1.StatefulSet
			decodeScheduledObject(t, find(t, docs, "StatefulSet", "qa-minio"), &minio)
			security := minio.Spec.Template.Spec.SecurityContext
			if security == nil || security.RunAsUser == nil || *security.RunAsUser != tc.user ||
				security.RunAsGroup == nil || *security.RunAsGroup != tc.group ||
				security.FSGroup == nil || *security.FSGroup != tc.fs {
				t.Fatalf("incorrect MinIO platform identity: %+v", security)
			}
			if security.RunAsNonRoot == nil || !*security.RunAsNonRoot || security.SeccompProfile == nil || security.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
				t.Fatal("custom identity must preserve non-root and RuntimeDefault protections")
			}
			if len(minio.Spec.Template.Spec.InitContainers) != 0 {
				t.Fatal("identity configuration must not introduce a root/chown init container")
			}
			if tc.name == "existing-pvc" {
				for _, doc := range docs {
					if doc["kind"] == "PersistentVolumeClaim" {
						t.Fatal("custom identity must not create or adopt a supplied PVC")
					}
				}
				if minio.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "user-owned-pvc" {
					t.Fatal("custom identity changed the existing PVC reference")
				}
			}
			container := minio.Spec.Template.Spec.Containers[0].SecurityContext
			if container == nil || container.AllowPrivilegeEscalation == nil || *container.AllowPrivilegeEscalation ||
				container.ReadOnlyRootFilesystem == nil || !*container.ReadOnlyRootFilesystem ||
				(container.Privileged != nil && *container.Privileged) ||
				container.Capabilities == nil || len(container.Capabilities.Add) != 0 ||
				len(container.Capabilities.Drop) != 1 || container.Capabilities.Drop[0] != "ALL" {
				t.Fatal("custom identity must not elevate MinIO privileges")
			}
			var initialize batchv1.Job
			decodeScheduledObject(t, find(t, docs, "Job", "qa-storage-init-1"), &initialize)
			jobSecurity := initialize.Spec.Template.Spec.SecurityContext
			if jobSecurity == nil || jobSecurity.RunAsUser == nil || *jobSecurity.RunAsUser != 1000 ||
				jobSecurity.RunAsGroup == nil || *jobSecurity.RunAsGroup != 1000 || jobSecurity.FSGroup == nil || *jobSecurity.FSGroup != 1000 {
				t.Fatal("MinIO PVC identity must not change the independent emptyDir initialization Job")
			}
		})
	}
}

func TestMinIORejectsRootOrInvalidIdentity(t *testing.T) {
	for _, key := range []string{"runAsUser", "runAsGroup", "fsGroup"} {
		for _, value := range []interface{}{0, -1, 1.5, "10001", true, int64(4294967296)} {
			if _, _, err := render(t, map[string]interface{}{"minio": map[string]interface{}{key: value}}); err == nil {
				t.Fatalf("unsafe or non-integer %s=%v accepted", key, value)
			}
		}
	}
}
