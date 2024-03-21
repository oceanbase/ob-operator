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
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type FlowBuilder interface {
	BuildFlow() *tasktypes.TaskFlow
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

func (b *flowBuilder) BuildFlow() *tasktypes.TaskFlow {
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

// FlowGenerator generate a flow for a given type
type FlowGenerator[T any] func(T) *tasktypes.TaskFlow

// FlowGeneratorBuilder build a flow generator with a given type
type FlowGeneratorBuilder[T any] interface {
	FlowBuilder

	BuildGenerator() FlowGenerator[T]
	NamedTaskStep(named NamedTask[T]) FlowGeneratorBuilder[T]
	NamedTaskSteps(named ...NamedTask[T]) FlowGeneratorBuilder[T]
	GenFunc(gen func(T) *tasktypes.TaskFlow) FlowGeneratorBuilder[T]
}

type flowGeneratorBuilder[T any] struct {
	*flowBuilder
	gen func(T) *tasktypes.TaskFlow
}

// BuildGenerator build a generator function for the flow
func (b *flowGeneratorBuilder[T]) BuildGenerator() FlowGenerator[T] {
	if b.gen != nil {
		return b.gen
	}
	return func(T) *tasktypes.TaskFlow {
		return b.BuildFlow()
	}
}

// NamedTaskStep add a named task to the flow
func (b *flowGeneratorBuilder[T]) NamedTaskStep(named NamedTask[T]) FlowGeneratorBuilder[T] {
	b.Step(named.Name())
	return b
}

// NamedTaskSteps add named tasks to the flow
func (b *flowGeneratorBuilder[T]) NamedTaskSteps(named ...NamedTask[T]) FlowGeneratorBuilder[T] {
	for _, n := range named {
		b.Step(n.Name())
	}
	return b
}

// GenFunc set the generator function for the flow
func (b *flowGeneratorBuilder[T]) GenFunc(gen func(T) *tasktypes.TaskFlow) FlowGeneratorBuilder[T] {
	b.gen = gen
	return b
}

// NewFlowGenerator create a new flow generator builder
func NewFlowGenerator[T any](name tasktypes.FlowName) FlowGeneratorBuilder[T] {
	return &flowGeneratorBuilder[T]{
		flowBuilder: &flowBuilder{
			operationContext: &tasktypes.OperationContext{
				Name:  name,
				Tasks: []tasktypes.TaskName{},
			},
		},
	}
}
