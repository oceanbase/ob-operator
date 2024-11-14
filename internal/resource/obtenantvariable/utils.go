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

package obtenantvariable

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func (m *OBTenantVariableManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBTenantVariable.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBTenantVariableManager) getOBTenant() (*v1alpha1.OBTenant, error) {
	obtenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBTenantVariable.Spec.OBTenant), obtenant)
	if err != nil {
		return nil, errors.Wrap(err, "get obtenant")
	}
	return obtenant, nil
}

func (m *OBTenantVariableManager) getOceanbaseOperationManager() (*operation.OceanbaseOperationManager, error) {
	obtenant, err := m.getOBTenant()
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, m.generateNamespacedName(obtenant.Spec.ClusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "Get obcluster from K8s")
	}
	return resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, obtenant.Spec.TenantName, obtenant.Status.Credentials.Root)
}

func (m *OBTenantVariableManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		variable := &v1alpha1.OBTenantVariable{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBTenantVariable.GetNamespace(),
			Name:      m.OBTenantVariable.GetName(),
		}, variable)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		variable.Status = *m.OBTenantVariable.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, variable)
	})
}
