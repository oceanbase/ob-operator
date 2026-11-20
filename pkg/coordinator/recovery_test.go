/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package coordinator

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tt "github.com/oceanbase/ob-operator/pkg/task/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type recoveryManager struct {
	ResourceManager
	events int
	calls  atomic.Int32
}

func (m *recoveryManager) GetMeta() metav1.Object              { return &metav1.ObjectMeta{Name: "test"} }
func (m *recoveryManager) PrintErrEvent(error)                 { m.events++ }
func (m *recoveryManager) CanReplayTask(name tt.TaskName) bool { return name == "wait ready" }
func (m *recoveryManager) GetTaskFunc(tt.TaskName) (tt.TaskFunc, error) {
	return func() tt.TaskError { m.calls.Add(1); return nil }, nil
}

// Embedding only ResourceManager deliberately hides the optional replay policy.
type noReplayPolicy struct{ ResourceManager }

func TestLostTaskRecovery(t *testing.T) {
	for _, tc := range []struct {
		name     string
		taskName tt.TaskName
		setup    string
		policy   bool
		want     tt.TaskStatus
	}{
		{"missing wait", "wait ready", "missing", true, taskstatus.Pending},
		{"missing bootstrap", "bootstrap", "missing", true, taskstatus.Failed},
		{"no opt in", "wait ready", "missing", false, taskstatus.Failed},
		{"corrupt channel", "wait ready", "corrupt", true, taskstatus.Failed},
		{"real task failure", "wait ready", "failed", true, taskstatus.Failed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := &recoveryManager{}
			var manager ResourceManager = m
			if !tc.policy {
				manager = &noReplayPolicy{m}
			}
			logger := logr.Discard()
			c := NewCoordinator(manager, &logger)
			id := tt.TaskID(t.Name())
			tm := task.GetTaskManager()
			t.Cleanup(func() { tm.ResultMap.Delete(id); tm.TaskResultCache.Delete(id) })
			switch tc.setup {
			case "corrupt":
				tm.ResultMap.Store(id, "invalid")
			case "failed":
				tm.TaskResultCache.Store(id, &tt.TaskResult{Status: taskstatus.Failed, Error: errors.New("readiness timeout")})
			}
			ctx := &tt.OperationContext{Task: tc.taskName, TaskId: id, TaskStatus: taskstatus.Running, Idx: 1, Tasks: []tt.TaskName{tc.taskName, "bootstrap"}}
			ctx.OnFailure.RetryCount = 2
			f := tt.NewTaskFlow(ctx)
			c.executeTaskFlow(f)
			if ctx.TaskStatus != tc.want {
				t.Fatalf("status=%s want=%s", ctx.TaskStatus, tc.want)
			}
			if ctx.Idx != 1 || ctx.Task != tc.taskName || ctx.OnFailure.RetryCount != 2 {
				t.Fatalf("recovery changed flow position or retries: %+v", ctx)
			}
			if tc.want != taskstatus.Pending {
				if m.events != 1 || m.calls.Load() != 0 {
					t.Fatal("failed task was replayed or failure hidden")
				}
				return
			}
			if ctx.TaskId != "" || m.events != 0 {
				t.Fatal("recovery retained stale ID or emitted failure")
			}
			c.executeTaskFlow(f)
			newID := ctx.TaskId
			t.Cleanup(func() { tm.CleanTaskResult(newID) })
			if newID == "" || newID == id || ctx.TaskStatus != taskstatus.Running {
				t.Fatal("task not resubmitted")
			}
			deadline := time.Now().Add(5 * time.Second)
			for ctx.TaskStatus == taskstatus.Running && time.Now().Before(deadline) {
				c.executeTaskFlow(f)
				time.Sleep(time.Millisecond)
			}
			if ctx.TaskStatus != taskstatus.Successful || m.calls.Load() != 1 || ctx.Idx != 1 {
				t.Fatal("wait task did not finish exactly once at original position")
			}
		})
	}
}

type statusWriteManager struct {
	recoveryManager
	stored   tt.OperationContext
	pending  *tt.OperationContext
	writeErr error
}

func (m *statusWriteManager) GetStatus() string { return "running" }
func (m *statusWriteManager) GetTaskFlow() (*tt.TaskFlow, error) {
	ctx := m.stored
	return tt.NewTaskFlow(&ctx), nil
}
func (m *statusWriteManager) SetOperationContext(ctx *tt.OperationContext) { m.pending = ctx }
func (m *statusWriteManager) UpdateStatus() error {
	if m.writeErr == nil {
		m.stored = *m.pending
	}
	return m.writeErr
}

func TestKeepFinishedTaskUntilStatusIsPersisted(t *testing.T) {
	for _, terminal := range []tt.TaskStatus{taskstatus.Successful, taskstatus.Failed} {
		t.Run(string(terminal), func(t *testing.T) {
			id := tt.TaskID(t.Name())
			writeErr := errors.New("status update conflict")
			manager := &statusWriteManager{stored: tt.OperationContext{Task: "apply change", TaskId: id, TaskStatus: taskstatus.Running}, writeErr: writeErr}
			logger := logr.Discard()
			coordinator := NewCoordinator(manager, &logger)
			tm := task.GetTaskManager()
			result := &tt.TaskResult{Status: terminal}
			ch := make(chan *tt.TaskResult, 1)
			ch <- result
			tm.ResultMap.Store(id, ch)
			t.Cleanup(func() { tm.ResultMap.Delete(id); tm.TaskResultCache.Delete(id) })

			if _, err := coordinator.Coordinate(); !errors.Is(err, writeErr) {
				t.Fatalf("first reconcile: %v", err)
			}
			if got, err := tm.GetTaskResult(id); err != nil || got != result {
				t.Fatalf("finished result after failed status write: got %v, err %v", got, err)
			}
			manager.writeErr = nil
			if _, err := coordinator.Coordinate(); err != nil {
				t.Fatal(err)
			}
			if manager.stored.TaskStatus != terminal || manager.calls.Load() != 0 {
				t.Fatalf("persisted status=%s, resubmissions=%d", manager.stored.TaskStatus, manager.calls.Load())
			}
			if _, err := tm.GetTaskResult(id); !errors.Is(err, task.ErrTaskNotFound) {
				t.Fatalf("finished result should be released after persistence: %v", err)
			}
		})
	}
}

// The API may persist a terminal state even when the write response times out.
func TestReleaseResultWhenPersistedFlowAdvances(t *testing.T) {
	id := tt.TaskID(t.Name())
	manager := &statusWriteManager{stored: tt.OperationContext{Task: "first", TaskId: id, TaskStatus: taskstatus.Successful, Tasks: []tt.TaskName{"first", "second"}, Idx: 1}}
	tm := task.GetTaskManager()
	tm.ResultMap.Store(id, make(chan *tt.TaskResult, 1))
	t.Cleanup(func() { tm.ResultMap.Delete(id); tm.TaskResultCache.Delete(id) })
	logger := logr.Discard()
	c := NewCoordinator(manager, &logger)
	if _, err := c.Coordinate(); err != nil {
		t.Fatal(err)
	}
	if manager.stored.Task != "second" || manager.stored.TaskStatus != taskstatus.Pending {
		t.Fatalf("unexpected next task: %+v", manager.stored)
	}
	if _, err := tm.GetTaskResult(id); !errors.Is(err, task.ErrTaskNotFound) {
		t.Fatalf("old result was retained after flow advanced: %v", err)
	}
}

type bootstrapVerifier struct {
	recoveryManager
	checks atomic.Int32
}

func (m *bootstrapVerifier) ObserveLostTask(name tt.TaskName) tt.TaskFunc {
	if name != "bootstrap" {
		return nil
	}
	return func() tt.TaskError { m.checks.Add(1); return nil }
}

func TestLostWriteVerifiesWithoutRepeatingMutation(t *testing.T) {
	m := &bootstrapVerifier{}
	logger := logr.Discard()
	c := NewCoordinator(m, &logger)
	op := &tt.OperationContext{Task: "bootstrap", TaskId: tt.TaskID(t.Name()), TaskStatus: taskstatus.Running}
	f := tt.NewTaskFlow(op)
	c.executeTaskFlow(f)
	if op.TaskId == tt.TaskID(t.Name()) || op.TaskStatus != taskstatus.Running {
		t.Fatalf("verification was not submitted, status=%s", op.TaskStatus)
	}
	t.Cleanup(func() { _ = task.GetTaskManager().CleanTaskResult(op.TaskId) })
	deadline := time.Now().Add(3 * time.Second)
	for op.TaskStatus == taskstatus.Running && time.Now().Before(deadline) {
		c.executeTaskFlow(f)
		time.Sleep(time.Millisecond)
	}
	if op.TaskStatus != taskstatus.Successful || m.calls.Load() != 0 || m.checks.Load() != 1 {
		t.Fatalf("status=%s writes=%d checks=%d", op.TaskStatus, m.calls.Load(), m.checks.Load())
	}
}
