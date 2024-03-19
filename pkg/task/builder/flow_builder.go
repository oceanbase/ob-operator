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

	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type FlowGenerator[T any] func(T) *tasktypes.TaskFlow
type FlowMap[T any] map[tasktypes.FlowName]FlowGenerator[T]

func NewFlowMap[T any]() FlowMap[T] {
	return make(map[tasktypes.FlowName]FlowGenerator[T])
}

func (f FlowMap[T]) RegisterFlow(name tasktypes.FlowName, flow FlowGenerator[T]) {
	f[name] = flow
}

func (f FlowMap[T]) GetFlow(name tasktypes.FlowName, resource T) (*tasktypes.TaskFlow, error) {
	gen, ok := f[name]
	if !ok {
		return nil, errors.New("TaskFlow not found: " + string(name))
	}
	return gen(resource), nil
}

type FlowBuilder interface {
	Build() *tasktypes.TaskFlow
	Step(task tasktypes.TaskName) FlowBuilder
	Steps(tasks ...tasktypes.TaskName) FlowBuilder
	To(status string) FlowBuilder
	FailedTo(status string) FlowBuilder
	RetryStrategy(strategy tasktypes.TaskFailureStrategy) FlowBuilder
	MaxRetry(max int) FlowBuilder
}

type flowBuilder struct {
	operationContext *tasktypes.OperationContext
}

func NewFlowBuilder(name tasktypes.FlowName) FlowBuilder {
	return &flowBuilder{
		operationContext: &tasktypes.OperationContext{
			Name:  name,
			Tasks: []tasktypes.TaskName{},
		},
	}
}

func (b *flowBuilder) Build() *tasktypes.TaskFlow {
	if b.operationContext.TargetStatus == "" {
		b.operationContext.TargetStatus = "running"
	}
	if b.operationContext.OnFailure.Strategy == "" {
		b.operationContext.OnFailure.Strategy = strategy.StartOver
	}
	if b.operationContext.OnFailure.MaxRetry == 0 {
		b.operationContext.OnFailure.MaxRetry = 32
	}
	if b.operationContext.OnFailure.NextTryStatus == "" {
		b.operationContext.OnFailure.NextTryStatus = "failed"
	}
	return &tasktypes.TaskFlow{
		OperationContext: b.operationContext,
	}
}

func (b *flowBuilder) Step(task tasktypes.TaskName) FlowBuilder {
	b.operationContext.Tasks = append(b.operationContext.Tasks, task)
	return b
}

func (b *flowBuilder) To(status string) FlowBuilder {
	b.operationContext.TargetStatus = status
	return b
}

func (b *flowBuilder) FailedTo(status string) FlowBuilder {
	b.operationContext.OnFailure.NextTryStatus = status
	return b
}

func (b *flowBuilder) RetryStrategy(strategy tasktypes.TaskFailureStrategy) FlowBuilder {
	b.operationContext.OnFailure.Strategy = strategy
	return b
}

func (b *flowBuilder) MaxRetry(max int) FlowBuilder {
	b.operationContext.OnFailure.MaxRetry = max
	return b
}

func (b *flowBuilder) Steps(tasks ...tasktypes.TaskName) FlowBuilder {
	b.operationContext.Tasks = append(b.operationContext.Tasks, tasks...)
	return b
}
