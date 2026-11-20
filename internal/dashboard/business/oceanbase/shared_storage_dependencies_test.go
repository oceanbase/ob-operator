package oceanbase

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

func TestSharedStorageDependencies(t *testing.T) {
	for _, tc := range []struct {
		name, namespace, status                             string
		missingLS, missingSecret, emptyKey, deleting, valid bool
	}{
		{name: "running dependencies", namespace: "test", status: "running", valid: true},
		{name: "missing logservice", namespace: "test", missingLS: true},
		{name: "cross namespace", namespace: "other", status: "running"},
		{name: "not ready", namespace: "test", status: "new"},
		{name: "deleting", namespace: "test", status: "running", deleting: true},
		{name: "missing secret", namespace: "test", status: "running", missingSecret: true},
		{name: "empty credential", namespace: "test", status: "running", emptyKey: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &param.CreateOBClusterParam{Namespace: tc.namespace, DeploymentMode: "shared_storage", LogServiceRef: &common.ObjectReference{Name: "ls"}, SharedStorageInfo: &common.SharedStorageSpec{SecretRef: common.ObjectReference{Name: "storage"}}}
			ls := &unstructured.Unstructured{Object: map[string]any{"apiVersion": "oceanbase.oceanbase.com/v1alpha1", "kind": "OBLogServiceCluster", "metadata": map[string]any{"name": "ls", "namespace": "test"}, "status": map[string]any{"status": tc.status}}}
			if tc.deleting {
				now := metav1.Now()
				ls.SetDeletionTimestamp(&now)
			}
			objects := []runtime.Object{}
			if !tc.missingLS {
				objects = append(objects, ls)
			}
			dyn := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "storage", Namespace: "test"}, Data: map[string][]byte{"access_id": []byte("test-only"), "access_key": []byte("test-only")}}
			if tc.emptyKey {
				secret.Data["access_key"] = nil
			}
			objects = nil
			if !tc.missingSecret {
				objects = append(objects, secret)
			}
			core := kubefake.NewSimpleClientset(objects...)
			if err := checkSharedStorageDependencies(context.Background(), p, dyn, core.CoreV1()); (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
}
