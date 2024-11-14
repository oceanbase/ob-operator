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
package obtenant

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
	cmdconst "github.com/oceanbase/ob-operator/internal/const/cmd"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/helper/converter"
	helpermodel "github.com/oceanbase/ob-operator/pkg/helper/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*OBTenantManager]()

func CheckTenant(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	tenantExist, err := m.tenantExist(tenantName)
	if err != nil {
		m.Logger.Error(err, "Check Whether Tenant exist failed", "tenantName", tenantName)
		return err
	}
	if tenantExist {
		err = errors.New("tenant has exist")
		m.Logger.Error(err, "tenant has exist", "tenantName", tenantName)
		return err
	}
	return nil
}

func CheckPoolAndConfig(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	client, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sys client when checking and applying tenant '%s' pool and config", tenantName)
	}
	params, err := client.GetParameter(m.Ctx, "__min_full_resource_pool_memory", nil)
	if err != nil {
		return errors.Wrapf(err, "Get parameter __min_full_resource_pool_memory when checking and applying tenant '%s' pool and config", tenantName)
	}
	if len(params) == 0 {
		return errors.New("Getting parameter __min_full_resource_pool_memory returns empty result")
	}
	minPoolMemory := params[0].Value
	minPoolMemoryQuant, err := resource.ParseQuantity(minPoolMemory)
	if err != nil {
		return errors.Wrapf(err, "Parse quantity when checking and applying tenant '%s' pool and config", tenantName)
	}
	for _, pool := range m.OBTenant.Spec.Pools {
		unitName := m.generateUnitName(pool.Zone)
		poolName := m.generatePoolName(pool.Zone)
		poolExist, err := m.poolExist(poolName)
		if err != nil {
			m.Logger.Error(err, "Check resource pool exist", "tenantName", tenantName, "poolName", poolName)
			return err
		}
		if poolExist {
			return err
		}

		unitExist, err := m.unitConfigV4Exist(unitName)
		if err != nil {
			m.Logger.Error(err, "Check unit config exist failed", "tenantName", tenantName, "unitName", unitName)
			return err
		}
		if unitExist {
			return err
		}
		if pool.UnitConfig.MemorySize.Cmp(minPoolMemoryQuant) < 0 {
			err = errors.New("pool memory size is less than min_full_resource_pool_memory")
			m.Logger.Error(err, "Check pool memory size", "tenantName", tenantName, "poolName", poolName)
			return err
		}
	}
	return nil
}

func CreateTenantWithClear(m *OBTenantManager) tasktypes.TaskError {
	err := CreateTenantTask(m)
	// clean created resource, restore to the initial state
	if err != nil {
		err := DeleteTenant(m)
		if err != nil {
			err = errors.Wrapf(err, "delete tenant when creating tenant")
			return err
		}
	}
	return err
}

func CreateResourcePoolAndConfig(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName

	for _, pool := range m.OBTenant.Spec.Pools {
		err := m.createUnitAndPoolV4(pool)
		if err != nil {
			m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
			return err
		}
	}
	return nil
}

func AddPool(m *OBTenantManager) tasktypes.TaskError {
	// handle add pool
	poolSpecs := m.getPoolsForAdd()
	for _, addPool := range poolSpecs {
		err := m.tenantAddPool(addPool)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeletePool(m *OBTenantManager) tasktypes.TaskError {
	// handle delete pool
	poolStatuses := m.getPoolsForDelete()
	for _, poolStatus := range poolStatuses {
		err := m.TenantDeletePool(poolStatus)
		if err != nil {
			return err
		}
	}
	return nil
}

func MaintainUnitConfig(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName

	version, err := m.getOBVersion()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed, check and apply unitConfigV4", "tenantName", tenantName)
		return err
	}
	if string(version[0]) == tenant.Version4 {
		return m.CheckAndApplyUnitConfigV4()
	}
	return errors.New("no match version for check and set unit config")
}

func DeleteTenant(m *OBTenantManager) tasktypes.TaskError {
	var err error
	tenantName := m.OBTenant.Spec.TenantName
	err = m.deleteTenant()
	if err != nil {
		return err
	}
	err = m.deletePool()
	if err != nil {
		return err
	}
	err = m.deleteUnitConfig()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete tenant successfully", "tenantName", tenantName)
	return nil
}

func CheckAndApplyCharset(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sql operator when checking and applying tenant '%s' charset ", tenantName)
	}
	specCharset := m.OBTenant.Spec.Charset
	if specCharset == "" {
		specCharset = tenant.Charset
	}
	if specCharset != m.OBTenant.Status.TenantRecordInfo.Charset {
		tenantSQLParam := model.TenantSQLParam{
			TenantName: tenantName,
			Charset:    specCharset,
		}
		err = oceanbaseOperationManager.SetTenant(m.Ctx, tenantSQLParam)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateEmptyStandbyTenant(m *OBTenantManager) tasktypes.TaskError {
	if m.OBTenant.Spec.Source == nil || m.OBTenant.Spec.Source.Tenant == nil {
		return errors.New("Empty standby tenant must have source tenant")
	}
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	ns := m.OBTenant.GetNamespace()
	tenantCRName := *m.OBTenant.Spec.Source.Tenant
	restoreSource, err := resourceutils.GetTenantRestoreSource(m.Ctx, m.Client, m.Logger, ns, tenantCRName)
	if err != nil {
		return err
	}
	poolList := m.generateSpecPoolList(m.OBTenant.Spec.Pools)
	primaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	locality := m.generateLocality(m.OBTenant.Spec.Pools)
	err = con.CreateEmptyStandbyTenant(m.Ctx, &model.CreateEmptyStandbyTenantParam{
		TenantName:    m.OBTenant.Spec.TenantName,
		RestoreSource: restoreSource,
		PrimaryZone:   primaryZone,
		Locality:      locality,
		PoolList:      poolList,
	})
	if err != nil {
		return err
	}
	m.Recorder.Event(m.OBTenant, "CreateEmptyStandby", "", "Succeed to create empty standby tenant")
	return nil
}

func CheckPrimaryTenantLsIntegrity(m *OBTenantManager) tasktypes.TaskError {
	var err error
	if m.OBTenant.Spec.Source == nil || m.OBTenant.Spec.Source.Tenant == nil {
		return errors.New("Primary tenant must have source tenant")
	}
	err = resourceutils.CheckTenantLSIntegrity(m.Ctx, m.Client, m.Logger, m.OBTenant.Namespace, *m.OBTenant.Spec.Source.Tenant)
	if err != nil {
		return errors.Wrap(err, "Check primary tenant LS integrity")
	}
	return nil
}

func CreateTenantRestoreJobCR(m *OBTenantManager) tasktypes.TaskError {
	var existingJobs v1alpha1.OBTenantRestoreList
	var err error

	err = m.Client.List(m.Ctx, &existingJobs,
		client.MatchingLabels{
			oceanbaseconst.LabelRefOBCluster: m.OBTenant.Spec.ClusterName,
			oceanbaseconst.LabelTenantName:   m.OBTenant.Spec.TenantName,
			oceanbaseconst.LabelRefUID:       string(m.OBTenant.GetUID()),
		},
		client.InNamespace(m.OBTenant.Namespace))
	if err != nil {
		return err
	}

	if len(existingJobs.Items) != 0 {
		return errors.New("There is already at least one restore job for this tenant")
	}

	restoreJob := &v1alpha1.OBTenantRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.OBTenant.Name + "-restore",
			Namespace: m.OBTenant.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         m.OBTenant.APIVersion,
				Kind:               m.OBTenant.Kind,
				Name:               m.OBTenant.Name,
				UID:                m.OBTenant.GetUID(),
				BlockOwnerDeletion: resourceutils.GetRef(true)}},
			Labels: map[string]string{
				oceanbaseconst.LabelRefOBCluster: m.OBTenant.Spec.ClusterName,
				oceanbaseconst.LabelTenantName:   m.OBTenant.Spec.TenantName,
				oceanbaseconst.LabelRefUID:       string(m.OBTenant.GetUID()),
			}},
		Spec: v1alpha1.OBTenantRestoreSpec{
			TargetTenant:  m.OBTenant.Spec.TenantName,
			TargetCluster: m.OBTenant.Spec.ClusterName,
			RestoreRole:   m.OBTenant.Spec.TenantRole,
			Source:        *m.OBTenant.Spec.Source.Restore,
			Option:        m.generateRestoreOption(),
			PrimaryTenant: m.OBTenant.Spec.Source.Tenant,
		},
	}
	err = m.Client.Create(m.Ctx, restoreJob)
	if err != nil {
		return err
	}
	return nil
}

func WatchRestoreJobToFinish(m *OBTenantManager) tasktypes.TaskError {
	var err error
	check := func() (bool, error) {
		runningRestore := &v1alpha1.OBTenantRestore{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.OBTenant.GetNamespace(),
			Name:      m.OBTenant.Name + "-restore",
		}, runningRestore)
		if err != nil {
			return false, err
		}
		if runningRestore.Status.Status == constants.RestoreJobSuccessful {
			return true, nil
		} else if runningRestore.Status.Status == constants.RestoreJobFailed {
			m.Recorder.Event(m.OBTenant, "RestoreJobFailed", "", "Restore job failed")
			return false, errors.New("Restore job failed")
		}
		return false, nil
	}
	// Tenant restoring is in common quite a slow process, so we need to wait for a longer time
	err = resourceutils.CheckJobWithTimeout(check, time.Second*time.Duration(obcfg.GetConfig().Time.LocalityChangeTimeoutSeconds))
	if err != nil {
		return errors.Wrap(err, "Failed to wait for restore job to finish")
	}
	tenantWhiteListMap.Store(m.OBTenant.Spec.TenantName, m.OBTenant.Spec.ConnectWhiteList)
	m.Recorder.Event(m.OBTenant, "RestoreJobFinished", "", "Restore job finished successfully")
	return nil
}

func CancelTenantRestoreJob(m *OBTenantManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	err = con.CancelRestoreOfTenant(m.Ctx, m.OBTenant.Spec.TenantName)
	if err != nil {
		return err
	}
	err = m.deletePool()
	if err != nil {
		return err
	}
	err = m.deleteUnitConfig()
	if err != nil {
		return err
	}
	err = m.Client.Delete(m.Ctx, &v1alpha1.OBTenantRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.OBTenant.Name + "-restore",
			Namespace: m.OBTenant.GetNamespace(),
		},
	})
	if err != nil {
		m.Logger.Error(err, "delete restore job CR")
		return err
	}
	err = m.Client.Delete(m.Ctx, m.OBTenant)
	if err != nil {
		m.Logger.Error(err, "delete tenant CR")
	}
	return nil
}

func UpgradeTenantIfNeeded(m *OBTenantManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient()
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
		if p.TenantID == int64(m.OBTenant.Status.TenantRecordInfo.TenantID) {
			restoredCompatible = p.Value
		}
	}
	if sysCompatible >= "4.1.0.0" && restoredCompatible < sysCompatible {
		err := con.UpgradeTenantWithName(m.Ctx, m.OBTenant.Spec.TenantName)
		if err != nil {
			return err
		}
		maxWait5secTimes := obcfg.GetConfig().Time.DefaultStateWaitTimeout/5 + 1
	outer:
		for i := 0; i < maxWait5secTimes; i++ {
			time.Sleep(5 * time.Second)
			params, err := con.ListParametersWithTenantID(m.Ctx, int64(m.OBTenant.Status.TenantRecordInfo.TenantID))
			if err != nil {
				return err
			}
			for _, p := range params {
				if p.Name == "compatible" && p.Value == sysCompatible {
					break outer
				}
			}
		}
	}
	return nil
}

func CheckAndApplyUnitNum(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sql operator when checking and applying unitnum of tenant %s", tenantName)
	}

	if m.OBTenant.Spec.UnitNumber != m.OBTenant.Status.TenantRecordInfo.UnitNumber {
		err = oceanbaseOperationManager.SetTenantUnitNum(m.Ctx, tenantName, m.OBTenant.Spec.UnitNumber)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckAndApplyWhiteList(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sql operator when checking and applying ob_tcp_invited_nodes for tenant %s", tenantName)
	}

	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.TenantRecordInfo.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if statusWhiteList != specWhiteList {
		m.Logger.Info("Found specWhiteList didn't match", "tenantName", tenantName,
			"statusWhiteList", statusWhiteList, "specWhiteList", specWhiteList)
		variableList := m.generateWhiteListInVariableForm(specWhiteList)
		err = oceanbaseOperationManager.SetTenantVariable(m.Ctx, tenantName, variableList)
		if err != nil {
			return err
		}
		// TODO: get whitelist variable by tenant account
		// Because getting a whitelist requires specifying a tenant , temporary use .Status.TenantRecordInfo.ConnectWhiteList as value in db
		tenantWhiteListMap.Store(tenantName, specWhiteList)
	}
	return nil
}

func CheckAndApplyPrimaryZone(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sql operator when processing priority of tenant %s", tenantName)
	}

	specPrimaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	specPrimaryZoneMap := m.generatePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZone := m.generateStatusPrimaryZone(m.OBTenant.Status.Pools)
	statusPrimaryZoneMap := m.generatePrimaryZoneMap(statusPrimaryZone)
	if !reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		tenantSQLParam := model.TenantSQLParam{
			TenantName:  tenantName,
			PrimaryZone: specPrimaryZone,
		}
		err = oceanbaseOperationManager.SetTenant(m.Ctx, tenantSQLParam)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckAndApplyLocality(m *OBTenantManager) tasktypes.TaskError {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrapf(err, "Get sql operator when prcoessing tenant '%s' locality ", tenantName)
	}
	specLocalityMap := m.generateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := m.generateStatusLocalityMap(m.OBTenant.Status.Pools)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		locality := m.generateLocality(m.OBTenant.Spec.Pools)
		tenantSQLParam := model.TenantSQLParam{
			TenantName: tenantName,
			Locality:   locality,
		}
		err = oceanbaseOperationManager.SetTenant(m.Ctx, tenantSQLParam)
		if err != nil {
			return err
		}
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Wait for tenant 'ALTER_TENANT' job of adding pool task", "tenantName", tenantName)
	check := func() (bool, error) {
		exist, err := oceanbaseOperationManager.CheckRsJobExistByTenantID(m.Ctx, m.OBTenant.Status.TenantRecordInfo.TenantID)
		if err != nil {
			return false, errors.Wrapf(err, "Get RsJob %s", tenantName)
		}
		return !exist, nil
	}
	err = resourceutils.CheckJobWithTimeout(check)
	if err != nil {
		return errors.Wrapf(err, "Failed to wait for 'ALTER_TENANT' job of adding pool task to finish %s", tenantName)
	}
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("'ALTER_TENANT' job of adding pool task succeeded", "tenantName", tenantName)
	return nil
}

func OptimizeTenantByScenario(m *OBTenantManager) tasktypes.TaskError {
	// start a job to read optimize parameters, ignore errors, only proceed with valid outputs and ignore the errors
	obcluster, err := m.getOBCluster()
	if err != nil {
		return errors.Wrap(err, "Get obcluster")
	}
	m.Logger.Info("Start to optimize tenant parameter and variable")
	jobName := fmt.Sprintf("optimize-tenant-%s-%s", m.OBTenant.Name, rand.String(6))
	output, code, _ := resourceutils.RunJob(m.Ctx, m.Client, m.Logger, m.OBTenant.Namespace,
		jobName,
		obcluster.Spec.OBServerTemplate.Image,
		obcluster.Spec.OBServerTemplate.PodFields,
		fmt.Sprintf("bin/oceanbase-helper optimize tenant %s", m.OBTenant.Spec.Scenario))
	if code == int32(cmdconst.ExitCodeOK) || code == int32(cmdconst.ExitCodeIgnorableErr) {
		optimizeConfig := &helpermodel.OptimizationResponse{}
		err := json.Unmarshal([]byte(output), optimizeConfig)
		if err != nil {
			m.Logger.Error(err, "Failed to parse optimization config")
		}
		conn, err := m.getTenantClient()
		if err != nil {
			m.Logger.Error(err, "Get tenant operation manager failed")
		}
		for _, parameter := range optimizeConfig.Parameters {
			m.Logger.Info("Set parameter %s to %v", parameter.Name, converter.ConvertFloat(parameter.Value))
			err := conn.SetParameter(m.Ctx, parameter.Name, converter.ConvertFloat(parameter.Value), nil)
			if err != nil {
				m.Logger.Error(err, "Failed to set parameter")
			}
		}
		for _, variable := range optimizeConfig.Variables {
			m.Logger.Info("Set variable %s to %v", variable.Name, converter.ConvertFloat(variable.Value))
			err := conn.SetGlobalVariable(m.Ctx, variable.Name, converter.ConvertFloat(variable.Value))
			if err != nil {
				m.Logger.Error(err, "Failed to set global variable")
			}
		}
	}
	return nil
}

func CreateUserWithCredentialSecrets(m *OBTenantManager) tasktypes.TaskError {
	if m.OBTenant.Spec.TenantRole == constants.TenantRoleStandby {
		// standby tenant can not create users
		return nil
	}
	err := CreateUserWithCredentials(m)
	if err != nil {
		m.Recorder.Event(m.OBTenant, corev1.EventTypeWarning, "Failed to create user or change password", err.Error())
		m.Logger.Error(err, "Failed to create user or change password, please check the credential secrets")
	}

	return nil
}

func MaintainTenantParameters(m *OBTenantManager) tasktypes.TaskError {
	parameterMap := make(map[string]apitypes.Parameter)
	for _, parameter := range m.OBTenant.Status.Parameters {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Build parameter map", "parameter", parameter.Name)
		parameterMap[parameter.Name] = parameter
	}
	for _, parameter := range m.OBTenant.Spec.Parameters {
		parameterStatus, parameterExists := parameterMap[parameter.Name]
		if !parameterExists {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter not exists, need create", "param", parameter.Name)
			err := m.createOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Create obparameter failed", "param", parameter.Name)
			}
		} else if parameterStatus.Value != parameter.Value {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Parameter value not matched, need update", "param", parameter.Name)
			err := m.updateOBParameter(&parameter)
			if err != nil {
				// since parameter is not a big problem, just log the error
				m.Logger.Error(err, "Update obparameter failed", "param", parameter.Name)
			}
		}
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Remove parameter from map", "parameter", parameter.Name)
		delete(parameterMap, parameter.Name)
	}

	// delete parameters that not in spec definition
	for _, parameter := range parameterMap {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Delete parameter", "parameter", parameter.Name)
		err := m.deleteOBParameter(&parameter)
		if err != nil {
			m.Logger.Error(err, "Failed to delete parameter")
		}
	}
	return nil
}
