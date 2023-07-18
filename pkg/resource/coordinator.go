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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/task"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
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

func (c *Coordinator) Coordinate() error {
	if c.Manager.IsNewResource() {
		c.Logger.Info("Need init status for resource")
		c.Manager.InitStatus()
	} else {
		f, err := c.Manager.GetTaskFlow()
		if err != nil {
			return errors.Wrap(err, "Get task flow")
		} else if f == nil {
			c.Logger.Info("No need to execute task flow")
		} else {
			c.Logger.Info("set operation context", "operation context", f.OperationContext)
			c.Manager.SetOperationContext(f.OperationContext)
			c.Logger.Info("Successfully got task flow")
			// execution errors reflects by task status
			c.executeTaskFlow(f)
		}
	}

	// handle instance deletion
	if c.Manager.IsDeleting() {
		err := c.Manager.CheckAndUpdateFinalizers()
		if err != nil {
			return errors.Wrapf(err, "Check and update finalizer failed")
		}
	}
	return c.Manager.UpdateStatus()
}

func (c *Coordinator) executeTaskFlow(f *task.TaskFlow) {
	switch f.OperationContext.TaskStatus {
	case taskstatus.Empty:
		if !f.HasNext() {
			c.Logger.Info("No task to execute")
			// clean task info sets resource status to normal, and context to nil
			c.Manager.ClearTaskInfo()
		} else {
			c.Logger.Info("Set first task to execute")
			f.NextTask()
		}
	case taskstatus.Pending:
		// run the current task while set task status to running
		taskFunc, err := c.Manager.GetTaskFunc(f.OperationContext.Task)
		if err != nil {
			c.Logger.Error(err, "No executable function found for task")
		} else {
			taskId := task.GetTaskManager().Submit(taskFunc)
			c.Logger.Info("Successfullly submit task", "taskid", taskId)
			f.OperationContext.TaskId = taskId
			f.OperationContext.TaskStatus = taskstatus.Running
		}
	case taskstatus.Running:
		// check task status and update cr status
		taskResult, err := task.GetTaskManager().GetTaskResult(f.OperationContext.TaskId)
		if err != nil {
			c.Logger.Error(err, "Get task result got error", "task id", f.OperationContext.TaskId)
			f.OperationContext.TaskStatus = taskstatus.Failed
		} else {
			if taskResult != nil {
				c.Logger.Info("task finished", "task id", f.OperationContext.TaskId, "task result", taskResult.Status)
				f.OperationContext.TaskStatus = taskResult.Status
			} else {
				c.Logger.Info("Didn't get task result, task is still running", "task id", f.OperationContext.TaskId)
			}
		}
	case taskstatus.Successful:
		// clean operation context and set status to target status
		if !f.HasNext() {
			c.Logger.Info("No more task to run, task flow successfully finished")
			c.Manager.FinishTask()
		} else {
			c.Logger.Info("Task finished successfully, set next task")
			f.NextTask()
		}
	case taskstatus.Failed:
		// TODO handle failed task
		c.Logger.Info("Task failed, back to initial status")
		c.Manager.ClearTaskInfo()
	}
	c.Logger.Info("Coordinate finished", "operation context", f.OperationContext)
}
