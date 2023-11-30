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

package types

import (
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
)

type FlowName string

type TaskFlow struct {
	OperationContext *OperationContext
}

func NewTaskFlow(c *OperationContext) *TaskFlow {
	return &TaskFlow{
		OperationContext: c,
	}
}

func (f *TaskFlow) NextTask() TaskName {
	if f.OperationContext.Idx >= len(f.OperationContext.Tasks) {
		f.OperationContext.Task = ""
	} else {
		f.OperationContext.TaskStatus = taskstatus.Pending
		f.OperationContext.Task = f.OperationContext.Tasks[f.OperationContext.Idx]
		f.OperationContext.Idx++
		f.OperationContext.TaskId = ""
	}
	return f.OperationContext.Task
}

func (f *TaskFlow) HasNext() bool {
	return f.OperationContext.Idx < len(f.OperationContext.Tasks)
}
