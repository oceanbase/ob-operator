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

package obtenantoperation

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	obtenantresource "github.com/oceanbase/ob-operator/internal/resource/obtenant"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func (m *ObTenantOperationManager) ChangeTenantRootPassword() tasktypes.TaskError {
	con, err := m.getTenantRootClient(m.Resource.Spec.ChangePwd.Tenant)
	if err != nil {
		return err
	}
	pwd, err := resourceutils.ReadPassword(m.Client, m.Resource.Namespace, m.Resource.Spec.ChangePwd.SecretRef)
	if err != nil {
		return err
	}
	err = con.ChangeTenantUserPassword(oceanbaseconst.RootUser, pwd)
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

func (m *ObTenantOperationManager) ActivateStandbyTenant() tasktypes.TaskError {
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

func (m *ObTenantOperationManager) CreateUsersForActivatedStandby() tasktypes.TaskError {
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		m.Recorder.Event(m.Resource, "Warning", "Can not get cluster operation client", err.Error())
		return err
	}

	// Wait for the tenant to be ready
	maxRetry := oceanbaseconst.TenantOpRetryTimes
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
		if t.TenantType == "USER" && t.TenantRole == "PRIMARY" && t.SwitchoverStatus == "NORMAL" {
			break
		}
		time.Sleep(oceanbaseconst.TenantOpRetryGapSeconds * time.Second)
		counter++
	}
	if counter >= maxRetry {
		return errors.New("wait for tenant status ready timeout")
	}

	tenantManager := &obtenantresource.OBTenantManager{
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
	_ = tenantManager.CreateUserWithCredentials()
	return nil
}

func (m *ObTenantOperationManager) SwitchTenantsRole() tasktypes.TaskError {
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
		maxRetry := oceanbaseconst.TenantOpRetryTimes
		counter := 0
		for counter < maxRetry {
			primary, err := con.ListTenantWithName(m.Resource.Status.PrimaryTenant.Spec.TenantName)
			if err != nil {
				return err
			}
			if len(primary) == 0 {
				return errors.New("primary tenant not found")
			}
			p := primary[0]
			if p.TenantRole != "STANDBY" || p.SwitchoverStatus != "NORMAL" {
				time.Sleep(oceanbaseconst.TenantOpRetryGapSeconds * time.Second)
				counter++
			} else {
				break
			}
		}
		primary := m.Resource.Status.PrimaryTenant.DeepCopy()
		primary.Status.TenantRole = constants.TenantRoleStandby
		primary.SetName(m.Resource.Spec.Switchover.PrimaryTenant)
		err = m.retryUpdateTenant(primary)
		if err != nil {
			return err
		}
		err = con.SwitchTenantRole(m.Resource.Status.SecondaryTenant.Spec.TenantName, "PRIMARY")
		if err != nil {
			return err
		}
		counter = 0
		for counter < maxRetry {
			standby, err := con.ListTenantWithName(m.Resource.Status.SecondaryTenant.Spec.TenantName)
			if err != nil {
				return err
			}
			if len(standby) == 0 {
				return errors.New("standby tenant not found")
			}
			s := standby[0]
			if s.TenantRole != "PRIMARY" || s.SwitchoverStatus != "NORMAL" {
				time.Sleep(oceanbaseconst.TenantOpRetryGapSeconds * time.Second)
				counter++
			} else {
				break
			}
		}
		standby := m.Resource.Status.SecondaryTenant.DeepCopy()
		standby.Status.TenantRole = constants.TenantRolePrimary
		standby.SetName(m.Resource.Spec.Switchover.StandbyTenant)
		err = m.retryUpdateTenant(standby)
		if err != nil {
			return err
		}
	} else if m.Resource.Status.Status == constants.TenantOpReverting {
		err = con.SwitchTenantRole(m.Resource.Status.PrimaryTenant.Spec.TenantName, "PRIMARY")
		if err != nil {
			return err
		}
		primary := m.Resource.Status.PrimaryTenant.DeepCopy()
		primary.Status.TenantRole = constants.TenantRolePrimary
		primary.SetName(m.Resource.Spec.Switchover.PrimaryTenant)
		err = m.retryUpdateTenant(primary)
		if err != nil {
			return err
		}
		err = con.SwitchTenantRole(m.Resource.Status.SecondaryTenant.Spec.TenantName, "STANDBY")
		if err != nil {
			return err
		}
		standby := m.Resource.Status.SecondaryTenant.DeepCopy()
		standby.Status.TenantRole = constants.TenantRoleStandby
		standby.SetName(m.Resource.Spec.Switchover.StandbyTenant)
		err = m.retryUpdateTenant(standby)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ObTenantOperationManager) SetTenantLogRestoreSource() tasktypes.TaskError {
	var err error
	if m.Resource.Status.Status == constants.TenantOpRunning {
		originStandby := m.Resource.Status.SecondaryTenant.DeepCopy()
		originStandby.SetName(m.Resource.Spec.Switchover.StandbyTenant)
		originStandby.SetNamespace(m.Resource.GetNamespace())
		tenantManager := &obtenantresource.OBTenantManager{
			Ctx:      m.Ctx,
			Client:   m.Client,
			Recorder: m.Recorder,
			Logger:   m.Logger,
			OBTenant: originStandby,
		}
		err = tenantManager.CreateUserWithCredentials()
		if err != nil {
			return err
		}

		con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
		if err != nil {
			return err
		}
		restoreSource, err := resourceutils.GetTenantRestoreSource(m.Ctx, m.Client, m.Logger, con, m.Resource.Namespace, m.Resource.Spec.Switchover.StandbyTenant)
		if err != nil {
			return err
		}
		err = con.SetParameter("LOG_RESTORE_SOURCE", restoreSource, &param.Scope{
			Name:  "TENANT",
			Value: m.Resource.Status.PrimaryTenant.Spec.TenantName,
		})
		if err != nil {
			m.Logger.Error(err, "Failed to set log restore source of original primary tenant")
			return err
		}
	}
	return nil
}

// get operation manager to exec sql
func (m *ObTenantOperationManager) getTenantRootClient(tenantName string) (*operation.OceanbaseOperationManager, error) {
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
	con, err = resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, tenant.Spec.TenantName, tenant.Status.Credentials.Root)
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
	con, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster sys client")
	}
	return con, nil
}

func (m *ObTenantOperationManager) retryUpdateTenant(obj *v1alpha1.OBTenant) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		tenant := &v1alpha1.OBTenant{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.Namespace,
			Name:      obj.Name,
		}, tenant)
		if err != nil {
			return errors.Wrap(err, "get tenant")
		}
		tenant.Status = obj.Status
		return m.Client.Status().Update(m.Ctx, tenant)
	})
}

func (m *ObTenantOperationManager) UpgradeTenant() tasktypes.TaskError {
	targetTenant := m.Resource.Status.PrimaryTenant
	con, err := m.getClusterSysClient(targetTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	var sysCompatible string
	var restoredCompatible string

	compatibles, err := con.SelectCompatibleOfTenants()
	if err != nil {
		return err
	}
	for _, p := range compatibles {
		if p.TenantID == 1 {
			sysCompatible = p.Value
		}
		if p.TenantID == int64(targetTenant.Status.TenantRecordInfo.TenantID) {
			restoredCompatible = p.Value
		}
	}

	if sysCompatible >= "4.1.0.0" && restoredCompatible < sysCompatible {
		err := con.UpgradeTenantWithName(targetTenant.Spec.TenantName)
		if err != nil {
			return err
		}
		maxWait5secTimes := oceanbaseconst.DefaultStateWaitTimeout/5 + 1
	outer:
		for i := 0; i < maxWait5secTimes; i++ {
			time.Sleep(5 * time.Second)
			params, err := con.ListParametersWithTenantID(int64(targetTenant.Status.TenantRecordInfo.TenantID))
			if err != nil {
				return err
			}
			for _, p := range params {
				if p.Name == "compatible" && p.Value == sysCompatible {
					break outer
				}
			}
		}
	} else if sysCompatible < "4.1.0.0" {
		return errors.New("The cluster is of version less than 4.1.0.0, which does not support tenant upgrade")
	} else if restoredCompatible >= sysCompatible {
		return errors.New("The version of target tenant is greater than the cluster")
	}
	return nil
}

func (m *ObTenantOperationManager) ReplayLogOfStandby() tasktypes.TaskError {
	targetTenant := m.Resource.Status.PrimaryTenant
	if targetTenant.Status.TenantRole != constants.TenantRoleStandby {
		return errors.New("The target tenant is not standby")
	}
	con, err := m.getClusterSysClient(targetTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	replayUntil := m.Resource.Spec.ReplayUntil
	if replayUntil == nil || replayUntil.Unlimited {
		err = con.ReplayStandbyLog(targetTenant.Spec.TenantName, "UNLIMITED")
	} else if replayUntil.Timestamp != nil {
		err = con.ReplayStandbyLog(targetTenant.Spec.TenantName, fmt.Sprintf("TIME='%s'", *replayUntil.Timestamp))
	} else if replayUntil.Scn != nil {
		err = con.ReplayStandbyLog(targetTenant.Spec.TenantName, fmt.Sprintf("SCN=%s", *replayUntil.Scn))
	} else {
		return errors.New("Replay until with limit must have a limit key, scn and timestamp are both nil now")
	}
	if err != nil {
		return err
	}
	return nil
}
