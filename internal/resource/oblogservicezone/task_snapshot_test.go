package oblogservicezone

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicezone"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestFailedZoneResumesAppropriateCreateFlow(t *testing.T) {
	for _, state := range []string{"new", "failed", "running"} {
		t.Run(state, func(t *testing.T) {
			scheme := runtime.NewScheme()
			_ = v1alpha1.AddToScheme(scheme)
			cluster := &v1alpha1.OBLogServiceCluster{ObjectMeta: metav1.ObjectMeta{Name: "ls", Namespace: "test"}}
			cluster.Status.Status = state
			zone := &v1alpha1.OBLogServiceZone{ObjectMeta: metav1.ObjectMeta{Name: "zone", Namespace: "test", Labels: map[string]string{oceanbaseconst.LabelRefOBLogServiceCluster: "ls"}}}
			zone.Status.Status = zonestatus.Failed
			logger := logr.Discard()
			m := &OBLogServiceZoneManager{Ctx: context.Background(), Resource: zone, Logger: &logger, Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(cluster).Build()}
			flow, err := m.GetTaskFlow()
			if err != nil || flow == nil {
				t.Fatalf("missing retry flow: %v", err)
			}
			want := zonestatus.BootstrapReady
			if state == "running" {
				want = zonestatus.Running
			}
			if flow.OperationContext.TargetStatus != want {
				t.Fatalf("target=%s want=%s", flow.OperationContext.TargetStatus, want)
			}
		})
	}
}

func TestZoneTaskUsesPrivateSnapshot(t *testing.T) {
	resource := &v1alpha1.OBLogServiceZone{ObjectMeta: metav1.ObjectMeta{Name: "zone", Namespace: "test", Labels: map[string]string{"k": "original"}}}
	m := &OBLogServiceZoneManager{Resource: resource}
	const taskName tasktypes.TaskName = "test zone snapshot"
	taskMap.Register(taskName, func(copy *OBLogServiceZoneManager) tasktypes.TaskError {
		if copy.Resource.Name != "zone" || copy.Resource.Labels["k"] != "original" {
			t.Fatal("task aliases reconciler resource")
		}
		copy.Resource.Status.Status = "task-only"
		return nil
	})
	task, err := m.GetTaskFunc(taskName)
	if err != nil {
		t.Fatal(err)
	}
	resource.Name = "updated"
	resource.Labels["k"] = "changed"
	if err := task(); err != nil {
		t.Fatal(err)
	}
	if resource.Status.Status != "" {
		t.Fatal("task mutated reconciler status")
	}
}
