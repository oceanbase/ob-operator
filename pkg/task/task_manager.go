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
			ResultMap:       make(map[string]chan *TaskResult),
			Logger:          &logger,
			TaskResultCache: make(map[string]*TaskResult, 0),
		}
	})
	return taskManager
}

type TaskResult struct {
	Status string
	Error  error
}

type TaskManager struct {
	ResultMap       map[string]chan *TaskResult
	Logger          *logr.Logger
	TaskResultCache map[string]*TaskResult
}

func (m *TaskManager) Submit(f func() error) string {
	retCh := make(chan *TaskResult, 1)
	taskId := uuid.New().String()
	// TODO add lock to keep ResultMap safe
	m.ResultMap[taskId] = retCh
	m.TaskResultCache[taskId] = nil
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
	retCh, exists := m.ResultMap[taskId]
	if !exists {
		return nil, errors.Errorf("Task %s not exists", taskId)
	}
	if m.TaskResultCache[taskId] == nil {
		select {
		case result := <-retCh:
			m.TaskResultCache[taskId] = result
			return result, nil
		default:
			return nil, nil
		}
	} else {
		return m.TaskResultCache[taskId], nil
	}
}

func (m *TaskManager) CleanTaskResult(taskId string) error {
	retCh, exists := m.ResultMap[taskId]
	if !exists {
		return nil
	}
	close(retCh)
	delete(m.ResultMap, taskId)
	delete(m.TaskResultCache, taskId)
	return nil
}
