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
)

var taskFlowRegistry *Registry
var taskFlowRegistryOnce sync.Once

func GetRegistry() *Registry {
	taskFlowRegistryOnce.Do(func() {
		taskFlowRegistry = &Registry{
			Store: make(map[string]func() *TaskFlow),
		}
	})
	return taskFlowRegistry
}

type Registry struct {
	Store map[string]func() *TaskFlow
}

func (r *Registry) Register(name string, f func() *TaskFlow) error {
	_, exists := r.Store[name]
	if exists {
		panic(fmt.Sprintf("Task flow %s already registered", name))
	}
	r.Store[name] = f
	return nil
}

func (r *Registry) Get(name string) (*TaskFlow, error) {
	f, exists := r.Store[name]
	if !exists {
		return nil, errors.Errorf("Task flow %s not registered", name)
	}
	return f(), nil
}
