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
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/task/status"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var taskManager *TaskManager
var taskManagerOnce sync.Once

func GetTaskManager() *TaskManager {
	taskManagerOnce.Do(func() {
		logger := log.FromContext(context.TODO())
		taskManager = &TaskManager{
			ResultMap: make(map[string]chan *TaskResult),
			Logger:    &logger,
		}
	})
	return taskManager
}

type TaskResult struct {
	Status string
	Error  error
}

type TaskManager struct {
	ResultMap map[string]chan *TaskResult
	Logger    *logr.Logger
}

func (m *TaskManager) Submit(f func() error) string {
	retCh := make(chan *TaskResult, 1)
	TaskId := uuid.New().String()
	// TODO add lock to keep ResultMap safe
	m.ResultMap[TaskId] = retCh
	go func() {
		err := f()
		if err != nil {
			m.Logger.Error(err, "Run task got error", "taskId", TaskId)
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
	return TaskId
}

// TODO currently result is only available for once, need store until clean
func (m *TaskManager) GetTaskResult(taskId string) (*TaskResult, error) {
	retCh, exists := m.ResultMap[taskId]
	if !exists {
		// m.Logger.Info("Query a task id that's not exists", "task id", taskId)
		return nil, errors.Errorf("Task %s not exists", taskId)
	}
	select {
	case result := <-retCh:
		return result, nil
	default:
		return nil, nil
	}
}

func (m *TaskManager) CleanTaskResult(taskId string) error {
	retCh, exists := m.ResultMap[taskId]
	if !exists {
		return errors.Errorf("Task %s not exists", taskId)
		// m.Logger.Error(err, "Task not exists", "task id", taskId)
	}
	close(retCh)
	delete(m.ResultMap, taskId)
	return nil
}
