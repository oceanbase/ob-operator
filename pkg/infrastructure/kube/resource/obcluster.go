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

type OBClusterResource struct {
	*Resource
}

func NewOBClusterResource(resource *Resource) ResourceOperator {
	return &OBClusterResource{resource}
}

// TODO: implement operations
func (r *OBClusterResource) Create(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *OBClusterResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var obClusterCurrent cloudv1.OBCluster
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &obClusterCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return obClusterCurrent, err
}

func (r *OBClusterResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *OBClusterResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *OBClusterResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	obCluster := obj.(cloudv1.OBCluster)
	err := r.Client.Status().Update(ctx, &obCluster)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *OBClusterResource) Delete(ctx context.Context, obj interface{}) error {
	obCluster := obj.(cloudv1.OBCluster)
	// kube.LogForAppActionStatus(obCluster.Kind, obCluster.Name, "delete", obCluster)
	err := r.Client.Delete(ctx, &obCluster)
	if err != nil {
		r.Recorder.Eventf(&obCluster, corev1.EventTypeWarning, FailedToDeleteOBCluster, "delete OBCluster"+obCluster.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(obCluster.Kind, obCluster.Name, "delete", "succeed")
	r.Recorder.Event(&obCluster, corev1.EventTypeNormal, DeletedOBCluster, "delete OBCluster"+obCluster.Name)
	return nil
}

func (r *OBClusterResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
