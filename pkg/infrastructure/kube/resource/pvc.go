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

type PVCResource struct {
	*Resource
}

func NewPVCResource(resource *Resource) ResourceOperator {
	return &PVCResource{resource}
}

func (r *PVCResource) Create(ctx context.Context, obj interface{}) error {
	pvcs := obj.([]corev1.PersistentVolumeClaim)
	var err error
	for _, pvc := range pvcs {
		// kube.LogForAppActionStatus(pvc.Kind, pvc.Name, "create", pvc)
		err = r.Client.Create(ctx, &pvc)
		if err != nil {
			r.Recorder.Eventf(&pvc, corev1.EventTypeWarning, FailedToCreatePVC, "create PVC"+pvc.Name)
			klog.Errorln(err)
			break
		}
		kube.LogForAppActionStatus(pvc.Kind, pvc.Name, "create", "succeed")
		r.Recorder.Event(&pvc, corev1.EventTypeNormal, CreatedPVC, "create PVC"+pvc.Name)
	}
	return err
}

func (r *PVCResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), pvc)
	if err != nil {
		klog.Errorln(err)
	}
	return *pvc, err
}

func (r *PVCResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *PVCResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *PVCResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	var res error
	return res
}

func (r *PVCResource) Delete(ctx context.Context, obj interface{}) error {
	pvc := obj.(corev1.PersistentVolumeClaim)
	// kube.LogForAppActionStatus(pvc.Kind, pvc.Name, "delete", pvc)
	err := r.Client.Delete(context.TODO(), &pvc)
	if err != nil {
		r.Recorder.Eventf(&pvc, corev1.EventTypeWarning, FailedToDeletePVC, "delete PVC"+pvc.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(pvc.Kind, pvc.Name, "delete", "succeed")
	r.Recorder.Event(&pvc, corev1.EventTypeNormal, DeletedPVC, "delete PVC"+pvc.Name)
	return nil
}

func (r *PVCResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
