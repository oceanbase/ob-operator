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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
)

type ServiceResource struct {
	*Resource
}

func NewServiceResource(resource *Resource) ResourceOperator {
	return &ServiceResource{resource}
}

func (r *ServiceResource) Create(ctx context.Context, obj interface{}) error {
	service := obj.(corev1.Service)
	// kube.LogForAppActionStatus(service.Kind, service.Name, "create", service)
	err := r.Client.Create(ctx, &service)
	if err != nil {
		r.Recorder.Eventf(&service, corev1.EventTypeWarning, FailedToCreateService, "create Service"+service.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(service.Kind, service.Name, "create", "succeed")
	r.Recorder.Event(&service, corev1.EventTypeNormal, CreatedService, "create Service"+service.Name)
	return nil
}

func (r *ServiceResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	service := &corev1.Service{}
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), service)
	if err != nil {
		klog.Errorln(err)
	}
	return *service, err
}

func (r *ServiceResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	return nil
}

func (r *ServiceResource) Update(ctx context.Context, obj interface{}) error {
	service := obj.(corev1.Service)
	err := r.Client.Update(ctx, &service)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *ServiceResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	var res error
	return res
}

func (r *ServiceResource) Delete(ctx context.Context, obj interface{}) error {
	service := obj.(corev1.Service)
	// kube.LogForAppActionStatus(service.Kind, service.Name, "delete", service)
	err := r.Client.Delete(ctx, &service)
	if err != nil {
		r.Recorder.Eventf(&service, corev1.EventTypeWarning, FailedToDeleteService, "delete Service"+service.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(service.Kind, service.Name, "delete", "succeed")
	r.Recorder.Event(&service, corev1.EventTypeNormal, DeletedService, "delete Service"+service.Name)
	return nil
}

func (r *ServiceResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
