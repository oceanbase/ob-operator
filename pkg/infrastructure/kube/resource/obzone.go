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

type OBZoneResource struct {
	*Resource
}

func NewOBZoneResource(resource *Resource) ResourceOperator {
	return &OBZoneResource{resource}
}

func (r *OBZoneResource) Create(ctx context.Context, obj interface{}) error {
	obZone := obj.(cloudv1.OBZone)
	// kube.LogForAppActionStatus(obZone.Kind, obZone.Name, "create", obZone)
	err := r.Client.Create(ctx, &obZone)
	if err != nil {
		r.Recorder.Eventf(&obZone, corev1.EventTypeWarning, FailedToCreateOBZone, "create OBZone"+obZone.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(obZone.Kind, obZone.Name, "create", "succeed")
	r.Recorder.Event(&obZone, corev1.EventTypeNormal, CreatedOBZone, "create OBZone"+obZone.Name)
	return nil
}

func (r *OBZoneResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var obZoneCurrent cloudv1.OBZone
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &obZoneCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return obZoneCurrent, err
}

func (r *OBZoneResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *OBZoneResource) Update(ctx context.Context, obj interface{}) error {
	obZone := obj.(cloudv1.OBZone)
	err := r.Client.Update(ctx, &obZone)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *OBZoneResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	obZone := obj.(cloudv1.OBZone)
	err := r.Client.Status().Update(ctx, &obZone)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *OBZoneResource) Delete(ctx context.Context, obj interface{}) error {
	obZone := obj.(cloudv1.OBZone)
	// kube.LogForAppActionStatus(obZone.Kind, obZone.Name, "delete", obZone)
	err := r.Client.Delete(ctx, &obZone)
	if err != nil {
		r.Recorder.Eventf(&obZone, corev1.EventTypeWarning, FailedToDeleteOBZone, "delete OBZone CR"+obZone.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(obZone.Kind, obZone.Name, "delete", "succeed")
	r.Recorder.Event(&obZone, corev1.EventTypeNormal, DeletedOBZone, "delete OBZone CR"+obZone.Name)
	return nil
}
