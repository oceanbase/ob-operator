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

package resource

import (
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

func (m *ObTenantOperationManager) ChangeTenantRootPassword() error {
	con, err := m.getOperationManager(m.Resource.Spec.ChangePwd.Tenant)
	if err != nil {
		return err
	}
	pwd, err := ReadPassword(m.Client, m.Resource.Namespace, m.Resource.Spec.ChangePwd.SecretRef)
	if err != nil {
		return err
	}
	err = con.ChangeTenantUserPassword("root", pwd)
	if err != nil {
		return err
	}
	return nil
}

func (m *ObTenantOperationManager) CheckTenantCRExistence() error {
	if m.Resource.Status.PrimaryTenant == nil {
		return errors.New("primary tenant not ready")
	} else if m.Resource.Spec.Type == constants.TenantOpSwitchover && m.Resource.Status.SecondaryTenant == nil {
		return errors.New("secondary tenant not ready")
	}
	return nil
}

// get operation manager to exec sql
func (m *ObTenantOperationManager) getOperationManager(tenantName string) (*operation.OceanbaseOperationManager, error) {
	if m.con != nil {
		return m.con, nil
	}
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      tenantName,
	}, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get tenant")
	}
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      tenant.Spec.ClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	m.con = con
	return con, nil
}
