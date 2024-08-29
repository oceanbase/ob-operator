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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
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
		RequeueAfter: cfg.ExecutionRequeueDuration,
	}
	var f *tasktypes.TaskFlow
	var err error
	beforeStatus := c.Manager.GetStatus()
	meta := c.Manager.GetMeta()
	if c.Manager.GetStatus() == "" {
		c.Manager.InitStatus()
	} else if meta.GetAnnotations()[cfg.PauseAnnotation] == "true" {
		c.Logger.V(2).Info("Pause annotation found, skip execution")
		result.RequeueAfter = cfg.PausedRequeueDuration
		return result, nil
	} else {
		f, err = c.Manager.GetTaskFlow()
		if err != nil {
			return result, errors.Wrap(err, "Get task flow")
		} else if f == nil {
			// No need to execute task flow
			result.RequeueAfter = cfg.NormalRequeueDuration
		} else {
			c.Logger.V(1).Info("Set operation context", "operation context", f.OperationContext)
			c.Manager.SetOperationContext(f.OperationContext)
			// execution errors reflects by task status
			c.executeTaskFlow(f)
			// if task status is `failed`, requeue after 2 ^ min(retryCount, threshold) * 500ms.
			// maximum backoff time is about 2 hrs with 14 as threshold.
			if f.OperationContext.OnFailure.RetryCount > 0 && f.OperationContext.TaskStatus == taskstatus.Failed {
				result.RequeueAfter = cfg.ExecutionRequeueDuration * (1 << Min(f.OperationContext.OnFailure.RetryCount, cfg.TaskRetryBackoffThreshold))
			}
		}
	}
	// handle instance deletion
	if meta.GetDeletionTimestamp() != nil {
		if meta.GetAnnotations()[cfg.IgnoreDeletionAnnotation] == "true" {
			c.Logger.V(2).Info("Ignore deletion annotation found, skip deletion")
		} else {
			err := c.Manager.CheckAndUpdateFinalizers()
			if err != nil {
				return result, errors.Wrapf(err, "Check and update finalizer failed")
			}
			result.RequeueAfter = cfg.ExecutionRequeueDuration
		}
	}
	err = c.cleanTaskResultMap(f)
	if err != nil {
		return result, errors.Wrap(err, "Clean task result map")
	}
	err = c.Manager.UpdateStatus()
	if err != nil {
		c.Logger.Error(err, "Failed to update status")
	}
	// When status changes(e.g. from running to other status), set a shorter `requeue after` to speed up processing.
	if c.Manager.GetStatus() != beforeStatus {
		result.RequeueAfter = cfg.ExecutionRequeueDuration
	}
	c.Logger.V(2).Info("Requeue after", "duration", result.RequeueAfter)
	return result, err
}

func (c *Coordinator) executeTaskFlow(f *tasktypes.TaskFlow) {
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
			c.Manager.PrintErrEvent(err)
		} else {
			c.Logger.V(1).Info("Successfully get task func " + f.OperationContext.Task.Display())
			taskId := task.GetTaskManager().Submit(taskFunc)
			c.Logger.V(1).Info("Successfully submit task", "taskId", taskId)
			f.OperationContext.TaskId = taskId
			f.OperationContext.TaskStatus = taskstatus.Running
		}
	case taskstatus.Running:
		// check task status and update cr status
		taskResult, err := task.GetTaskManager().GetTaskResult(f.OperationContext.TaskId)
		if err != nil || taskResult != nil {
			resMeta := c.Manager.GetMeta()
			ns := resMeta.GetNamespace()
			resVersion := resMeta.GetResourceVersion()
			resName := resMeta.GetName()
			loggingPairs := []interface{}{
				"flowName", f.OperationContext.Name,
				"taskId", f.OperationContext.TaskId,
				"taskName", f.OperationContext.Task,
				"namespace", ns,
				"resourceVersion", resVersion,
				"resourceName", resName,
			}
			if err != nil {
				c.Logger.Error(err, "Get task result got error", loggingPairs...)
				c.Manager.PrintErrEvent(err)
				f.OperationContext.TaskStatus = taskstatus.Failed
			} else if taskResult != nil {
				f.OperationContext.TaskStatus = taskResult.Status
				if taskResult.Error != nil {
					c.Logger.Error(taskResult.Error, "Task failed", loggingPairs...)
					c.Manager.PrintErrEvent(taskResult.Error)
				} else {
					c.Logger.V(1).Info("Task finished successfully", loggingPairs...)
				}
			}
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
			maxRetry := cfg.TaskMaxRetryTimes
			if f.OperationContext.OnFailure.MaxRetry != 0 {
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

func (c *Coordinator) cleanTaskResultMap(f *tasktypes.TaskFlow) error {
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
