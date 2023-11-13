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

	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
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

type TaskResult struct {
	Status string
	Error  error
}

type TaskManager struct {
	ResultMap       sync.Map
	Logger          *logr.Logger
	TaskResultCache sync.Map
}

func (m *TaskManager) Submit(f func() error) string {
	retCh := make(chan *TaskResult, 1)
	taskId := uuid.New().String()
	// TODO add lock to keep ResultMap safe
	m.ResultMap.Store(taskId, retCh)
	m.TaskResultCache.Delete(taskId)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				retCh <- &TaskResult{
					Status: taskstatus.Failed,
					Error:  errors.Errorf("Observed a panic: %v, stacktrace: %s", r, string(debug.Stack())),
				}
			}
		}()
		err := f()
		if err != nil {
			m.Logger.Error(err, "Run task got error", "taskId", taskId)
			retCh <- &TaskResult{
				Status: taskstatus.Failed,
				Error:  err,
			}
		}
		retCh <- &TaskResult{
			Status: taskstatus.Successful,
			Error:  nil,
		}
	}()
	return taskId
}

func (m *TaskManager) GetTaskResult(taskId string) (*TaskResult, error) {
	retChAny, exists := m.ResultMap.Load(taskId)
	if !exists {
		return nil, errors.Errorf("Task %s not exists", taskId)
	}
	retCh, ok := retChAny.(chan *TaskResult)
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
	return result.(*TaskResult), nil
}

func (m *TaskManager) CleanTaskResult(taskId string) error {
	retChAny, exists := m.ResultMap.Load(taskId)
	if !exists {
		return nil
	}
	retCh, ok := retChAny.(chan *TaskResult)
	if !ok {
		return nil
	}
	close(retCh)
	m.ResultMap.Delete(taskId)
	m.TaskResultCache.Delete(taskId)
	return nil
}
