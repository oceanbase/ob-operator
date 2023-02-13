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

type TenantBackupResource struct {
	*Resource
}

func NewTenantBackupResource(resource *Resource) ResourceOperator {
	return &TenantBackupResource{resource}
}

func (r *TenantBackupResource) Create(ctx context.Context, obj interface{}) error {
	tenantBackup := obj.(cloudv1.TenantBackup)
	err := r.Client.Create(ctx, &tenantBackup)
	if err != nil {
		r.Recorder.Eventf(&tenantBackup, corev1.EventTypeWarning, FailedToCreateTenantBackup, "create TenantBackup"+tenantBackup.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(tenantBackup.Kind, tenantBackup.Name, "create", "succeed")
	r.Recorder.Event(&tenantBackup, corev1.EventTypeNormal, CreatedTenantBackup, "create TenantBackup"+tenantBackup.Name)
	return nil
}

func (r *TenantBackupResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var tenantBackupCurrent cloudv1.TenantBackup
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &tenantBackupCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return tenantBackupCurrent, err
}

func (r *TenantBackupResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	var res interface{}
	return res
}

func (r *TenantBackupResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *TenantBackupResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	tenantBackup := obj.(cloudv1.TenantBackup)
	err := r.Client.Status().Update(ctx, &tenantBackup)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *TenantBackupResource) Delete(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *TenantBackupResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	return nil
}
