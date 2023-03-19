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
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RestoreResource struct {
	*Resource
}

func NewRestoreResource(resource *Resource) ResourceOperator {
	return &RestoreResource{resource}
}

func (r *RestoreResource) Create(ctx context.Context, obj interface{}) error {
	restore := obj.(cloudv1.Restore)
	err := r.Client.Create(ctx, &restore)
	if err != nil {
		r.Recorder.Eventf(&restore, corev1.EventTypeWarning, FailedToCreateRestore, "create Restore"+restore.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(restore.Kind, restore.Name, "create", "succeed")
	r.Recorder.Event(&restore, corev1.EventTypeNormal, CreatedRestore, "create Restore"+restore.Name)
	return nil
}

func (r *RestoreResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var restoreCurrent cloudv1.Restore
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &restoreCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return restoreCurrent, err
}

func (r *RestoreResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	restoreList := &cloudv1.RestoreList{}
	err := r.Client.List(ctx, restoreList, client.InNamespace(namespace), listOption)
	if err != nil {
		// can definitely get a value, so errors are not returned
		klog.Errorln(err)
	}
	return *restoreList
}

func (r *RestoreResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *RestoreResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}

func (r *RestoreResource) PatchStatus(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}

func (r *RestoreResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	restore := obj.(cloudv1.Restore)
	err := r.Client.Status().Update(ctx, &restore)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *RestoreResource) Delete(ctx context.Context, obj interface{}) error {
	Restore := obj.(cloudv1.Restore)
	err := r.Client.Delete(ctx, &Restore)
	if err != nil {
		r.Recorder.Eventf(&Restore, corev1.EventTypeWarning, FailedToDeleteRestore, "delete Restore"+Restore.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(Restore.Kind, Restore.Name, "delete", "succeed")
	r.Recorder.Event(&Restore, corev1.EventTypeNormal, DeletedRestore, "delete Restore"+Restore.Name)
	return nil
}
