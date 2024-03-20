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

package builder

import (
	"errors"

	tt "github.com/oceanbase/ob-operator/pkg/task/types"
)

type TypedTask[T any] func(T) tt.TaskError
type TaskMap[T any] map[tt.TaskName]TypedTask[T]

func NewTaskMap[T any]() TaskMap[T] {
	return make(map[tt.TaskName]TypedTask[T])
}

func (t TaskMap[T]) Register(name tt.TaskName, task TypedTask[T]) {
	t[name] = task
}

func (t TaskMap[T]) GetTask(name tt.TaskName, resource T) (tt.TaskFunc, error) {
	task, ok := t[name]
	if !ok {
		return nil, errors.New("Task not found: " + string(name))
	}
	return func() tt.TaskError {
		return task(resource)
	}, nil
}

func (t TaskMap[T]) Run(name tt.TaskName, resource T) tt.TaskError {
	task, err := t.GetTask(name, resource)
	if err != nil {
		return err
	}
	return task()
}
