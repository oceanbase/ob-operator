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

	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
)

const (
	NormalRequeueDuration    = 60 * time.Second
	ExecutionRequeueDuration = 5 * time.Second
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
			c.Logger.Info("Set operation context", "operation context", f.OperationContext)
			c.Manager.SetOperationContext(f.OperationContext)
			// execution errors reflects by task status
			c.executeTaskFlow(f)
		}
	}
	// handle instance deletion
	if c.Manager.IsDeleting() {
		err := c.Manager.CheckAndUpdateFinalizers()
		if err != nil {
			return result, errors.Wrapf(err, "Check and update finalizer failed")
		}
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
			c.Logger.Info("Successfully get task flow")
			taskId := task.GetTaskManager().Submit(taskFunc)
			c.Logger.Info("Successfully submit task", "taskId", taskId)
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
		} else {
			if taskResult != nil {
				c.Logger.Info("Task finished", "task id", f.OperationContext.TaskId, "task result", taskResult)
				f.OperationContext.TaskStatus = taskResult.Status
				if taskResult.Error != nil {
					c.Manager.PrintErrEvent(taskResult.Error)
				}
			} else {
				// Didn't get task result, task is still running"
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
		c.Logger.Info("Task failed, back to initial status")
		c.Manager.HandleFailure()
	}
	c.cleanTaskResultMap(f)
	// Coordinate finished
}

// TODO clean task result map and cache map to free memory
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
