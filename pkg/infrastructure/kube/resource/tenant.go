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

type TenantResource struct {
	*Resource
}

func NewTenantResource(resource *Resource) ResourceOperator {
	return &TenantResource{resource}
}

func (r *TenantResource) Create(ctx context.Context, obj interface{}) error {
	tenant := obj.(cloudv1.Tenant)
	err := r.Client.Create(ctx, &tenant)
	if err != nil {
		r.Recorder.Eventf(&tenant, corev1.EventTypeWarning, FailedToCreateTenant, "create Tenant"+tenant.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(tenant.Kind, tenant.Name, "create", "succeed")
	r.Recorder.Event(&tenant, corev1.EventTypeNormal, CreatedTenant, "create Tenant"+tenant.Name)
	return nil
}

func (r *TenantResource) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	var TenantCurrent cloudv1.Tenant
	err := r.Client.Get(ctx, kube.GenerateNamespacedName(namespace, name), &TenantCurrent)
	if err != nil {
		klog.Errorln(err)
	}
	return TenantCurrent, err
}

func (r *TenantResource) List(ctx context.Context, namespace string, listOption client.ListOption) interface{} {
	tenantList := &cloudv1.TenantList{}
	err := r.Client.List(ctx, tenantList, client.InNamespace(namespace), listOption)
	if err != nil {
		// can definitely get a value, so errors are not returned
		klog.Errorln(err)
	}
	return *tenantList
}

func (r *TenantResource) Update(ctx context.Context, obj interface{}) error {
	return nil
}

func (r *TenantResource) UpdateStatus(ctx context.Context, obj interface{}) error {
	Tenant := obj.(cloudv1.Tenant)
	err := r.Client.Status().Update(ctx, &Tenant)
	if err != nil {
		klog.Errorln(err)
	}
	return err
}

func (r *TenantResource) Delete(ctx context.Context, obj interface{}) error {
	Tenant := obj.(cloudv1.Tenant)
	// kube.LogForAppActionStatus(Tenant.Kind, Tenant.Name, "delete", Tenant)
	err := r.Client.Delete(ctx, &Tenant)
	if err != nil {
		r.Recorder.Eventf(&Tenant, corev1.EventTypeWarning, FailedToDeleteTenant, "delete Tenant"+Tenant.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(Tenant.Kind, Tenant.Name, "delete", "succeed")
	r.Recorder.Event(&Tenant, corev1.EventTypeNormal, DeletedTenant, "delete Tenant"+Tenant.Name)
	return nil
}

func (r *TenantResource) Patch(ctx context.Context, obj interface{}, patch client.Patch) error {
	tenant := obj.(cloudv1.Tenant)
	err := r.Client.Patch(ctx, &tenant, patch)
	if err != nil {
		r.Recorder.Eventf(&tenant, corev1.EventTypeWarning, FailedToCreatePod, "Patch Tenant"+tenant.Name)
		klog.Errorln(err)
		return err
	}
	kube.LogForAppActionStatus(tenant.Kind, tenant.Name, "Patch", "succeed")
	r.Recorder.Event(&tenant, corev1.EventTypeNormal, PatchedTenant, "Patch tenant"+tenant.Name)
	return nil
}

func (r *TenantResource) PatchStatus(ctx context.Context, obj interface{}, patch client.Patch) error {
	tenant := obj.(cloudv1.Tenant)
	err := r.Client.Status().Patch(ctx, &tenant, patch)
	if err != nil {
		r.Recorder.Eventf(&tenant, corev1.EventTypeWarning, FailedToCreatePod, "Patch Tenant"+tenant.Name)
		klog.Errorln(err)
		return err
	}
	r.Recorder.Event(&tenant, corev1.EventTypeNormal, PatchedTenant, "Patch tenant"+tenant.Name)
	return nil
}
