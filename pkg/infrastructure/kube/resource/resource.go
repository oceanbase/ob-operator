/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package resource

import (
	"context"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource struct {
	Client   client.Client
	Recorder record.EventRecorder
}

func NewResource(client client.Client, record record.EventRecorder) *Resource {
	return &Resource{
		Client:   client,
		Recorder: record,
	}
}

type ResourceOperator interface {
	Create(ctx context.Context, obj interface{}) error
	Get(ctx context.Context, namespace, name string) (interface{}, error)
	List(ctx context.Context, namespace string, listOption client.ListOption) interface{}
	Update(ctx context.Context, obj interface{}) error
	UpdateStatus(ctx context.Context, obj interface{}) error
	Delete(ctx context.Context, obj interface{}) error
	Patch(ctx context.Context, obj interface{}, patch client.Patch) error
	PatchStatus(ctx context.Context, obj interface{}, patch client.Patch) error
}
