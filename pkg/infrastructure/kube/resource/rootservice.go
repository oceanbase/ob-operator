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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
)

type RootServiceResource struct {
	*Resource
}

func NewRootServiceResource(resource *Resource) ResourceOperator {
	return &RootServiceResource{resource}
}

func (r *RootServiceResource) Create(ctx context.Context, obj interface{}) error {
	rootService := obj.(cloudv1.RootService)
	// kube.LogForAppActionStatus(rootService.Kind, rootService.Name, "create", rootService)
	err := r.Client.Create(ctx, &rootService)
	if err != nil {
		r.Recorder.Eventf(&rootService, corev1.EventTypeWarning, FailedToCreateRootService, "create RootService"+rootService.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(rootService.Kind, rootService.Name, "create", "succeed")
	r.Recorder.Event(&rootService, corev1.EventTypeNormal, CreatedRootService, "create RootService"+rootService.Name)
	return nil
}

func (r *RootServiceResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var rootServiceCurrent cloudv1.RootService
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &rootServiceCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return rootServiceCurrent, err
}

func (r *RootServiceResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *RootServiceResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *RootServiceResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	rootService := obj.(cloudv1.RootService)
	err := r.Client.Status().Update(ctx, &rootService)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *RootServiceResource) Delete(ctx context.Context, obj interface{}) error {
	rootService := obj.(cloudv1.RootService)
	// kube.LogForAppActionStatus(rootService.Kind, rootService.Name, "delete", rootService)
	err := r.Client.Delete(ctx, &rootService)
	if err != nil {
		r.Recorder.Eventf(&rootService, corev1.EventTypeWarning, FailedToDeleteRootService, "delete RootService CR"+rootService.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(rootService.Kind, rootService.Name, "delete", "succeed")
	r.Recorder.Event(&rootService, corev1.EventTypeNormal, DeletedRootService, "delete RootService CR"+rootService.Name)
	return nil
}

func (r *RootServiceResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
