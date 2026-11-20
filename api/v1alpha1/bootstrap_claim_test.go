package v1alpha1

import (
	"reflect"
	"strings"
	"testing"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	obconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestClassicBootstrapClaimCannotBeCopiedOrCleared(t *testing.T) {
	old := &OBCluster{}
	old.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "uid/1"}
	if _, err := old.ValidateCreate(); err == nil || !strings.Contains(err.Error(), "bootstrap dispatch claim") {
		t.Fatalf("clone carried another object's bootstrap claim: %v", err)
	}
	for _, annotations := range []map[string]string{nil, {obconst.AnnotationsClassicBootstrapClaim: "other/1"}} {
		next := old.DeepCopy()
		next.Annotations = annotations
		if _, err := next.ValidateUpdate(old); err == nil || !strings.Contains(err.Error(), "bootstrap dispatch claim") {
			t.Fatalf("updated cluster lost its durable claim: %v", err)
		}
	}
}

func TestClassicBootstrapClaimInsertionRequiresTheLiveTask(t *testing.T) {
	old := &OBCluster{}
	old.UID, old.Generation = "uid", 3
	old.Status.Status = clusterstatus.New
	old.Status.OperationContext = &tasktypes.OperationContext{Name: "bootstrap obcluster", Task: "dispatch classic bootstrap", TaskStatus: taskstatus.Pending}
	for _, stage := range []tasktypes.TaskStatus{taskstatus.Pending, taskstatus.Running} {
		old.Status.OperationContext.TaskStatus = stage
		newObj := old.DeepCopy()
		if err := newObj.validateBootstrapClaimUpdate(old); err != nil {
			t.Fatal(err)
		}
		newObj.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "uid/3"}
		if err := newObj.validateBootstrapClaimUpdate(old); err != nil {
			t.Fatal(err)
		}
	}
	for _, edit := range []func(*OBCluster){
		func(r *OBCluster) { r.Status.Status = clusterstatus.Running },
		func(r *OBCluster) { r.Status.OperationContext.Task = "bootstrap" },
		func(r *OBCluster) { r.Spec.DeploymentMode = obconst.DeploymentModeSharedStorage },
	} {
		invalid := old.DeepCopy()
		edit(invalid)
		next := invalid.DeepCopy()
		next.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "uid/3"}
		if err := next.validateBootstrapClaimUpdate(invalid); err == nil {
			t.Fatal("late or unrelated claim accepted")
		}
	}
	next := old.DeepCopy()
	next.Annotations = map[string]string{obconst.AnnotationsClassicBootstrapClaim: "another/3"}
	if err := next.validateBootstrapClaimUpdate(old); err == nil {
		t.Fatal("foreign claim accepted")
	}
	next.Annotations[obconst.AnnotationsClassicBootstrapClaim] = "uid/3"
	next.Spec.ClusterName = "changed"
	if err := next.validateBootstrapClaimUpdate(old); err == nil {
		t.Fatal("claim modified desired spec")
	}
}

func TestMetadataUpdatePreservesPersistedParameterOrder(t *testing.T) {
	defaults := func() map[string]apitypes.Parameter {
		return map[string]apitypes.Parameter{"a": {Name: "a", Value: "default"}, "b": {Name: "b", Value: "B"}, "c": {Name: "c", Value: "C"}}
	}
	stored := []apitypes.Parameter{{Name: "z", Value: "Z"}, {Name: "c", Value: "custom"}, {Name: "a", Value: "old"}, {Name: "a", Value: "newest"}}
	want := []apitypes.Parameter{stored[0], stored[1], stored[3], {Name: "b", Value: "B"}}
	got := applyParameterDefaults(stored, defaults())
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(applyParameterDefaults(got, defaults()), want) {
		t.Fatalf("metadata admission reordered existing values or replaced overrides: %#v", got)
	}
}
