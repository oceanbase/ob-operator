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

package resource

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	obconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
)

const (
	// If no task flow, requeue after 30 sec.
	NormalRequeueDuration = 30 * time.Second
	// In task flow, requeue after 1 sec.
	ExecutionRequeueDuration = 1 * time.Second
)

type Coordinator struct {
	Manager ResourceManager
	Logger  *logr.Logger
}

func NewCoordinator(m ResourceManager, logger *logr.Logger) *Coordinator {
	return &Coordinator{
		Manager: m,
		Logger:  logger,
	}
}

// 1. If the returned error is non-nil, the Result is ignored and the request will be
// requeued using exponential backoff. The only exception is if the error is a
// TerminalError in which case no requeuing happens.
//
// 2. If the error is nil and the returned Result has a non-zero result.RequeueAfter, the request
// will be requeued after the specified duration.
//
// 3. If the error is nil and result.RequeueAfter is zero and result.Reque is true, the request
// will be requeued using exponential backoff.
func (c *Coordinator) Coordinate() (ctrl.Result, error) {
	result := ctrl.Result{
		RequeueAfter: ExecutionRequeueDuration,
	}
	var f *task.TaskFlow
	var err error
	if c.Manager.IsNewResource() {
		c.Manager.InitStatus()
	} else {
		f, err = c.Manager.GetTaskFlow()
		if err != nil {
			return result, errors.Wrap(err, "Get task flow")
		} else if f == nil {
			// No need to execute task flow
			result.RequeueAfter = NormalRequeueDuration
		} else {
			c.Logger.V(obconst.LogLevelDebug).Info("Set operation context", "operation context", f.OperationContext)
			c.Manager.SetOperationContext(f.OperationContext)
			// execution errors reflects by task status
			c.executeTaskFlow(f)
			// if task status is `failed`, requeue after 2 ^ min(retryCount, threshold) * 500ms.
			// maximum backoff time is about 2 hrs with 14 as threshold.
			if f.OperationContext.OnFailure.RetryCount > 0 && f.OperationContext.TaskStatus == taskstatus.Failed {
				result.RequeueAfter = ExecutionRequeueDuration * (1 << min(f.OperationContext.OnFailure.RetryCount, obconst.TaskRetryBackoffThreshold))
			}
		}
	}
	// handle instance deletion
	if c.Manager.IsDeleting() {
		err := c.Manager.CheckAndUpdateFinalizers()
		if err != nil {
			return result, errors.Wrapf(err, "Check and update finalizer failed")
		}
		result.RequeueAfter = ExecutionRequeueDuration
	}
	err = c.cleanTaskResultMap(f)
	if err != nil {
		return result, errors.Wrap(err, "Clean task result map")
	}
	err = c.Manager.UpdateStatus()
	if err != nil {
		c.Logger.Error(err, "Failed to update status")
	}
	return result, err
}

func (c *Coordinator) executeTaskFlow(f *task.TaskFlow) {
	switch f.OperationContext.TaskStatus {
	case taskstatus.Empty:
		if !f.HasNext() {
			// clean task info sets resource status to normal, and context to nil
			c.Manager.ClearTaskInfo()
		} else {
			f.NextTask()
		}
	case taskstatus.Pending:
		// run the current task while set task status to running
		taskFunc, err := c.Manager.GetTaskFunc(f.OperationContext.Task)
		if err != nil {
			c.Logger.Error(err, "No executable function found for task")
		} else {
			c.Logger.V(obconst.LogLevelDebug).Info("Successfully get task flow")
			taskId := task.GetTaskManager().Submit(taskFunc)
			c.Logger.V(obconst.LogLevelDebug).Info("Successfully submit task", "taskId", taskId)
			f.OperationContext.TaskId = taskId
			f.OperationContext.TaskStatus = taskstatus.Running
		}
	case taskstatus.Running:
		// check task status and update cr status
		taskResult, err := task.GetTaskManager().GetTaskResult(f.OperationContext.TaskId)

		if err != nil {
			c.Logger.Error(err, "Get task result got error", "task id", f.OperationContext.TaskId)
			c.Manager.PrintErrEvent(err)
			f.OperationContext.TaskStatus = taskstatus.Failed
		} else if taskResult != nil {
			c.Logger.V(obconst.LogLevelDebug).Info("Task finished", "task id", f.OperationContext.TaskId, "task result", taskResult)
			f.OperationContext.TaskStatus = taskResult.Status
			if taskResult.Error != nil {
				c.Manager.PrintErrEvent(taskResult.Error)
			}
			// Didn't get task result, task is still running
		}
	case taskstatus.Successful:
		// clean operation context and set status to target status
		if !f.HasNext() {
			c.Manager.FinishTask()
		} else {
			f.NextTask()
		}
	case taskstatus.Failed:
		switch f.OperationContext.OnFailure.Strategy {
		case strategy.RetryFromCurrent, strategy.StartOver:
			// if strategy is retry or start over, limit the maximum retry times
			maxRetry := obconst.TaskMaxRetryTimes
			if !isZero(f.OperationContext.OnFailure.MaxRetry) {
				maxRetry = f.OperationContext.OnFailure.MaxRetry
			}
			if f.OperationContext.OnFailure.RetryCount > maxRetry {
				c.Logger.Info("Retry count exceeds limit, archive the resource")
				c.Manager.ArchiveResource()
			} else {
				c.Manager.HandleFailure()
				f.OperationContext.OnFailure.RetryCount++
			}
		default:
			c.Manager.HandleFailure()
		}
	}
	// Coordinate finished
}

func (c *Coordinator) cleanTaskResultMap(f *task.TaskFlow) error {
	if f == nil || f.OperationContext == nil {
		return nil
	}
	if f.OperationContext.TaskStatus == taskstatus.Successful || f.OperationContext.TaskStatus == taskstatus.Failed {
		err := task.GetTaskManager().CleanTaskResult(f.OperationContext.TaskId)
		if err != nil {
			return err
		}
	}
	return nil
}
