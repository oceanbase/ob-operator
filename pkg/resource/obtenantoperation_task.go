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
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

func (m *ObTenantOperationManager) ChangeTenantRootPassword() error {
	con, err := m.getTenantSysClient(m.Resource.Spec.ChangePwd.Tenant)
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
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		tenant := &v1alpha1.OBTenant{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      m.Resource.Spec.ChangePwd.Tenant,
		}, tenant)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		tenant.Status.Credentials.Root = m.Resource.Spec.ChangePwd.SecretRef
		return m.Client.Status().Update(m.Ctx, tenant)
	})
}

func (m *ObTenantOperationManager) ActivateStandbyTenant() error {
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	err = con.ActivateStandby(m.Resource.Status.PrimaryTenant.Spec.TenantName)
	if err != nil {
		return err
	}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		tenant := &v1alpha1.OBTenant{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      m.Resource.Spec.Failover.StandbyTenant,
		}, tenant)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		tenant.Status.TenantRole = constants.TenantRolePrimary
		return m.Client.Status().Update(m.Ctx, tenant)
	})
}

func (m *ObTenantOperationManager) CreateUsersForActivatedStandby() error {
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		m.Recorder.Event(m.Resource, "Warning", "Can not get cluster operation client", err.Error())
		return err
	}

	// Wait for the tenant to be ready
	maxRetry := 9
	counter := 0
	for counter < maxRetry {
		tenants, err := con.ListTenantWithName(m.Resource.Status.PrimaryTenant.Spec.TenantName)
		if err != nil {
			return err
		}
		if len(tenants) == 0 {
			return errors.New("tenant not found")
		}
		t := tenants[0]
		if t.TenantType == "USER" && t.TenantRole == "PRIMARY" && t.Status == "NORMAL" {
			break
		}
		time.Sleep(9 * time.Second)
		counter++
	}
	if counter >= maxRetry {
		return errors.New("wait for tenant status ready timeout")
	}

	tenantManager := &OBTenantManager{
		Ctx:      m.Ctx,
		Client:   m.Client,
		Recorder: m.Recorder,
		Logger:   m.Logger,
	}
	if m.Resource.Spec.Type == constants.TenantOpSwitchover {
		tenantManager.OBTenant = m.Resource.Status.SecondaryTenant
		tenantManager.OBTenant.ObjectMeta.SetName(m.Resource.Spec.Switchover.StandbyTenant)
	} else {
		tenantManager.OBTenant = m.Resource.Status.PrimaryTenant
		tenantManager.OBTenant.ObjectMeta.SetName(m.Resource.Spec.Failover.StandbyTenant)
	}
	// Hack:
	tenantManager.OBTenant.ObjectMeta.SetNamespace(m.Resource.Namespace)
	// Just reuse the logic of creating users for new coming tenant
	_ = tenantManager.createUserByCredentials()
	return nil
}

func (m *ObTenantOperationManager) SwitchTenantsRole() error {
	// TODO: check whether the two tenants are in the same cluster
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	if m.Resource.Status.Status == constants.TenantOpRunning {
		err = con.SwitchTenantRole(m.Resource.Status.PrimaryTenant.Spec.TenantName, "STANDBY")
		if err != nil {
			return err
		}
		err = con.SwitchTenantRole(m.Resource.Status.SecondaryTenant.Spec.TenantName, "PRIMARY")
		if err != nil {
			return err
		}
	} else if m.Resource.Status.Status == constants.TenantOpReverting {
		err = con.SwitchTenantRole(m.Resource.Status.PrimaryTenant.Spec.TenantName, "PRIMARY")
		if err != nil {
			return err
		}
		err = con.SwitchTenantRole(m.Resource.Status.SecondaryTenant.Spec.TenantName, "STANDBY")
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ObTenantOperationManager) SetTenantLogRestoreSource() error {
	return nil
}

// get operation manager to exec sql
func (m *ObTenantOperationManager) getTenantSysClient(tenantName string) (*operation.OceanbaseOperationManager, error) {
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
	var con *operation.OceanbaseOperationManager
	con, err = GetTenantOperationClient(m.Client, m.Logger, obcluster, tenant.Spec.TenantName, tenant.Status.Credentials.Root)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (m *ObTenantOperationManager) getClusterSysClient(clusterName string) (*operation.OceanbaseOperationManager, error) {
	var err error
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := GetOceanbaseOperationManagerFromOBCluster(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster sys client")
	}
	return con, nil
}
