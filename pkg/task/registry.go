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
	"fmt"
	"sync"

	"github.com/pkg/errors"

	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

var taskFlowRegistry *Registry
var taskFlowRegistryOnce sync.Once
var mu sync.Mutex

func GetRegistry() *Registry {
	taskFlowRegistryOnce.Do(func() {
		taskFlowRegistry = &Registry{
			Store: make(map[tasktypes.FlowName]func() *tasktypes.TaskFlow),
		}
	})
	return taskFlowRegistry
}

type Registry struct {
	Store map[tasktypes.FlowName]func() *tasktypes.TaskFlow
}

func (r *Registry) Register(name tasktypes.FlowName, f func() *tasktypes.TaskFlow) {
	_, exists := r.Store[name]
	if exists {
		panic(fmt.Sprintf("Task flow %s already registered", name))
	}
	mu.Lock()
	defer mu.Unlock()
	r.Store[name] = f
}

func (r *Registry) Get(name tasktypes.FlowName) (*tasktypes.TaskFlow, error) {
	f, exists := r.Store[name]
	if !exists {
		return nil, errors.Errorf("Task flow %s not registered", name)
	}
	return f(), nil
}
