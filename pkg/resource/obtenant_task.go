package resource

import (
	"fmt"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---------- task entry point ----------

func (m *OBTenantManager) CreateTenantTaskWithClear() error {
	err := m.CreateTenantTask()
	// clean created resource, restore to the initial state
	if err != nil {
		err := m.DeleteTenantTask()
		if err != nil {
			err = errors.Wrapf(err,"delete tenant when creating tenant")
			return err
		}
	}
	return err
}

func (m *OBTenantManager) CreateTenantTask() error {
	m.Logger.Info("===== debug: create tenant task =====")
	tenantName := m.OBTenant.Spec.TenantName


	m.Logger.Info("===== debug: Create Tenant begin =====")
	for _, pool := range m.OBTenant.Spec.Pools {
		err := m.CreateUnitAndPoolV4(pool)
		if err != nil {
			m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
			return err
		}
	}

	err := m.CreateTenant()
	if err != nil {
		m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
		return err
	}

	m.Logger.Info("Create Tenant success", "tenantName", tenantName)
	return nil
}

func (m *OBTenantManager) CheckTenantTask() error {
	tenantName := m.OBTenant.Spec.TenantName
	tenantExist, err := m.TenantExist(tenantName)
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
		unitName := m.GenerateUnitName(pool.ZoneList)
		poolName := m.GeneratePoolName(pool.ZoneList)
		poolExist, err := m.PoolExist(poolName)
		if err != nil{
			m.Logger.Error(err, "Check Resource Pool Exist", "tenantName", tenantName, "poolName", poolName)
			return err
		}
		if poolExist {
			return err
		}

		unitExist, err := m.UnitConfigV4Exist(unitName)
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
	m.Logger.Info("debug: add pool", "obtenant", m.OBTenant, "spec", m.OBTenant.Spec, "status", m.OBTenant.Status)

	// handle add pool
	poolSpecs := m.GetPoolsForAdd()
	m.Logger.Info(fmt.Sprintf("debug: poolSpecs %v", poolSpecs))
	for _, addPool := range poolSpecs {
		err := m.TenantAddPool(addPool)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) DeletePoolTask() error {
	m.Logger.Info("debug: delete pool", "obtenant", m.OBTenant, "spec", m.OBTenant.Spec, "status", m.OBTenant.Status)

	// handle delete pool
	poolStatuses := m.GetPoolsForDelete()
	m.Logger.Info(fmt.Sprintf("debug: poolStatuses %v", poolStatuses))
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

	version, err := m.GetOBVersion()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed, check and apply unitConfigV4", "tenantName", tenantName)
		return err
	}
	switch string(version[0]) {
	case tenant.Version4:
		return m.CheckAndApplyUnitConfigV4()
	}
	return errors.New("no match version for check and set unit config")
}

func (m *OBTenantManager) DeleteTenantTask() error {
	var err error
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Delete Tenant", "tenantName",tenantName)
	err = m.DeleteTenant()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Pool", "tenantName",tenantName)
	err = m.DeletePool()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Unit", "tenantName",tenantName)
	err = m.DeleteUnitConfig()
	if err != nil {
		return err
	}
	m.Logger.Info("Delete Tenant Success", "tenantName",tenantName)
	return nil
}

// ---------- Check And Apply function ----------

func (m *OBTenantManager) CheckAndApplyWhiteList() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Applying ob_tcp_invited_nodes For Tenant ",tenantName))
	}

	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.TenantRecordInfo.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if statusWhiteList != specWhiteList {
		m.Logger.Info("found specWhiteList didn't match", "tenantName", tenantName,
			"statusWhiteList",statusWhiteList, "specWhiteList", specWhiteList)
		variableList := m.GenerateWhiteListInVariableForm(specWhiteList)
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
	specUnitConfigMap := m.GenerateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := m.GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		match := true
		specUnitConfig := specUnitConfigMap[pool.ZoneList]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.ZoneList]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist{
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			m.Logger.Info("found unit config v4 didn't match", "tenantName", tenantName, "zoneName", pool.ZoneList,
				"statusUnitConfig", FormatUnitConfigV4(statusUnitConfigMap[pool.ZoneList]), "specUnitConfig",FormatUnitConfigV4(specUnitConfigMap[pool.ZoneList]))
			match = false
		}
		if !match {
			unitName := m.GenerateUnitName(pool.ZoneList)
			err := m.SetUnitConfigV4(unitName, specUnitConfigMap[pool.ZoneList])
			if err != nil {
				m.Logger.Error(err,"Set Tenant Unit failed","tenantName", tenantName, "unitName", unitName)
				return err
			}
		}
	}
	return nil
}



func (m *OBTenantManager) CheckAndApplyUnitNum() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
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
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Priority ", tenantName))
	}

	specPrimaryZone := m.GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	specPrimaryZoneMap := m.GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZone := m.GenerateStatusPrimaryZone(m.OBTenant.Status.Pools)
	statusPrimaryZoneMap := m.GeneratePrimaryZoneMap(statusPrimaryZone)
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
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Locality ", tenantName))
	}
	specLocalityMap := m.GenerateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := m.GenerateStatusLocalityMap(m.OBTenant.Status.Pools)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		locality := m.GenerateLocality(m.OBTenant.Spec.Pools)
		tenantSQLParam := model.TenantSQLParam{
			TenantName:  tenantName,
			Locality: locality,
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
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
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

func (m *OBTenantManager) CreateTenant() error {
	tenantName := m.OBTenant.Spec.TenantName
	pools := m.OBTenant.Spec.Pools
	m.Logger.Info("Create Tenant","tenantName", tenantName)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Tenant")
	}

	tenantSQLParam := model.TenantSQLParam{
		TenantName: tenantName,
		PrimaryZone: m.GenerateSpecPrimaryZone(pools),
		VariableList: m.GenerateWhiteListInVariableForm(m.OBTenant.Spec.ConnectWhiteList),
		Charset: m.OBTenant.Spec.Charset,
		PoolList: m.GenerateSpecPoolList(pools),
		Locality: m.GenerateLocality(pools),
		Collate: m.OBTenant.Spec.Collate,
	}
	if tenantSQLParam.Charset == "" {
		tenantSQLParam.Charset = tenant.Charset
	}

	err = oceanbaseOperationManager.AddTenant(tenantSQLParam)
	if err != nil {
		return err
	}
	GlobalWhiteListMap[tenantName] = m.OBTenant.Spec.ConnectWhiteList
	return nil
}

func (m *OBTenantManager) CreateUnitConfigV4(unitName string, unitConfig v1alpha1.UnitConfig) error {
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Create UnitConfig", "tenantName", tenantName, "unitName", unitName)
	unitModel := m.GenerateModelUnitConfigV4SQLParam(unitName, m.GenerateModelUnitConfigV4(unitConfig))
	if unitModel.MemorySize == 0 {
		err := errors.New("unit memorySize is empty")
		m.Logger.Error(err, "unit memorySize cannot be zero", "tenantName", tenantName, "unitName", unitName)
		return err
	}
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource UnitConfigV4")
	}

	return oceanbaseOperationManager.AddUnitConfigV4(unitModel)
}

func (m *OBTenantManager) SetUnitConfigV4(unitName string, unitConfig model.UnitConfigV4) error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	unitModel := m.GenerateModelUnitConfigV4SQLParam(unitName, unitConfig)
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", tenantName))
	}
	return oceanbaseOperationManager.SetUnitConfigV4(unitModel)
}

func (m *OBTenantManager) GetPoolsForAdd() []v1alpha1.ResourcePoolSpec {
	var pools []v1alpha1.ResourcePoolSpec
	for _, specZone := range m.OBTenant.Spec.Pools {
		exist := false
		for _, statusZone := range m.OBTenant.Status.Pools {
			if statusZone.ZoneList == specZone.ZoneList {
				exist = true
			}
		}
		if !exist {
			pools = append(pools, specZone)
		}
	}
	return pools
}

func (m *OBTenantManager) GetPoolsForDelete() []v1alpha1.ResourcePoolStatus {
	var poolStatuses []v1alpha1.ResourcePoolStatus
	for _, statusPool := range m.OBTenant.Status.Pools {
		exist := false
		for _, specPool := range m.OBTenant.Spec.Pools {
			if statusPool.ZoneList == specPool.ZoneList {
				exist = true
			}
		}
		if !exist {
			poolStatuses = append(poolStatuses, statusPool)
		}
	}
	return poolStatuses
}

func (m *OBTenantManager) TenantAddPool(poolAdd v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Add Pool", tenantName))
	}

	// step 1: create unit and poolAdd
	err = m.CreateUnitAndPoolV4(poolAdd)
	if err != nil {
		return err
	}

	// step 2.1: update locality and resource poolAdd list
	poolStatusAdd := v1alpha1.ResourcePoolStatus{
		ZoneList:   poolAdd.ZoneList,
		Type:       poolAdd.Type,
		UnitNumber: m.OBTenant.Spec.UnitNumber,
	}

	resourcePoolStatusList := append(m.OBTenant.Status.Pools, poolStatusAdd)
	statusLocalityMap := m.GenerateStatusLocalityMap(resourcePoolStatusList)
	localityList := m.GenerateLocalityList(statusLocalityMap)
	poolList := m.GenerateStatusPoolList(resourcePoolStatusList)
	specPrimaryZone := m.GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)

	tenantSQLParam := model.TenantSQLParam{
		TenantName: tenantName,
		Locality:   strings.Join(localityList, ","),
		PoolList: poolList,
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

	m.Logger.Info("Succeed add poolAdd", "deleted poolName", poolAdd.ZoneList)
	return nil
}

func (m *OBTenantManager) TenantDeletePool(poolDelete v1alpha1.ResourcePoolStatus) error {

	m.Logger.Info("===== debug: begin TenantDeletePool", "poolDelete", poolDelete)
	tenantName := m.OBTenant.Spec.TenantName
	poolName := m.GeneratePoolName(poolDelete.ZoneList)
	unitName := m.GenerateUnitName(poolDelete.ZoneList)

	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Delete Pool ", tenantName))
	}
	var zoneList []v1alpha1.ResourcePoolStatus
	for _, zone := range m.OBTenant.Status.Pools {
		if zone.ZoneList != poolDelete.ZoneList {
			zoneList = append(zoneList, zone)
		}
	}

	statusLocalityMap := m.GenerateStatusLocalityMap(zoneList)
	localityList := m.GenerateLocalityList(statusLocalityMap)
	poolList := m.GenerateStatusPoolList(zoneList)
	specPrimaryZone := m.GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)

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

	m.Logger.Info("debug: step 1.1: 'ALTER_TENANT' Job for deletePool successes", "tenantName", tenantName)

	// step 1.2: update resource pool list
	tenantSQLParam = model.TenantSQLParam{
		TenantName: tenantName,
		PoolList: poolList,
		PrimaryZone: specPrimaryZone,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		m.Logger.Error(err, "Modify Tenant, update poolList", "tenantName", tenantName)
		return err
	}
	m.Logger.Info("debug: step 1.2: update resource pool list successes", "tenantName", tenantName)

	// step 2: delete resource pool
	poolExist, err := m.PoolExist(poolName)
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

	m.Logger.Info("debug: step2: delete resource pool successes", "tenantName", tenantName)

	// step 3: delete unit
	unitExist, err := m.UnitConfigV4Exist(unitName)
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
	m.Logger.Info("debug: step3: delete resource unit config successes", "tenantName", tenantName)

	m.Logger.Info("Succeed delete pool", "deleted poolName", poolDelete.ZoneList)
	return nil
}

// ---------- compare helper function ----------

func IsUnitConfigV4Equal(specUnitConfig model.UnitConfigV4, statusUnitConfig model.UnitConfigV4) bool {
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

func FormatUnitConfigV4(unit model.UnitConfigV4) string {
	return fmt.Sprintf("MaxCPU: %f MinCPU:%f MemorySize:%d MaxIops:%d MinIops:%d IopsWeight:%d LogDiskSize:%d",
		unit.MaxCPU, unit.MinCPU, unit.MemorySize, unit.MaxIops, unit.MinIops, unit.IopsWeight, unit.LogDiskSize)
}

// ---------- generate "zoneName-value" map function ----------

func(m *OBTenantManager) GeneratePrimaryZoneMap(str string) map[int][]string {
	res := make(map[int][]string, 0)
	levelCuts := strings.Split(str, ";")
	for idx, levelCut := range levelCuts {
		cut := strings.Split(levelCut, ",")
		res[idx] = cut
		sort.Strings(res[idx])
	}
	return res
}

func(m *OBTenantManager) GenerateSpecUnitConfigV4Map(spec v1alpha1.OBTenantSpec) map[string]model.UnitConfigV4 {
	var unitConfigMap = make(map[string]model.UnitConfigV4, 0)
	for _, pool := range spec.Pools {
		unitConfigMap[pool.ZoneList] = m.GenerateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func (m *OBTenantManager) GenerateStatusUnitConfigV4Map(status v1alpha1.OBTenantStatus) map[string]model.UnitConfigV4 {
	var unitConfigMap = make(map[string]model.UnitConfigV4, 0)
	for _, pool := range status.Pools {
		unitConfigMap[pool.ZoneList] = m.GenerateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func (m *OBTenantManager) GenerateModelUnitConfigV4(unitConfig v1alpha1.UnitConfig) model.UnitConfigV4 {
	var modelUnitConfigV4 model.UnitConfigV4
	modelUnitConfigV4.MaxCPU = unitConfig.MaxCPU.AsApproximateFloat64()
	modelUnitConfigV4.MinCPU = unitConfig.MinCPU.AsApproximateFloat64()
	modelUnitConfigV4.MaxIops = int64(unitConfig.MaxIops)
	modelUnitConfigV4.MinIops = int64(unitConfig.MinIops)
	modelUnitConfigV4.IopsWeight = int64(unitConfig.IopsWeight)
	modelUnitConfigV4.MemorySize = unitConfig.MemorySize.Value()
	modelUnitConfigV4.LogDiskSize = unitConfig.LogDiskSize.Value()
	return modelUnitConfigV4
}

func (m *OBTenantManager) GenerateModelUnitConfigV4SQLParam(unitConfigName string, unitConfig model.UnitConfigV4) model.UnitConfigV4SQLParam{
	var modelUnitConfigV4 model.UnitConfigV4SQLParam
	modelUnitConfigV4.UnitConfigName = unitConfigName
	modelUnitConfigV4.MaxCPU = unitConfig.MaxCPU
	modelUnitConfigV4.MinCPU = unitConfig.MinCPU
	modelUnitConfigV4.MaxIops = unitConfig.MaxIops
	modelUnitConfigV4.MinIops = unitConfig.MinIops
	modelUnitConfigV4.IopsWeight = unitConfig.IopsWeight
	modelUnitConfigV4.MemorySize = unitConfig.MemorySize
	modelUnitConfigV4.LogDiskSize = unitConfig.LogDiskSize
	return modelUnitConfigV4
}

func(m *OBTenantManager) GenerateSpecUnitNumMap(spec v1alpha1.OBTenantSpec) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range spec.Pools {
		unitNumMap[zone.ZoneList] = spec.UnitNumber
	}
	return unitNumMap
}

func(m *OBTenantManager) GenerateSpecLocalityMap(pools []v1alpha1.ResourcePoolSpec) map[string]v1alpha1.LocalityType {
	localityMap := make(map[string]v1alpha1.LocalityType, 0)
	for _, pool := range pools {
		localityMap[pool.ZoneList] = v1alpha1.LocalityType{
			Name: strings.ToUpper(pool.Type.Name), // locality type in DB is Upper
			Replica: pool.Type.Replica,
			IsActive: pool.Type.IsActive,
		}
	}
	return localityMap
}

func (m *OBTenantManager) GenerateStatusLocalityMap(pools []v1alpha1.ResourcePoolStatus) map[string]v1alpha1.LocalityType {
	localityMap := make(map[string]v1alpha1.LocalityType, 0)
	for _, pool := range pools {
		localityMap[pool.ZoneList] = pool.Type
	}
	return localityMap
}

func (m *OBTenantManager) GenerateLocalityList(localityMap map[string]v1alpha1.LocalityType) []string {
	var locality []string
	var zoneSortList []string
	for k := range localityMap{
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

func (m *OBTenantManager) GenerateSpecZoneList(pools []v1alpha1.ResourcePoolSpec) []string {
	var zoneList []string
	for _, pool := range pools {
		zoneList = append(zoneList, pool.ZoneList)
	}
	return zoneList
}

func (m *OBTenantManager) GenerateStatusZoneList(pools []v1alpha1.ResourcePoolStatus) []string {
	var zoneList []string
	for _, pool := range pools {
		zoneList = append(zoneList, pool.ZoneList)
	}
	return zoneList
}

func (m *OBTenantManager) GenerateSpecPoolList(pools []v1alpha1.ResourcePoolSpec) []string {
	var poolList []string
	for _, pool := range pools {
		poolName := m.GeneratePoolName(pool.ZoneList)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func (m *OBTenantManager) GenerateStatusPoolList(pools []v1alpha1.ResourcePoolStatus) []string {
	var poolList []string
	for _, pool := range pools {
		poolName := m.GeneratePoolName(pool.ZoneList)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func(m *OBTenantManager) GenerateSpecPrimaryZone(pools []v1alpha1.ResourcePoolSpec) string {
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
		primaryZone = fmt.Sprint(primaryZone ,strings.Join(zones, ","), ";")
	}
	return primaryZone
}

func(m *OBTenantManager) GenerateStatusPrimaryZone(pools []v1alpha1.ResourcePoolStatus) string {
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

func (m *OBTenantManager) GenerateLocality(zones []v1alpha1.ResourcePoolSpec) string {
	specLocalityMap := m.GenerateSpecLocalityMap(zones)
	localityList := m.GenerateLocalityList(specLocalityMap)
	return strings.Join(localityList, ",")
}

func (m *OBTenantManager) GenerateWhiteListInVariableForm(whiteList string) string {
	if whiteList == "" {
		return fmt.Sprintf("%s = '%s'", tenant.OBTcpInvitedNodes, tenant.DefaultOBTcpInvitedNodes)
	} else {
		return fmt.Sprintf("%s = '%s'", tenant.OBTcpInvitedNodes, whiteList)
	}
}

func (m *OBTenantManager) GenerateStatusTypeMapFromLocalityStr(locality string) map[string]v1alpha1.LocalityType {
	typeMap := make(map[string]v1alpha1.LocalityType, 0)
	typeList := strings.Split(locality, ", ")
	for _, type1 := range typeList {
		split1 := strings.Split(type1, "@")
		typeName := strings.Split(split1[0], "{")[0]
		typeReplica := type1[strings.Index(type1, "{")+1 : strings.Index(type1, "}")]
		replicaInt, _ := strconv.Atoi(typeReplica)
		typeMap[split1[1]] = v1alpha1.LocalityType{
			Name:    typeName,
			Replica: replicaInt,
			IsActive: true,
		}
	}
	return typeMap
}

func (m *OBTenantManager) GenerateStatusPriorityMap(primaryZone string) map[string]int {
	priorityMap := make(map[string]int, 0)
	cutZones := strings.Split(primaryZone, ";")
	priority := len(cutZones)
	for _, cutZone := range cutZones {
		zoneList := strings.Split(cutZone, ",")
		for _, zone := range zoneList {
			priorityMap[zone] = priority
		}
		priority -= 1
	}
	return priorityMap
}


// ---------- sql operator wrap ----------

func (m *OBTenantManager) GetOBVersion() (string, error) {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Get OB Version")
	}
	version, err := oceanbaseOperationManager.GetVersion()
	if err != nil {
		return "", errors.Wrapf(err,"Tenant '%s' get ob version from db failed", tenantName)
	}
	m.Logger.Info("debug version", "version", version)
	return version.Version, nil
}

func (m *OBTenantManager) GenerateUnitName(zoneList string) string {
	tenantName := m.OBTenant.Spec.TenantName
	unitName := fmt.Sprintf("unitconfig_%s_%s", tenantName, zoneList)
	return unitName
}

func (m *OBTenantManager) GeneratePoolName( zoneList string) string {
	tenantName := m.OBTenant.Spec.TenantName
	poolName := fmt.Sprintf("pool_%s_%s", tenantName, zoneList)
	return poolName
}

// sql wrap function

func (m *OBTenantManager) GetCharset() (string, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Charset")
	}
	charset, err := oceanbaseOperationManager.GetCharset()
	if err != nil {
		return "", errors.Wrap(err, "Get sql error when get charset")
	}
	return charset.Charset, nil
}

func (m *OBTenantManager) GetVariable(variableName string) (string, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Variable")
	}
	variable, err := oceanbaseOperationManager.GetVariable(variableName)
	if err != nil {
		return "", errors.Wrap(err, "Get sql error when get variable")
	}
	return variable.Value, nil
}

func (m *OBTenantManager) GetTenantByName(tenantName string) (*model.Tenant, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant")
	}
	tenant, err := oceanbaseOperationManager.GetTenantByName(tenantName)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (m *OBTenantManager) GetPoolByName(poolName string) (*model.Pool, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Pool by poolName")
	}
	pool, err := oceanbaseOperationManager.GetPoolByName(poolName)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (m *OBTenantManager) GetUnitConfigV4ByName(unitName string) (*model.UnitConfigV4, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting UnitConfigV4 By unitConfig name")
	}
	unit, err := oceanbaseOperationManager.GetUnitConfigV4ByName(unitName)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (m *OBTenantManager) TenantExist(tenantName string) (bool, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether tenant exist")
	}
	isExist, err := oceanbaseOperationManager.CheckTenantExistByName(tenantName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (m *OBTenantManager) PoolExist(poolName string) (bool, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether pool exist")
	}
	isExist, err := oceanbaseOperationManager.CheckPoolExistByName(poolName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func (m *OBTenantManager) UnitConfigV4Exist(unitConfigName string) (bool, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return false, errors.Wrap(err, "Get Sql Operator Error When Check whether UnitConfigV4 exist")
	}
	isExist, err := oceanbaseOperationManager.CheckUnitConfigExistByName(unitConfigName)
	if err != nil {
		return false, err
	}
	return isExist, nil
}


func (m *OBTenantManager) CreatePool(poolName, unitName string, pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Create Resource Pool", "tenantName", tenantName, "poolName", poolName)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	poolSQLParam := model.PoolSQLParam{
		PoolName: poolName,
		UnitName: unitName,
		UnitNum:  int64(m.OBTenant.Spec.UnitNumber),
		ZoneList: pool.ZoneList,
	}
	return oceanbaseOperationManager.AddPool(poolSQLParam)
}

func (m *OBTenantManager) CreateUnitAndPoolV4(pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	unitName := m.GenerateUnitName(pool.ZoneList)
	poolName := m.GeneratePoolName(pool.ZoneList)

	err := m.CreateUnitConfigV4(unitName, pool.UnitConfig)
	if err != nil {
		m.Logger.Error(err, "Create UnitConfigV4 Failed", "tenantName", tenantName, "unitName", unitName)
		return err
	}
	err = m.CreatePool(poolName, unitName, pool)
	if err != nil {
		m.Logger.Error(err,"Create Tenant Failed", "tenantName", tenantName, "poolName", poolName)
		return err
	}
	return nil
}

func (m *OBTenantManager) DeleteTenant() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Tenant ", tenantName))
	}

	tenantExist, err := m.TenantExist(tenantName)
	if err != nil {
		m.Logger.Error(err, "Check Whether The Tenant Exists Failed", "tenantName", tenantName)
		return err
	}
	if tenantExist {
		return oceanbaseOperationManager.DeleteTenant(tenantName, m.OBTenant.Spec.ForceDelete)
	}
	return nil
}

func (m *OBTenantManager) DeletePool() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Pool", tenantName))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		poolName := m.GeneratePoolName(zone.ZoneList)
		poolExist,  err := m.PoolExist(poolName)
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

func (m *OBTenantManager) DeleteUnitConfig() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Unit", tenantName))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		unitName := m.GenerateUnitName(zone.ZoneList)
		unitExist, err := m.UnitConfigV4Exist(unitName)
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
