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

package task

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskManager *TaskManager
var taskManagerOnce sync.Once

func GetTaskManager() *TaskManager {
	taskManagerOnce.Do(func() {
		logger := log.FromContext(context.TODO())
		taskManager = &TaskManager{
			Logger: &logger,
		}
	})
	return taskManager
}

type TaskManager struct {
	ResultMap       sync.Map
	Logger          *logr.Logger
	TaskResultCache sync.Map
}

func (m *TaskManager) Submit(f tasktypes.TaskFunc) tasktypes.TaskID {
	retCh := make(chan *tasktypes.TaskResult, 1)
	// Notes: casting type here is important as equality of interface including type equality and value equality
	taskId := tasktypes.TaskID(uuid.New().String())
	m.ResultMap.Store(taskId, retCh)
	m.TaskResultCache.Delete(taskId)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				retCh <- &tasktypes.TaskResult{
					Status: taskstatus.Failed,
					Error:  errors.Errorf("Observed a panic: %v, stacktrace: %s", r, string(debug.Stack())),
				}
			}
		}()
		err := f()
		if err != nil {
			m.Logger.Error(err, "Run task got error", "taskId", taskId)
			retCh <- &tasktypes.TaskResult{
				Status: taskstatus.Failed,
				Error:  err,
			}
		} else {
			retCh <- &tasktypes.TaskResult{
				Status: taskstatus.Successful,
				Error:  nil,
			}
		}
	}()
	return taskId
}

func (m *TaskManager) GetTaskResult(taskId tasktypes.TaskID) (*tasktypes.TaskResult, error) {
	retChAny, exists := m.ResultMap.Load(taskId)
	if !exists {
		return nil, errors.Errorf("Task %s not exists", taskId)
	}
	retCh, ok := retChAny.(chan *tasktypes.TaskResult)
	if !ok {
		return nil, errors.Errorf("Task %s not exists", taskId)
	}
	result, exists := m.TaskResultCache.Load(taskId)
	if !exists {
		select {
		case result := <-retCh:
			m.TaskResultCache.Store(taskId, result)
			return result, nil
		default:
			return nil, nil
		}
	}
	return result.(*tasktypes.TaskResult), nil
}

func (m *TaskManager) CleanTaskResult(taskId tasktypes.TaskID) error {
	retChAny, exists := m.ResultMap.Load(taskId)
	if !exists {
		return nil
	}
	retCh, ok := retChAny.(chan *tasktypes.TaskResult)
	if !ok {
		return nil
	}
	close(retCh)
	m.ResultMap.Delete(taskId)
	m.TaskResultCache.Delete(taskId)
	return nil
}
