package obzone

import (
	"testing"

	"github.com/go-logr/logr"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func TestBootstrapWaitFailureKeepsZoneInItsCurrentPhase(t *testing.T) {
	for _, tc := range []struct {
		status string
		task   tasktypes.TaskName
	}{{zonestatus.New, tWaitOBServerBootstrapReady}, {zonestatus.BootstrapReady, tWaitOBServerRunning}} {
		logger := logr.Discard()
		m := &OBZoneManager{OBZone: &v1alpha1.OBZone{}, Logger: &logger}
		m.OBZone.Status.Status = tc.status
		m.OBZone.Status.OperationContext = &tasktypes.OperationContext{Task: tc.task, TaskStatus: taskstatus.Failed, OnFailure: tasktypes.FailureRule{Strategy: strategy.StartOver, NextTryStatus: zonestatus.Running}}
		flow, err := m.GetTaskFlow()
		if err != nil || flow.OperationContext.OnFailure.Strategy != strategy.RetryFromCurrent {
			t.Fatalf("legacy boot wait %s: flow=%v err=%v", tc.task, flow, err)
		}
		m.HandleFailure()
		if m.OBZone.Status.Status != tc.status || m.OBZone.Status.OperationContext.TaskStatus != taskstatus.Pending {
			t.Fatalf("failed wait %s skipped to %s: %+v", tc.task, m.OBZone.Status.Status, m.OBZone.Status.OperationContext)
		}
	}
}
