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
	"fmt"
	"os"
	"runtime"
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

func taskManagerInit() {
	logger := log.FromContext(context.TODO())
	taskManager = &TaskManager{
		Logger: &logger,
		tokens: make(chan struct{}, taskPoolSize),
	}
}

func runTask(f tasktypes.TaskFunc, ch chan<- *tasktypes.TaskResult, tokens chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			ch <- &tasktypes.TaskResult{
				Status: taskstatus.Failed,
				Error:  errors.Errorf("Observed a panic: %v, stacktrace: %s", r, string(debug.Stack())),
			}
		}
		<-tokens
		close(ch)
	}()

	tokens <- struct{}{}
	err := f()
	if err != nil {
		ch <- &tasktypes.TaskResult{
			Status: taskstatus.Failed,
			Error:  err,
		}
	} else {
		ch <- &tasktypes.TaskResult{
			Status: taskstatus.Successful,
			Error:  nil,
		}
	}
}

func GetTaskManager() *TaskManager {
	taskManagerOnce.Do(taskManagerInit)
	return taskManager
}

type TaskManager struct {
	ResultMap       sync.Map
	Logger          *logr.Logger
	TaskResultCache sync.Map

	mu          sync.Mutex
	workerCount uint32
	tokens      chan struct{}
}

func (m *TaskManager) Submit(f tasktypes.TaskFunc) tasktypes.TaskID {
	retCh := make(chan *tasktypes.TaskResult, 1)
	// Notes: casting type here is important as equality of interface including type equality and value equality
	taskId := tasktypes.TaskID(uuid.New().String())
	m.ResultMap.Store(taskId, retCh)
	m.TaskResultCache.Delete(taskId)

	go runTask(f, retCh, m.tokens)

	if debugTask {
		mu.Lock()
		m.workerCount++
		mu.Unlock()

		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		m.Logger.Info("[Submit] Memory usage",
			"Alloc", fmt.Sprintf("%v MiB", ms.Alloc>>20),
			"TotalAlloc", fmt.Sprintf("%v MiB", ms.TotalAlloc>>20),
			"Sys", fmt.Sprintf("%v MiB", ms.Sys>>20),
			"NumGC", ms.NumGC,
			"Running task workers", len(m.tokens),
			"Total task workers", m.workerCount,
			"Pool size", cap(m.tokens),
		)
	}
	return taskId
}

func (m *TaskManager) GetTaskResult(taskId tasktypes.TaskID) (*tasktypes.TaskResult, error) {
	if debugTask {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		m.Logger.Info("[GetResult] Memory usage",
			"Alloc", fmt.Sprintf("%v MiB", ms.Alloc>>20),
			"TotalAlloc", fmt.Sprintf("%v MiB", ms.TotalAlloc>>20),
			"Sys", fmt.Sprintf("%v MiB", ms.Sys>>20),
			"NumGC", ms.NumGC,
			"Running task workers", len(m.tokens),
			"Total task workers", m.workerCount,
			"Pool size", cap(m.tokens),
		)
	}
	result, exists := m.TaskResultCache.Load(taskId)
	if !exists {
		retChAny, exists := m.ResultMap.Load(taskId)
		if !exists {
			return nil, errors.Errorf("Task %s not exists", taskId)
		}
		retCh, ok := retChAny.(chan *tasktypes.TaskResult)
		if !ok {
			return nil, errors.Errorf("Task %s not exists", taskId)
		}
		select {
		case result, ok := <-retCh:
			if !ok {
				return nil, errors.Errorf("Result channel of task %s was closed", taskId)
			}
			m.TaskResultCache.Store(taskId, result)
			return result, nil
		default:
			return nil, nil
		}
	}
	return result.(*tasktypes.TaskResult), nil
}

func (m *TaskManager) CleanTaskResult(taskId tasktypes.TaskID) error {
	_, exists := m.ResultMap.Load(taskId)
	if !exists {
		return nil
	}
	m.ResultMap.Delete(taskId)
	m.TaskResultCache.Delete(taskId)

	if os.Getenv("DEBUG_TASK") == "true" {
		mu.Lock()
		m.workerCount--
		mu.Unlock()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)

		m.Logger.Info("[Clean] Memory usage",
			"Alloc", fmt.Sprintf("%v MiB", ms.Alloc>>20),
			"TotalAlloc", fmt.Sprintf("%v MiB", ms.TotalAlloc>>20),
			"Sys", fmt.Sprintf("%v MiB", ms.Sys>>20),
			"NumGC", ms.NumGC,
			"Running task workers", len(m.tokens),
			"Total task workers", m.workerCount,
			"Pool size", cap(m.tokens),
		)
	}

	return nil
}
