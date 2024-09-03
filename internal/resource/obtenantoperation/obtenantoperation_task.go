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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
	obtenantresource "github.com/oceanbase/ob-operator/internal/resource/obtenant"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*ObTenantOperationManager]()

func ChangeTenantRootPassword(m *ObTenantOperationManager) tasktypes.TaskError {
	con, err := m.getTenantRootClient(m.Resource.Spec.ChangePwd.Tenant)
	if err != nil {
		return err
	}
	pwd, err := resourceutils.ReadPassword(m.Client, m.Resource.Namespace, m.Resource.Spec.ChangePwd.SecretRef)
	if err != nil {
		return err
	}
	err = con.ChangeTenantUserPassword(m.Ctx, oceanbaseconst.RootUser, pwd)
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

func ActivateStandbyTenant(m *ObTenantOperationManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	err = con.ActivateStandby(m.Ctx, m.Resource.Status.PrimaryTenant.Spec.TenantName)
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

func CreateUsersForActivatedStandby(m *ObTenantOperationManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		m.Recorder.Event(m.Resource, "Warning", "Can not get cluster operation client", err.Error())
		return err
	}

	// Wait for the tenant to be ready
	maxRetry := obcfg.GetConfig().Time.TenantOpRetryTimes
	counter := 0
	for counter < maxRetry {
		tenants, err := con.ListTenantWithName(m.Ctx, m.Resource.Status.PrimaryTenant.Spec.TenantName)
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
		time.Sleep(time.Duration(obcfg.GetConfig().Time.TenantOpRetryGapSeconds) * time.Second)
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
	_ = obtenantresource.CreateUserWithCredentials(tenantManager)
	return nil
}

func SwitchTenantsRole(m *ObTenantOperationManager) tasktypes.TaskError {
	// TODO: check whether the two tenants are in the same cluster
	con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	if m.Resource.Status.Status == constants.TenantOpRunning {
		err = con.SwitchTenantRole(m.Ctx, m.Resource.Status.PrimaryTenant.Spec.TenantName, "STANDBY")
		if err != nil {
			return err
		}
		maxRetry := obcfg.GetConfig().Time.TenantOpRetryTimes
		counter := 0
		for counter < maxRetry {
			primary, err := con.ListTenantWithName(m.Ctx, m.Resource.Status.PrimaryTenant.Spec.TenantName)
			if err != nil {
				return err
			}
			if len(primary) == 0 {
				return errors.New("primary tenant not found")
			}
			p := primary[0]
			if p.TenantRole != "STANDBY" || p.SwitchoverStatus != "NORMAL" {
				time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.TenantOpRetryGapSeconds))
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
		err = con.SwitchTenantRole(m.Ctx, m.Resource.Status.SecondaryTenant.Spec.TenantName, "PRIMARY")
		if err != nil {
			return err
		}
		counter = 0
		for counter < maxRetry {
			standby, err := con.ListTenantWithName(m.Ctx, m.Resource.Status.SecondaryTenant.Spec.TenantName)
			if err != nil {
				return err
			}
			if len(standby) == 0 {
				return errors.New("standby tenant not found")
			}
			s := standby[0]
			if s.TenantRole != "PRIMARY" || s.SwitchoverStatus != "NORMAL" {
				time.Sleep(time.Second * time.Duration(obcfg.GetConfig().Time.TenantOpRetryGapSeconds))
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
		err = con.SwitchTenantRole(m.Ctx, m.Resource.Status.PrimaryTenant.Spec.TenantName, "PRIMARY")
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
		err = con.SwitchTenantRole(m.Ctx, m.Resource.Status.SecondaryTenant.Spec.TenantName, "STANDBY")
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

func SetTenantLogRestoreSource(m *ObTenantOperationManager) tasktypes.TaskError {
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
		err = obtenantresource.CreateUserWithCredentials(tenantManager)
		if err != nil {
			return err
		}

		con, err := m.getClusterSysClient(m.Resource.Status.PrimaryTenant.Spec.ClusterName)
		if err != nil {
			return err
		}
		restoreSource, err := resourceutils.GetTenantRestoreSource(m.Ctx, m.Client, m.Logger, m.Resource.Namespace, m.Resource.Spec.Switchover.StandbyTenant)
		if err != nil {
			return err
		}
		err = con.SetParameter(m.Ctx, "LOG_RESTORE_SOURCE", restoreSource, &param.Scope{
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

func UpgradeTenant(m *ObTenantOperationManager) tasktypes.TaskError {
	targetTenant := m.Resource.Status.PrimaryTenant
	con, err := m.getClusterSysClient(targetTenant.Spec.ClusterName)
	if err != nil {
		return err
	}
	var sysCompatible string
	var restoredCompatible string

	compatibles, err := con.SelectCompatibleOfTenants(m.Ctx)
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
		err := con.UpgradeTenantWithName(m.Ctx, targetTenant.Spec.TenantName)
		if err != nil {
			return err
		}
		maxWait5secTimes := obcfg.GetConfig().Time.DefaultStateWaitTimeout/5 + 1
	outer:
		for i := 0; i < maxWait5secTimes; i++ {
			time.Sleep(5 * time.Second)
			params, err := con.ListParametersWithTenantID(m.Ctx, int64(targetTenant.Status.TenantRecordInfo.TenantID))
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

func ReplayLogOfStandby(m *ObTenantOperationManager) tasktypes.TaskError {
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
		err = con.ReplayStandbyLog(m.Ctx, targetTenant.Spec.TenantName, "UNLIMITED")
	} else if replayUntil.Timestamp != nil {
		err = con.ReplayStandbyLog(m.Ctx, targetTenant.Spec.TenantName, fmt.Sprintf("TIME='%s'", *replayUntil.Timestamp))
	} else if replayUntil.Scn != nil {
		err = con.ReplayStandbyLog(m.Ctx, targetTenant.Spec.TenantName, fmt.Sprintf("SCN=%s", *replayUntil.Scn))
	} else {
		return errors.New("Replay until with limit must have a limit key, scn and timestamp are both nil now")
	}
	if err != nil {
		return err
	}
	return nil
}

func UpdateOBTenantResource(m *ObTenantOperationManager) tasktypes.TaskError {
	obtenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      *m.Resource.Spec.TargetTenant,
	}, obtenant)
	if err != nil {
		return err
	}
	if m.Resource.Spec.Force {
		obtenant.Status.Status = tenantstatus.Running
		obtenant.Status.OperationContext = nil
	} else if obtenant.Status.Status != tenantstatus.Running {
		return errors.New("obtenant is not running")
	}
	origin := obtenant.DeepCopy()
	switch m.Resource.Spec.Type {
	case constants.TenantOpSetUnitNumber:
		obtenant.Spec.UnitNumber = m.Resource.Spec.UnitNumber
	case constants.TenantOpSetForceDelete:
		obtenant.Spec.ForceDelete = *m.Resource.Spec.ForceDelete
	case constants.TenantOpSetConnectWhiteList:
		obtenant.Spec.ConnectWhiteList = m.Resource.Spec.ConnectWhiteList
	case constants.TenantOpSetCharset:
		obtenant.Spec.Charset = m.Resource.Spec.Charset
	case constants.TenantOpAddResourcePools:
		for _, pool := range m.Resource.Spec.AddResourcePools {
			obtenant.Spec.Pools = append(obtenant.Spec.Pools, pool)
		}
	case constants.TenantOpModifyResourcePools:
		modifiedPools := make(map[string]*v1alpha1.ResourcePoolSpec)
		for _, pool := range m.Resource.Spec.ModifyResourcePools {
			modifiedPools[pool.Zone] = &pool
		}
		for i := range obtenant.Spec.Pools {
			pool := obtenant.Spec.Pools[i]
			if modified, ok := modifiedPools[pool.Zone]; ok {
				obtenant.Spec.Pools[i] = *modified
			}
		}
	case constants.TenantOpDeleteResourcePools:
		deletedPools := make(map[string]any)
		for _, pool := range m.Resource.Spec.DeleteResourcePools {
			deletedPools[pool] = struct{}{}
		}
		newPools := make([]v1alpha1.ResourcePoolSpec, 0)
		for i := range obtenant.Spec.Pools {
			pool := obtenant.Spec.Pools[i]
			if _, ok := deletedPools[pool.Zone]; !ok {
				newPools = append(newPools, pool)
			}
		}
		obtenant.Spec.Pools = newPools
	}
	oldResourceVersion := obtenant.ResourceVersion
	err = m.Client.Patch(m.Ctx, obtenant, client.MergeFrom(origin))
	if err != nil {
		m.Logger.Error(err, "Failed to patch obtenant")
		return err
	}
	newResourceVersion := obtenant.ResourceVersion
	if oldResourceVersion == newResourceVersion {
		m.Logger.Info("obcluster not changed")
		return nil
	}
	if m.Resource.Spec.Type == constants.TenantOpSetForceDelete {
		// This type of operation only affects the spec of the CRD, and the status won't change.
		return nil
	}
	notRunningMatcher := func(t *v1alpha1.OBTenant) bool {
		return t.Status.Status != tenantstatus.Running
	}
	return m.waitForOBTenantToBeStatus(obcfg.GetConfig().Time.DefaultStateWaitTimeout, notRunningMatcher)
}

func WaitForOBTenantReturnRunning(m *ObTenantOperationManager) tasktypes.TaskError {
	runningMatcher := func(t *v1alpha1.OBTenant) bool {
		return t.Status.Status == tenantstatus.Running
	}
	return m.waitForOBTenantToBeStatus(obcfg.GetConfig().Time.DefaultStateWaitTimeout, runningMatcher)
}
