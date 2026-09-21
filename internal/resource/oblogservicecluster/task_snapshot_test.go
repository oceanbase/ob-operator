package oblogservicecluster

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestClusterTaskUsesPrivateSnapshot(t *testing.T) {
	resource := &v1alpha1.OBLogServiceCluster{ObjectMeta: metav1.ObjectMeta{Name: "ls", Namespace: "test", Labels: map[string]string{"k": "original"}}}
	m := &OBLogServiceClusterManager{Resource: resource}
	const taskName tasktypes.TaskName = "test cluster snapshot"
	taskMap.Register(taskName, func(copy *OBLogServiceClusterManager) tasktypes.TaskError {
		if copy.Resource.Name != "ls" || copy.Resource.Labels["k"] != "original" {
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
