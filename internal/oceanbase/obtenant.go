/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
)

func CreateOBTenant(ctx context.Context, tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	return TenantClient.Create(ctx, tenant, metav1.CreateOptions{})
}

func UpdateOBTenant(ctx context.Context, tenant *v1alpha1.OBTenant) (*v1alpha1.OBTenant, error) {
	return TenantClient.Update(ctx, tenant, metav1.UpdateOptions{})
}

func ListAllOBTenants(ctx context.Context, listOptions metav1.ListOptions) (*v1alpha1.OBTenantList, error) {
	list := &v1alpha1.OBTenantList{}
	err := TenantClient.List(ctx, "", list, listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "List all tenants")
	}
	return list, nil
}

func GetOBTenant(ctx context.Context, nn types.NamespacedName) (*v1alpha1.OBTenant, error) {
	return TenantClient.Get(ctx, nn.Namespace, nn.Name, metav1.GetOptions{})
}

func DeleteOBTenant(ctx context.Context, nn types.NamespacedName) error {
	return TenantClient.Delete(ctx, nn.Namespace, nn.Name, metav1.DeleteOptions{})
}

func CreateOBTenantOperation(ctx context.Context, op *v1alpha1.OBTenantOperation) (*v1alpha1.OBTenantOperation, error) {
	return OperationClient.Create(ctx, op, metav1.CreateOptions{})
}

func GetTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) (*v1alpha1.OBTenantBackupPolicy, error) {
	policyListOptions := metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelTenantName + "=" + nn.Name,
	}
	p := &v1alpha1.OBTenantBackupPolicyList{}
	err := BackupPolicyClient.List(ctx, nn.Namespace, p, policyListOptions)
	if err != nil {
		return nil, errors.Wrap(err, "Get tenant backup policy")
	}
	if len(p.Items) == 0 {
		return nil, nil
	}
	return &p.Items[0], nil
}

func CreateTenantBackupPolicy(ctx context.Context, policy *v1alpha1.OBTenantBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	return BackupPolicyClient.Create(ctx, policy, metav1.CreateOptions{})
}

func UpdateTenantBackupPolicy(ctx context.Context, policy *v1alpha1.OBTenantBackupPolicy) (*v1alpha1.OBTenantBackupPolicy, error) {
	return BackupPolicyClient.Update(ctx, policy, metav1.UpdateOptions{})
}

func DeleteTenantBackupPolicy(ctx context.Context, nn types.NamespacedName) error {
	return BackupPolicyClient.Delete(ctx, nn.Namespace, nn.Name, metav1.DeleteOptions{})
}

func ListBackupJobs(ctx context.Context, listOption metav1.ListOptions) (*v1alpha1.OBTenantBackupList, error) {
	list := &v1alpha1.OBTenantBackupList{}
	err := BackupJobClient.List(ctx, "", list, listOption)
	if err != nil {
		return nil, errors.Wrap(err, "List backup jobs")
	}
	return list, nil
}
