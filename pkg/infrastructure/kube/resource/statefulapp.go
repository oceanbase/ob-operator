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

type StatefulAppResource struct {
	*Resource
}

func NewStatefulAppResource(resource *Resource) ResourceOperator {
	return &StatefulAppResource{resource}
}

func (r *StatefulAppResource) Create(ctx context.Context, obj interface{}) error {
	statefulApp := obj.(cloudv1.StatefulApp)
	// kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "create", statefulApp)
	err := r.Client.Create(ctx, &statefulApp)
	if err != nil {
		r.Recorder.Eventf(&statefulApp, corev1.EventTypeWarning, FailedToCreateStatefulApp, "create StatefulApp"+statefulApp.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "create", "succeed")
	r.Recorder.Event(&statefulApp, corev1.EventTypeNormal, CreatedStatefulApp, "create StatefulApp"+statefulApp.Name)
	return nil
}

func (r *StatefulAppResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var statefulAppCurrent cloudv1.StatefulApp
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &statefulAppCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return statefulAppCurrent, err
}

func (r *StatefulAppResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *StatefulAppResource) Update(ctx context.Context, obj interface{}) error {
	statefulAppNew := obj.(cloudv1.StatefulApp)
	err := r.Client.Update(ctx, &statefulAppNew)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *StatefulAppResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	statefulAppNew := obj.(cloudv1.StatefulApp)
	err := r.Client.Status().Update(ctx, &statefulAppNew)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *StatefulAppResource) Delete(ctx context.Context, obj interface{}) error {
	statefulApp := obj.(cloudv1.StatefulApp)
	// kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "delete", statefulApp)
	err := r.Client.Delete(ctx, &statefulApp)
	if err != nil {
		r.Recorder.Eventf(&statefulApp, corev1.EventTypeWarning, FailedToDeleteStatefulApp, "delete StatefulApp"+statefulApp.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(statefulApp.Kind, statefulApp.Name, "delete", "succeed")
	r.Recorder.Event(&statefulApp, corev1.EventTypeNormal, DeletedStatefulApp, "delete StatefulApp"+statefulApp.Name)
	return nil
}

func (r *StatefulAppResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}

func (r *StatefulAppResource) PatchStatus(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
