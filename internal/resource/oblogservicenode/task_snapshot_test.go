package oblogservicenode

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
)

type refreshDuringCreateClient struct {
	client.Client
	refresh func()
}

func TestFailedNodeResumesAppropriateCreateFlow(t *testing.T) {
	for _, state := range []string{"new", "failed", "running"} {
		t.Run(state, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = v1alpha1.AddToScheme(scheme)
			cluster := &v1alpha1.OBLogServiceCluster{ObjectMeta: metav1.ObjectMeta{Name: "ls", Namespace: "test"}}
			cluster.Status.Status = state
			node := &v1alpha1.OBLogServiceNode{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "test"}}
			node.Spec.ClusterName = "ls"
			node.Status.Status = nodestatus.Failed
			logger := logr.Discard()
			m := &OBLogServiceNodeManager{Ctx: context.Background(), Resource: node, Logger: &logger, Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(cluster).Build()}
			flow, err := m.GetTaskFlow()
			if err != nil || flow == nil {
				t.Fatalf("failed node has no retry flow: %v", err)
			}
			want := nodestatus.BootstrapReady
			if state == "running" {
				want = nodestatus.Running
			}
			if flow.OperationContext.TargetStatus != want {
				t.Fatalf("target=%s want=%s", flow.OperationContext.TargetStatus, want)
			}
		})
	}
}

func TestUpdateStatusObservesBootstrapPod(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	node := &v1alpha1.OBLogServiceNode{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "test"}}
	node.Status.Status = nodestatus.New
	node.Status.PodName = "node"
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "node", Namespace: "test", Annotations: map[string]string{oceanbaseconst.AnnotationCalicoValidate: "test"}}, Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.1", Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
	logger := logr.Discard()
	m := &OBLogServiceNodeManager{Ctx: context.Background(), Resource: node.DeepCopy(), Logger: &logger, Client: fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(node).WithObjects(node, pod).Build()}
	if err := m.UpdateStatus(); err != nil {
		t.Fatal(err)
	}
	stored := &v1alpha1.OBLogServiceNode{}
	if err := m.Client.Get(m.Ctx, client.ObjectKeyFromObject(node), stored); err != nil {
		t.Fatal(err)
	}
	if !stored.Status.Ready || stored.Status.CNI != oceanbaseconst.CNICalico || stored.Status.PodIP != "10.0.0.1" {
		t.Fatalf("observations were not persisted: %+v", stored.Status)
	}
}

func (c *refreshDuringCreateClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if obj.GetNamespace() == "" {
		return fmt.Errorf("an empty namespace may not be set during creation")
	}
	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	if strings.HasSuffix(obj.GetName(), "-store") {
		c.refresh()
	}
	return nil
}

// Coordinator submits the task and then updates the same CR's status. The
// client's response decoder may overwrite that CR while the task is running.
// Simulate that replacement exactly between the two PVC creations.
func TestPVCTaskSurvivesReconcileResourceRefresh(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	node := &v1alpha1.OBLogServiceNode{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1alpha1.GroupVersion.String(), Kind: "OBLogServiceNode"},
		ObjectMeta: metav1.ObjectMeta{Name: "ls-zone1-node", Namespace: "ss-test", UID: "test-uid"},
	}
	node.Spec.Storage = &apitypes.LogServiceStorageSpec{
		StoreStorage: &apitypes.StorageSpec{Size: resource.MustParse("30Gi"), StorageClass: "test"},
		LogStorage:   &apitypes.StorageSpec{Size: resource.MustParse("15Gi"), StorageClass: "test"},
	}
	k8s := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger := logr.Discard()
	m := &OBLogServiceNodeManager{Ctx: context.Background(), Resource: node, Logger: &logger}
	m.Client = &refreshDuringCreateClient{Client: k8s, refresh: func() { *node = v1alpha1.OBLogServiceNode{} }}
	task, err := m.GetTaskFunc(tCreatePVC)
	if err != nil {
		t.Fatal(err)
	}
	if err := task(); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"store", "log"} {
		pvc := &corev1.PersistentVolumeClaim{}
		if err := k8s.Get(context.Background(), client.ObjectKey{Namespace: "ss-test", Name: "ls-zone1-node-" + suffix}, pvc); err != nil {
			t.Fatal(err)
		}
		if pvc.OwnerReferences[0].UID != "test-uid" {
			t.Fatal("owner identity changed during task")
		}
	}
}
