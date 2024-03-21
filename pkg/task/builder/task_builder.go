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

type TypedTaskFunc[T any] func(T) tt.TaskError

type NamedTask[T any] struct {
	taskName tt.TaskName
	fn       TypedTaskFunc[T]
}

func (t NamedTask[T]) Name() tt.TaskName {
	return t.taskName
}

func (t NamedTask[T]) Run(resource T) tt.TaskError {
	return t.fn(resource)
}

func (t NamedTask[T]) Func() TypedTaskFunc[T] {
	return t.fn
}

type TaskHub[T any] interface {
	Register(tt.TaskName, TypedTaskFunc[T])
	GetTask(tt.TaskName, T) (tt.TaskFunc, error)
	GetTypedTask(tt.TaskName) (TypedTaskFunc[T], error)
	Build(tt.TaskName, TypedTaskFunc[T]) NamedTask[T]
}

type taskMap[T any] struct {
	m map[tt.TaskName]TypedTaskFunc[T]
}

func NewTaskHub[T any]() TaskHub[T] {
	return &taskMap[T]{m: make(map[tt.TaskName]TypedTaskFunc[T])}
}

func (t *taskMap[T]) Build(name tt.TaskName, taskFunc TypedTaskFunc[T]) NamedTask[T] {
	t.Register(name, taskFunc)
	return NamedTask[T]{taskName: name, fn: taskFunc}
}

func (t taskMap[T]) Register(name tt.TaskName, taskFunc TypedTaskFunc[T]) {
	t.m[name] = taskFunc
}

func (t taskMap[T]) GetTask(name tt.TaskName, resource T) (tt.TaskFunc, error) {
	task, err := t.GetTypedTask(name)
	if err != nil {
		return nil, err
	}
	return func() tt.TaskError {
		return task(resource)
	}, nil
}

func (t taskMap[T]) GetTypedTask(name tt.TaskName) (TypedTaskFunc[T], error) {
	task, ok := t.m[name]
	if !ok {
		return nil, errors.New("Task not found: " + string(name))
	}
	return task, nil
}
