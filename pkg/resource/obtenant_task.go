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
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

// ---------- task entry point ----------

func (m *OBTenantManager) CreateTenantTaskWithClear() error {
	err := m.CreateTenantTask()
	// clean created resource, restore to the initial state
	if err != nil {
		err := m.DeleteTenantTask()
		if err != nil {
			err = errors.Wrapf(err, "delete tenant when creating tenant")
			return err
		}
	}
	return err
}

func (m *OBTenantManager) CreateResourcePoolAndConfigTaskWithClear() error {
	err := m.CreateResourcePoolAndConfigTask()
	// clean created resource, restore to the initial state
	if err != nil {
		err := m.DeleteTenantTask()
		if err != nil {
			err = errors.Wrapf(err, "delete tenant when creating tenant")
			return err
		}
	}
	return err
}

func (m *OBTenantManager) CreateTenantTask() error {
	tenantName := m.OBTenant.Spec.TenantName
	err := m.createTenant()
	if err != nil {
		m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
		return err
	}
	return nil
}

func (m *OBTenantManager) CreateResourcePoolAndConfigTask() error {
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

func (m *OBTenantManager) CheckTenantTask() error {
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

func (m *OBTenantManager) CheckPoolAndConfigTask() error {
	tenantName := m.OBTenant.Spec.TenantName

	for _, pool := range m.OBTenant.Spec.Pools {
		unitName := m.generateUnitName(pool.Zone)
		poolName := m.generatePoolName(pool.Zone)
		poolExist, err := m.poolExist(poolName)
		if err != nil {
			m.Logger.Error(err, "Check Resource Pool Exist", "tenantName", tenantName, "poolName", poolName)
			return err
		}
		if poolExist {
			return err
		}

		unitExist, err := m.unitConfigV4Exist(unitName)
		if err != nil {
			m.Logger.Error(err, "Check UnitConfig Exist Failed", "tenantName", tenantName, "unitName", unitName)
			return err
		}
		if unitExist {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) MaintainWhiteListTask() error {
	tenantName := m.OBTenant.Spec.TenantName
	err := m.CheckAndApplyWhiteList()
	if err != nil {
		m.Logger.Error(err, "maintain tenant, check and set whitelist (tcp invited node)", "tenantName", tenantName)
		return err
	}
	return nil
}

func (m *OBTenantManager) AddPoolTask() error {
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

func (m *OBTenantManager) DeletePoolTask() error {
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

func (m *OBTenantManager) MaintainUnitConfigTask() error {
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

func (m *OBTenantManager) DeleteTenantTask() error {
	var err error
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Delete Tenant", "tenantName", tenantName)
	err = m.deleteTenant()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Pool", "tenantName", tenantName)
	err = m.deletePool()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Unit", "tenantName", tenantName)
	err = m.deleteUnitConfig()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Tenant Success", "tenantName", tenantName)
	return nil
}

func (m *OBTenantManager) AddFinalizerTask() error {
	return nil
}

// ---------- Check And Apply function ----------

func (m *OBTenantManager) CheckAndApplyWhiteList() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Applying ob_tcp_invited_nodes For Tenant ", tenantName))
	}

	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.TenantRecordInfo.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if statusWhiteList != specWhiteList {
		m.Logger.Info("found specWhiteList didn't match", "tenantName", tenantName,
			"statusWhiteList", statusWhiteList, "specWhiteList", specWhiteList)
		variableList := m.generateWhiteListInVariableForm(specWhiteList)
		err = oceanbaseOperationManager.SetTenantVariable(tenantName, variableList)
		if err != nil {
			return err
		}
		// TODO get whitelist variable by tenant account
		// Because getting a whitelist requires specifying a tenant , temporary use .Status.TenantRecordInfo.ConnectWhiteList as value in db
		GlobalWhiteListMap[tenantName] = specWhiteList
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyUnitConfigV4() error {
	tenantName := m.OBTenant.Spec.TenantName
	specUnitConfigMap := m.generateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := m.GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		match := true
		specUnitConfig := specUnitConfigMap[pool.Zone]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.Zone]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist {
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			m.Logger.Info("found unit config v4 didn't match", "tenantName", tenantName, "zoneName", pool.Zone,
				"statusUnitConfig", FormatUnitConfigV4(statusUnitConfigMap[pool.Zone]), "specUnitConfig", FormatUnitConfigV4(specUnitConfigMap[pool.Zone]))
			match = false
		}
		if !match {
			unitName := m.generateUnitName(pool.Zone)
			err := m.setUnitConfigV4(unitName, specUnitConfigMap[pool.Zone])
			if err != nil {
				m.Logger.Error(err, "Set Tenant Unit failed", "tenantName", tenantName, "unitName", unitName)
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyUnitNum() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Applying Tenant UnitNum", tenantName))
	}

	if m.OBTenant.Spec.UnitNumber != m.OBTenant.Status.TenantRecordInfo.UnitNumber {
		err = oceanbaseOperationManager.SetTenantUnitNum(tenantName, m.OBTenant.Spec.UnitNumber)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyPrimaryZone() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Priority ", tenantName))
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
		err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyLocality() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Locality ", tenantName))
	}
	specLocalityMap := m.generateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := m.generateStatusLocalityMap(m.OBTenant.Status.Pools)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		locality := m.generateLocality(m.OBTenant.Spec.Pools)
		tenantSQLParam := model.TenantSQLParam{
			TenantName: tenantName,
			Locality:   locality,
		}
		err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
		if err != nil {
			return err
		}
	}
	m.Logger.Info("Wait For Tenant 'ALTER_TENANT' Job for addPool Finished", "tenantName", tenantName)
	for {
		exist, err := oceanbaseOperationManager.CheckRsJobExistByTenantID(m.OBTenant.Status.TenantRecordInfo.TenantID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Get RsJob %s", tenantName))
		}
		if !exist {
			break
		}
		time.Sleep(config.PollingJobSleepTime)
	}
	m.Logger.Info("'ALTER_TENANT' Job for addPool successes", "tenantName", tenantName)
	return nil
}

func (m *OBTenantManager) CheckAndApplyCharset() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Checking and Applying Tenant '%s' Charset ", tenantName))
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
		err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------- action function ----------

func (m *OBTenantManager) createTenant() error {
	tenantName := m.OBTenant.Spec.TenantName
	pools := m.OBTenant.Spec.Pools
	m.Logger.Info("Create Tenant", "tenantName", tenantName)
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Tenant")
	}

	tenantSQLParam := model.TenantSQLParam{
		TenantName:   tenantName,
		PrimaryZone:  m.generateSpecPrimaryZone(pools),
		VariableList: m.generateWhiteListInVariableForm(m.OBTenant.Spec.ConnectWhiteList),
		Charset:      m.OBTenant.Spec.Charset,
		PoolList:     m.generateSpecPoolList(pools),
		Locality:     m.generateLocality(pools),
		Collate:      m.OBTenant.Spec.Collate,
	}
	if tenantSQLParam.Charset == "" {
		tenantSQLParam.Charset = tenant.Charset
	}

	err = oceanbaseOperationManager.AddTenant(tenantSQLParam)
	if err != nil {
		m.Telemetry.Event(m.OBTenant, corev1.EventTypeWarning, "failed to create OBTenant", err.Error())
		return err
	}
	GlobalWhiteListMap[tenantName] = m.OBTenant.Spec.ConnectWhiteList
	// Create user or change password of root, do not return error
	m.Telemetry.Event(m.OBTenant, "Create", "", "create OBTenant successfully")
	return nil
}

func (m *OBTenantManager) createUnitConfigV4(unitName string, unitConfig *v1alpha1.UnitConfig) error {
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Create UnitConfig", "tenantName", tenantName, "unitName", unitName)
	unitModel := m.generateModelUnitConfigV4SQLParam(unitName, m.generateModelUnitConfigV4(unitConfig))
	if unitModel.MemorySize == 0 {
		err := errors.New("unit memorySize is empty")
		m.Logger.Error(err, "unit memorySize cannot be zero", "tenantName", tenantName, "unitName", unitName)
		return err
	}
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource UnitConfigV4")
	}

	return oceanbaseOperationManager.AddUnitConfigV4(unitModel)
}

func (m *OBTenantManager) setUnitConfigV4(unitName string, unitConfig *model.UnitConfigV4) error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	unitModel := m.generateModelUnitConfigV4SQLParam(unitName, unitConfig)
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", tenantName))
	}
	return oceanbaseOperationManager.SetUnitConfigV4(unitModel)
}

func (m *OBTenantManager) getPoolsForAdd() []v1alpha1.ResourcePoolSpec {
	var pools []v1alpha1.ResourcePoolSpec
	for _, specZone := range m.OBTenant.Spec.Pools {
		exist := false
		for _, statusZone := range m.OBTenant.Status.Pools {
			if statusZone.ZoneList == specZone.Zone {
				exist = true
			}
		}
		if !exist {
			pools = append(pools, specZone)
		}
	}
	return pools
}

func (m *OBTenantManager) getPoolsForDelete() []v1alpha1.ResourcePoolStatus {
	var poolStatuses []v1alpha1.ResourcePoolStatus
	for _, statusPool := range m.OBTenant.Status.Pools {
		exist := false
		for _, specPool := range m.OBTenant.Spec.Pools {
			if statusPool.ZoneList == specPool.Zone {
				exist = true
			}
		}
		if !exist {
			poolStatuses = append(poolStatuses, statusPool)
		}
	}
	return poolStatuses
}

func (m *OBTenantManager) tenantAddPool(poolAdd v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Add Pool", tenantName))
	}

	// step 1: create unit and poolAdd
	err = m.createUnitAndPoolV4(poolAdd)
	if err != nil {
		return err
	}

	// step 2.1: update locality and resource poolAdd list
	poolStatusAdd := v1alpha1.ResourcePoolStatus{
		ZoneList:   poolAdd.Zone,
		Type:       poolAdd.Type,
		UnitNumber: m.OBTenant.Spec.UnitNumber,
	}
	resourcePoolStatusList := m.OBTenant.Status.Pools
	resourcePoolStatusList = append(resourcePoolStatusList, poolStatusAdd)
	statusLocalityMap := m.generateStatusLocalityMap(resourcePoolStatusList)
	localityList := m.generateLocalityList(statusLocalityMap)
	poolList := m.generateStatusPoolList(resourcePoolStatusList)
	specPrimaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)

	tenantSQLParam := model.TenantSQLParam{
		TenantName:  tenantName,
		Locality:    strings.Join(localityList, ","),
		PoolList:    poolList,
		PrimaryZone: specPrimaryZone,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		return err
	}

	// step 2.2: Wait for task finished
	m.Logger.Info("Wait For Tenant 'ALTER_TENANT' Job for addPool Finished", "tenantName", tenantName)
	for {
		exist, err := oceanbaseOperationManager.CheckRsJobExistByTenantID(m.OBTenant.Status.TenantRecordInfo.TenantID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Get RsJob %s", tenantName))
		}
		if !exist {
			break
		}
		time.Sleep(config.PollingJobSleepTime)
	}
	m.Logger.Info("'ALTER_TENANT' Job for addPool successes", "tenantName", tenantName)

	m.Logger.Info("Succeed add poolAdd", "deleted poolName", poolAdd.Zone)
	return nil
}

func (m *OBTenantManager) TenantDeletePool(poolDelete v1alpha1.ResourcePoolStatus) error {
	tenantName := m.OBTenant.Spec.TenantName
	poolName := m.generatePoolName(poolDelete.ZoneList)
	unitName := m.generateUnitName(poolDelete.ZoneList)

	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Delete Pool ", tenantName))
	}
	var zoneList []v1alpha1.ResourcePoolStatus
	for _, zone := range m.OBTenant.Status.Pools {
		if zone.ZoneList != poolDelete.ZoneList {
			zoneList = append(zoneList, zone)
		}
	}

	statusLocalityMap := m.generateStatusLocalityMap(zoneList)
	localityList := m.generateLocalityList(statusLocalityMap)
	poolList := m.generateStatusPoolList(zoneList)
	specPrimaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)

	// step 1.1: update locality
	// note：this operator is async in oceanbase, polling until update locality task success
	tenantSQLParam := model.TenantSQLParam{
		TenantName: tenantName,
		Locality:   strings.Join(localityList, ","),
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		m.Logger.Error(err, "Modify Tenant, update locality", "tenantName", tenantName)
		return err
	}

	m.Logger.Info("Wait For Tenant 'ALTER_TENANT' Job for deletePool Finished", "tenantName", tenantName)

	for {
		exist, err := oceanbaseOperationManager.CheckRsJobExistByTenantID(m.OBTenant.Status.TenantRecordInfo.TenantID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Get RsJob %s", tenantName))
		}
		if !exist {
			break
		}
		time.Sleep(config.PollingJobSleepTime)
	}

	// step 1.2: update resource pool list
	tenantSQLParam = model.TenantSQLParam{
		TenantName:  tenantName,
		PoolList:    poolList,
		PrimaryZone: specPrimaryZone,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		m.Logger.Error(err, "Modify Tenant, update poolList", "tenantName", tenantName)
		return err
	}
	// step 2: delete resource pool
	poolExist, err := m.poolExist(poolName)
	if err != nil {
		m.Logger.Error(err, "Check ResourcePool exist", "poolName", poolName)
		return err
	}
	if poolExist {
		err = oceanbaseOperationManager.DeletePool(poolName)
		if err != nil {
			return err
		}
	}

	// step 3: delete unit
	unitExist, err := m.unitConfigV4Exist(unitName)
	if err != nil {
		m.Logger.Error(err, "Check UnitConfigV4 Exist", "unitName", unitName)
		return err
	}
	if unitExist {
		err = oceanbaseOperationManager.DeleteUnitConfig(unitName)
		if err != nil {
			return err
		}
	}
	m.Logger.Info("Succeed delete pool", "deleted poolName", poolDelete.ZoneList)
	return nil
}

// ---------- compare helper function ----------

func IsUnitConfigV4Equal(specUnitConfig *model.UnitConfigV4, statusUnitConfig *model.UnitConfigV4) bool {
	if specUnitConfig.MaxCPU == statusUnitConfig.MaxCPU &&
		specUnitConfig.MemorySize == statusUnitConfig.MemorySize {
		if (specUnitConfig.MinIops != 0 && specUnitConfig.MinIops != statusUnitConfig.MinIops) ||
			(specUnitConfig.MaxIops != 0 && specUnitConfig.MaxIops != statusUnitConfig.MaxIops) ||
			(specUnitConfig.MinCPU != 0 && specUnitConfig.MinCPU != statusUnitConfig.MinCPU) ||
			(specUnitConfig.LogDiskSize != 0 && specUnitConfig.LogDiskSize != statusUnitConfig.LogDiskSize) ||
			(specUnitConfig.IopsWeight != 0 && specUnitConfig.IopsWeight != statusUnitConfig.IopsWeight) {
			return false
		}
		return true
	}
	return false
}

func FormatUnitConfigV4(unit *model.UnitConfigV4) string {
	return fmt.Sprintf("MaxCPU: %f MinCPU:%f MemorySize:%d MaxIops:%d MinIops:%d IopsWeight:%d LogDiskSize:%d",
		unit.MaxCPU, unit.MinCPU, unit.MemorySize, unit.MaxIops, unit.MinIops, unit.IopsWeight, unit.LogDiskSize)
}

// ---------- generate "zoneName-value" map function ----------

func (m *OBTenantManager) generatePrimaryZoneMap(str string) map[int][]string {
	res := make(map[int][]string, 0)
	levelCuts := strings.Split(str, ";")
	for idx, levelCut := range levelCuts {
		cut := strings.Split(levelCut, ",")
		res[idx] = cut
		sort.Strings(res[idx])
	}
	return res
}

func (m *OBTenantManager) generateSpecUnitConfigV4Map(spec v1alpha1.OBTenantSpec) map[string]*model.UnitConfigV4 {
	var unitConfigMap = make(map[string]*model.UnitConfigV4, 0)
	for _, pool := range spec.Pools {
		unitConfigMap[pool.Zone] = m.generateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func (m *OBTenantManager) GenerateStatusUnitConfigV4Map(status v1alpha1.OBTenantStatus) map[string]*model.UnitConfigV4 {
	var unitConfigMap = make(map[string]*model.UnitConfigV4, 0)
	for _, pool := range status.Pools {
		unitConfigMap[pool.ZoneList] = m.generateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func (m *OBTenantManager) generateModelUnitConfigV4(unitConfig *v1alpha1.UnitConfig) *model.UnitConfigV4 {
	var modelUnitConfigV4 model.UnitConfigV4
	modelUnitConfigV4.MaxCPU = unitConfig.MaxCPU.AsApproximateFloat64()
	modelUnitConfigV4.MinCPU = unitConfig.MinCPU.AsApproximateFloat64()
	modelUnitConfigV4.MaxIops = int64(unitConfig.MaxIops)
	modelUnitConfigV4.MinIops = int64(unitConfig.MinIops)
	modelUnitConfigV4.IopsWeight = int64(unitConfig.IopsWeight)
	modelUnitConfigV4.MemorySize = unitConfig.MemorySize.Value()
	modelUnitConfigV4.LogDiskSize = unitConfig.LogDiskSize.Value()
	return &modelUnitConfigV4
}

func (m *OBTenantManager) generateModelUnitConfigV4SQLParam(unitConfigName string, unitConfig *model.UnitConfigV4) *model.UnitConfigV4SQLParam {
	var modelUnitConfigV4 model.UnitConfigV4SQLParam
	modelUnitConfigV4.UnitConfigName = unitConfigName
	modelUnitConfigV4.MaxCPU = unitConfig.MaxCPU
	modelUnitConfigV4.MinCPU = unitConfig.MinCPU
	modelUnitConfigV4.MaxIops = unitConfig.MaxIops
	modelUnitConfigV4.MinIops = unitConfig.MinIops
	modelUnitConfigV4.IopsWeight = unitConfig.IopsWeight
	modelUnitConfigV4.MemorySize = unitConfig.MemorySize
	modelUnitConfigV4.LogDiskSize = unitConfig.LogDiskSize
	return &modelUnitConfigV4
}

func (m *OBTenantManager) generateSpecUnitNumMap(spec v1alpha1.OBTenantSpec) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range spec.Pools {
		unitNumMap[zone.Zone] = spec.UnitNumber
	}
	return unitNumMap
}

func (m *OBTenantManager) generateSpecLocalityMap(pools []v1alpha1.ResourcePoolSpec) map[string]*v1alpha1.LocalityType {
	localityMap := make(map[string]*v1alpha1.LocalityType, 0)
	for _, pool := range pools {
		localityMap[pool.Zone] = &v1alpha1.LocalityType{
			Name:     strings.ToUpper(pool.Type.Name), // locality type in DB is Upper
			Replica:  pool.Type.Replica,
			IsActive: pool.Type.IsActive,
		}
	}
	return localityMap
}

func (m *OBTenantManager) generateStatusLocalityMap(pools []v1alpha1.ResourcePoolStatus) map[string]*v1alpha1.LocalityType {
	localityMap := make(map[string]*v1alpha1.LocalityType, 0)
	for _, pool := range pools {
		localityMap[pool.ZoneList] = pool.Type
	}
	return localityMap
}

func (m *OBTenantManager) generateStatusUnitNumMap(zones []v1alpha1.ResourcePoolSpec) (map[string]int, error) {
	unitNumMap := make(map[string]int, 0)
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get sql error when get pool list")
	}
	for _, zone := range zones {
		poolName := m.generatePoolName(zone.Zone)
		for _, pool := range poolList {
			if pool.Name == poolName {
				unitNumMap[zone.Zone] = int(pool.UnitNum)
			}
		}
	}
	return unitNumMap, nil
}

func (m *OBTenantManager) generateLocalityList(localityMap map[string]*v1alpha1.LocalityType) []string {
	var locality []string
	var zoneSortList []string
	for k := range localityMap {
		zoneSortList = append(zoneSortList, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(zoneSortList)))
	for _, zoneList := range zoneSortList {
		zoneType := localityMap[zoneList]
		if zoneType.IsActive {
			locality = append(locality, fmt.Sprintf("%s{%d}@%s", zoneType.Name, zoneType.Replica, zoneList))
		}
	}
	return locality
}

func (m *OBTenantManager) generateSpecZoneList(pools []v1alpha1.ResourcePoolSpec) []string {
	var zoneList []string
	for _, pool := range pools {
		zoneList = append(zoneList, pool.Zone)
	}
	return zoneList
}

func (m *OBTenantManager) generateStatusZoneList(pools []v1alpha1.ResourcePoolStatus) []string {
	var zoneList []string
	for _, pool := range pools {
		zoneList = append(zoneList, pool.ZoneList)
	}
	return zoneList
}

func (m *OBTenantManager) generateSpecPoolList(pools []v1alpha1.ResourcePoolSpec) []string {
	var poolList []string
	for _, pool := range pools {
		poolName := m.generatePoolName(pool.Zone)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func (m *OBTenantManager) generateStatusPoolList(pools []v1alpha1.ResourcePoolStatus) []string {
	var poolList []string
	for _, pool := range pools {
		poolName := m.generatePoolName(pool.ZoneList)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func (m *OBTenantManager) generateSpecPrimaryZone(pools []v1alpha1.ResourcePoolSpec) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, pool := range pools {
		if pool.Type.IsActive {
			zones := zoneMap[pool.Priority]
			zones = append(zones, pool.Zone)
			zoneMap[pool.Priority] = zones
		}
	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		primaryZone = fmt.Sprint(primaryZone, strings.Join(zones, ","), ";")
	}
	return primaryZone
}

func (m *OBTenantManager) generateStatusPrimaryZone(pools []v1alpha1.ResourcePoolStatus) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, pool := range pools {
		if pool.Type.IsActive {
			zones := zoneMap[pool.Priority]
			zones = append(zones, pool.ZoneList)
			zoneMap[pool.Priority] = zones
		}
	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		primaryZone = fmt.Sprint(primaryZone, strings.Join(zones, ","), ";")
	}
	return primaryZone
}

func (m *OBTenantManager) generateLocality(zones []v1alpha1.ResourcePoolSpec) string {
	specLocalityMap := m.generateSpecLocalityMap(zones)
	localityList := m.generateLocalityList(specLocalityMap)
	return strings.Join(localityList, ",")
}

func (m *OBTenantManager) generateWhiteListInVariableForm(whiteList string) string {
	if whiteList == "" {
		return fmt.Sprintf("%s = '%s'", tenant.OBTcpInvitedNodes, tenant.DefaultOBTcpInvitedNodes)
	}
	return fmt.Sprintf("%s = '%s'", tenant.OBTcpInvitedNodes, whiteList)
}

func (m *OBTenantManager) generateStatusTypeMapFromLocalityStr(locality string) map[string]v1alpha1.LocalityType {
	typeMap := make(map[string]v1alpha1.LocalityType, 0)
	typeList := strings.Split(locality, ", ")
	for _, type1 := range typeList {
		split1 := strings.Split(type1, "@")
		typeName := strings.Split(split1[0], "{")[0]
		typeReplica := type1[strings.Index(type1, "{")+1 : strings.Index(type1, "}")]
		replicaInt, _ := strconv.Atoi(typeReplica)
		typeMap[split1[1]] = v1alpha1.LocalityType{
			Name:     typeName,
			Replica:  replicaInt,
			IsActive: true,
		}
	}
	return typeMap
}

func (m *OBTenantManager) generateStatusPriorityMap(primaryZone string) map[string]int {
	priorityMap := make(map[string]int, 0)
	cutZones := strings.Split(primaryZone, ";")
	priority := len(cutZones)
	for _, cutZone := range cutZones {
		zoneList := strings.Split(cutZone, ",")
		for _, zone := range zoneList {
			priorityMap[zone] = priority
		}
		priority--
	}
	return priorityMap
}

func (m *OBTenantManager) generateUnitName(zoneList string) string {
	tenantName := m.OBTenant.Spec.TenantName
	unitName := fmt.Sprintf("unitconfig_%s_%s", tenantName, zoneList)
	return unitName
}

func (m *OBTenantManager) generatePoolName(zoneList string) string {
	tenantName := m.OBTenant.Spec.TenantName
	poolName := fmt.Sprintf("pool_%s_%s", tenantName, zoneList)
	return poolName
}

// ---------- sql operator wrap ----------

func (m *OBTenantManager) getOBVersion() (string, error) {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Get OB Version")
	}
	version, err := oceanbaseOperationManager.GetVersion()
	if err != nil {
		return "", errors.Wrapf(err, "Tenant '%s' get ob version from db failed", tenantName)
	}
	return version.Version, nil
}

// sql wrap function

func (m *OBTenantManager) getCharset() (string, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Charset")
	}
	charset, err := oceanbaseOperationManager.GetCharset()
	if err != nil {
		return "", errors.Wrap(err, "Get sql error when get charset")
	}
	return charset.Charset, nil
}

func (m *OBTenantManager) getVariable(variableName string) (string, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Variable")
	}
	variable, err := oceanbaseOperationManager.GetVariable(variableName)
	if err != nil {
		return "", errors.Wrap(err, "Get sql error when get variable")
	}
	return variable.Value, nil
}

func (m *OBTenantManager) getTenantByName(tenantName string) (*model.OBTenant, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant")
	}
	tenant, err := oceanbaseOperationManager.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (m *OBTenantManager) getPoolByName(poolName string) (*model.Pool, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Pool by poolName")
	}
	pool, err := oceanbaseOperationManager.GetPoolByName(poolName)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (m *OBTenantManager) getUnitConfigV4ByName(unitName string) (*model.UnitConfigV4, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting UnitConfigV4 By unitConfig name")
	}
	unit, err := oceanbaseOperationManager.GetUnitConfigV4ByName(unitName)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (m *OBTenantManager) tenantExist(tenantName string) (bool, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether tenant exist")
	}
	isExist, err := oceanbaseOperationManager.CheckTenantExistByName(tenantName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (m *OBTenantManager) poolExist(poolName string) (bool, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether pool exist")
	}
	isExist, err := oceanbaseOperationManager.CheckPoolExistByName(poolName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (m *OBTenantManager) unitConfigV4Exist(unitConfigName string) (bool, error) {
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether UnitConfigV4 exist")
	}
	isExist, err := oceanbaseOperationManager.CheckUnitConfigExistByName(unitConfigName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (m *OBTenantManager) createPool(poolName, unitName string, pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Create Resource Pool", "tenantName", tenantName, "poolName", poolName)
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	poolSQLParam := model.PoolSQLParam{
		PoolName: poolName,
		UnitName: unitName,
		UnitNum:  int64(m.OBTenant.Spec.UnitNumber),
		ZoneList: pool.Zone,
	}
	return oceanbaseOperationManager.AddPool(poolSQLParam)
}

func (m *OBTenantManager) createUnitAndPoolV4(pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	unitName := m.generateUnitName(pool.Zone)
	poolName := m.generatePoolName(pool.Zone)

	err := m.createUnitConfigV4(unitName, pool.UnitConfig)
	if err != nil {
		m.Logger.Error(err, "Create UnitConfigV4 Failed", "tenantName", tenantName, "unitName", unitName)
		return err
	}
	err = m.createPool(poolName, unitName, pool)
	if err != nil {
		m.Logger.Error(err, "Create Tenant Failed", "tenantName", tenantName, "poolName", poolName)
		return err
	}
	return nil
}

func (m *OBTenantManager) deleteTenant() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Tenant ", tenantName))
	}

	tenantExist, err := m.tenantExist(tenantName)
	if err != nil {
		m.Logger.Error(err, "Check Whether The Tenant Exists Failed", "tenantName", tenantName)
		return err
	}
	if tenantExist {
		return oceanbaseOperationManager.DeleteTenant(tenantName, m.OBTenant.Spec.ForceDelete)
	}
	return nil
}

func (m *OBTenantManager) deletePool() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Pool", tenantName))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		poolName := m.generatePoolName(zone.Zone)
		poolExist, err := m.poolExist(poolName)
		if err != nil {
			m.Logger.Error(err, "Check Whether The Resource Pool Exists Failed")
			return err
		}
		if poolExist {
			err = oceanbaseOperationManager.DeletePool(poolName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) deleteUnitConfig() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getClusterSysClient()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Unit", tenantName))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		unitName := m.generateUnitName(zone.Zone)
		unitExist, err := m.unitConfigV4Exist(unitName)
		if err != nil {
			m.Logger.Error(err, "Check Whether The Resource Unit Exists Failed", "tenantName", tenantName, "unitName", unitName)
			return err
		}
		if unitExist {
			err = oceanbaseOperationManager.DeleteUnitConfig(unitName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) CreateUserWithCredentialSecrets() error {
	if m.OBTenant.Spec.TenantRole == constants.TenantRoleStandby {
		// standby tenant can not need to create user
		return nil
	}
	err := m.createUserWithCredentials()
	if err != nil {
		m.Telemetry.Event(m.OBTenant, corev1.EventTypeWarning, "Failed to create user or change password", err.Error())
		m.Logger.Error(err, "Failed to create user or change password, please check the credential secrets")
	}

	return nil
}

func (m *OBTenantManager) createUserWithCredentials() error {
	var con *operation.OceanbaseOperationManager
	var err error

	creds := m.OBTenant.Spec.Credentials
	if creds.Root != "" {
		con, err = m.getTenantClient()
		if err != nil {
			return err
		}
		rootPwd, err := ReadPassword(m.Client, m.OBTenant.Namespace, creds.Root)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				m.Logger.Error(err, "Failed to get root password secret")
				return err
			}
		} else if rootPwd != "" {
			err = con.ChangeTenantUserPassword(obconst.RootUser, rootPwd)
			if err != nil {
				m.Logger.Error(err, "Failed to change root password")
				return err
			}
		}
	}
	if creds.StandbyRO != "" {
		if con == nil {
			con, err = m.getTenantClient()
			if err != nil {
				return err
			}
		}
		standbyROPwd, err := ReadPassword(m.Client, m.OBTenant.Namespace, creds.StandbyRO)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				m.Logger.Error(err, "Failed to get standbyRO password secret")
				return err
			}
		} else {
			if standbyROPwd != "" {
				err = con.CreateUserWithPwd(obconst.StandbyROUser, standbyROPwd)
				if err != nil {
					m.Logger.Error(err, "Failed to create standbyRO user with password")
					return err
				}
			} else {
				err = con.CreateUser(obconst.StandbyROUser)
				if err != nil {
					m.Logger.Error(err, "Failed to create standbyRO user")
					return err
				}
			}
			err = con.GrantPrivilege(obconst.SelectPrivilege, obconst.OceanbaseAllScope, obconst.StandbyROUser)
			if err != nil {
				m.Logger.Error(err, "Failed to grant privilege to standbyRO")
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) CreateEmptyStandbyTenant() error {
	if m.OBTenant.Spec.Source == nil || m.OBTenant.Spec.Source.Tenant == nil {
		return errors.New("Empty standby tenant must have source tenant")
	}
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	restoreSource, err := getTenantRestoreSource(m.Ctx, m.Client, m.Logger, con, m.OBTenant.Namespace, *m.OBTenant.Spec.Source.Tenant)
	if err != nil {
		return err
	}
	poolList := m.generateSpecPoolList(m.OBTenant.Spec.Pools)
	primaryZone := m.generateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	locality := m.generateLocality(m.OBTenant.Spec.Pools)
	err = con.CreateEmptyStandbyTenant(&model.CreateEmptyStandbyTenantParam{
		TenantName:    m.OBTenant.Spec.TenantName,
		RestoreSource: restoreSource,
		PrimaryZone:   primaryZone,
		Locality:      locality,
		PoolList:      poolList,
	})
	if err != nil {
		return err
	}
	m.Telemetry.Event(m.OBTenant, "CreateEmptyStandby", "", "Succeed to create empty standby tenant")
	return nil
}

func (m *OBTenantManager) CheckPrimaryTenantLSIntegrity() error {
	var err error
	if m.OBTenant.Spec.Source == nil || m.OBTenant.Spec.Source.Tenant == nil {
		return errors.New("Primary tenant must have source tenant")
	}
	tenantCR := &v1alpha1.OBTenant{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBTenant.Namespace,
		Name:      *m.OBTenant.Spec.Source.Tenant,
	}, tenantCR)
	if err != nil {
		return err
	}

	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	lsDeletion, err := con.ListLSDeletion(int64(tenantCR.Status.TenantRecordInfo.TenantID))
	if err != nil {
		return err
	}
	if len(lsDeletion) > 0 {
		return errors.New("LS deletion set is not empty, log is of not integrity")
	}
	logStats, err := con.ListLogStats(int64(tenantCR.Status.TenantRecordInfo.TenantID))
	if err != nil {
		return err
	}
	if len(logStats) == 0 {
		return errors.New("Log stats is empty, out of expectation")
	}
	for _, ls := range logStats {
		if ls.BeginLSN != 0 {
			return errors.New("Log stats begin SCN is not 0, log is of not integrity")
		}
	}

	return nil
}
