package resource

import (
	"fmt"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/const/status/obtenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
	kuberesource "k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ---------- task entry point ----------

func (m *OBTenantManager) CreateTenantTask() error {
	m.Logger.Info("===== debug: create tenant task =====")
	tenantName := m.OBTenant.Spec.TenantName
	tenantExist, err := m.TenantExist(tenantName)
	if err != nil {
		m.Logger.Error(err, "Check Whether Tenant exist failed", "tenantName", tenantName)
		return err
	}
	if tenantExist {
		m.OBTenant.Status.Status = obtenant.Pending
		err = errors.New("tenant has exist")
		m.Logger.Error(err, "tenant has exist", "tenantName", tenantName)
		return err
	}

	m.Logger.Info("===== debug: Create Tenant begin =====")
	for _, pool := range m.OBTenant.Spec.Pools{
		err := m.CreateUnitAndPoolV4(pool)
		if err != nil {
			m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
			return err
		}
	}

	err = m.CreateTenant()
	if err != nil {
		m.Logger.Error(err, "Create Tenant failed", "tenantName", tenantName)
		return err
	}

	m.Logger.Info("Create Tenant success", "tenantName", tenantName)
	return nil
}

func (m *OBTenantManager) MaintainTenantTask() error {
	tenantName := m.OBTenant.Spec.TenantName

	err := m.CheckAndApplyTcpInvitedNode()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed ----- check and set tcp invited node failed", "tenantName", tenantName)
		return err
	}
	err = m.CheckAndApplyUnitConfig()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed ----- check and apply unitConfigV4 failed", "tenantName", tenantName)
		return err
	}
	err = m.CheckAndApplyResourcePool()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed ----- Check And Apply Resource Pool failed","tenantName", tenantName)
		return err
	}
	err = m.CheckAndApplyTenant()
	if err != nil {
		m.Logger.Error(err, "maintain tenant failed ----- check and apply tenant failed", "tenantName", tenantName)
		return err
	}
	m.Logger.Info("Maintain tenant success", "tenantName", tenantName)
	return nil
}

func (m *OBTenantManager) hasModifiedTenantTask() bool {
	tenantName := m.OBTenant.Spec.TenantName

	hasModifiedTcpInvitedNode := m.hasModifiedTcpInvitedNode()
	if hasModifiedTcpInvitedNode {
		m.Logger.Info("Maintain Tenant ----- Tcp Invited Node has modified", "tenantName", tenantName)
	}
	hasModifiedUnitConfigV4 := m.hasModifiedUnitConfigV4()
	if hasModifiedUnitConfigV4 {
		m.Logger.Info("Maintain Tenant ----- UnitConfigV4 has modified", "tenantName", tenantName)
	}
	hasModifiedResourcePool := m.hasModifiedResourcePool()
	if hasModifiedResourcePool {
		m.Logger.Info("Maintain Tenant ----- Resource Pool has modified", "tenantName", tenantName)
	}
	hasModifiedTenant := m.hasModifiedTenant()
	if hasModifiedTenant {
		m.Logger.Info("Maintain Tenant ----- Tenant has modified", "tenantName", tenantName)
	}
	return hasModifiedTcpInvitedNode || hasModifiedUnitConfigV4 || hasModifiedResourcePool || hasModifiedTenant
}

func (m *OBTenantManager) DeleteTenantTask() error {
	var err error
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Begin Delete Tenant", "tenantName",tenantName)
	err = m.DeleteTenant()
	if err != nil {
		return err
	}
	m.Logger.Info("Begin Delete Pool", "tenantName",tenantName)
	err = m.DeletePool()
	if err != nil {
		return err
	}
	m.Logger.Info("Begin Delete Unit", "tenantName",tenantName)
	err = m.DeleteUnitConfig()
	if err != nil {
		return err
	}
	m.Logger.Info("Succeed Delete Tenant", "tenantName",tenantName)
	return nil
}

// ---------- Check And Apply function ----------

func (m *OBTenantManager) CheckAndApplyTcpInvitedNode() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Applying ob_tcp_invited_nodes For Tenant ",tenantName))
	}

	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.ConnectWhiteList

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
		m.OBTenant.Status.ConnectWhiteList = specWhiteList
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyUnitConfig() error {
	version, err := m.GetOBVersion()
	if err != nil {
		return err
	}
	switch string(version[0]) {
	case tenant.Version4:
		return m.CheckAndApplyUnitConfigV4()
	}
	return errors.New("no match version for check and set unit config")
}

func (m *OBTenantManager) CheckAndApplyUnitConfigV4() error {
	tenantName := m.OBTenant.Spec.TenantName
	specUnitConfigMap := GenerateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
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

func (m *OBTenantManager) CheckAndApplyResourcePool() error {
	m.Logger.Info("===== debug: begin CheckAndApplyResourcePool")
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Applying Resource Pool ", tenantName))
	}

	// handle add pool
	for _, addPool := range m.GetPoolsForAdd(){
		if addPool.ZoneList != "" {
			err := m.TenantAddPool(addPool)
			if err != nil {
				return err
			}
		}
	}

	// handle delete pool
	poolStatuses := m.GetPoolsForDelete()
	if len(poolStatuses) > 0 {
		if err != nil {
			return err
		}
		for _, poolStatus := range poolStatuses {
			err = m.TenantDeletePool(poolStatus)
			if err != nil {
				return err
			}
		}
	}

	// handle pool unitNum changed
	specUnitNumMap := GenerateSpecUnitNumMap(m.OBTenant.Spec)
	statusUnitNumMap := GenerateStatusUnitNumMap(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		specUnitNum := specUnitNumMap[pool.ZoneList]
		statusUnitNum := statusUnitNumMap[pool.ZoneList]
		if specUnitNum != statusUnitNum {
			m.Logger.Info("found unit_num didn't match","zoneName", pool.ZoneList,"statusUnitNum", statusUnitNum, "specUnitNum", specUnitNum)
			poolName := m.GeneratePoolName(pool.ZoneList)
			m.Logger.Info("set pool unit_num","zoneName", pool.ZoneList, "specUnitNum", specUnitNum)
			err = oceanbaseOperationManager.SetPoolUnitNum(poolName, specUnitNum)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyTenant() error {
	var err error
	err = m.CheckAndApplyPriority()
	if err != nil {
		return err
	}
	err = m.CheckAndApplyLocality()
	if err != nil {
		return err
	}
	err = m.CheckAndApplyCharset()
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyPriority() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()

	specPrimaryZone := GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	statusPrimaryZone := GenerateStatusPrimaryZone(m.OBTenant.Status.Pools)
	specPrimaryZoneMap := GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := GeneratePrimaryZoneMap(statusPrimaryZone)
	if reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Priority ", tenantName))
	}
	tenantSQLParam := model.TenantSQLParam{
		TenantName:  tenantName,
		PrimaryZone: specPrimaryZone,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyLocality() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	specLocalityMap := GenerateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := GenerateStatusLocalityMap(m.OBTenant.Status.Pools)
	if reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		return nil
	}
	locality := m.GenerateLocality(m.OBTenant.Spec.Pools)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Locality ", tenantName))
	}
	tenantSQLParam := model.TenantSQLParam{
		TenantName:  tenantName,
		Locality: locality,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyCharset() error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Checking and Applying Tenant '%s' Charset ", tenantName))
	}
	charset := m.OBTenant.Spec.Charset
	if charset != "" {
		charset = fmt.Sprintf("CHARSET = %s", charset)
		tenantSQLParam := model.TenantSQLParam{
			TenantName:  tenantName,
			Charset: charset,
		}
		err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
		if err != nil {
			return err
		}
	}
	return nil
}

// ---------- Check function ----------

func (m *OBTenantManager) hasModifiedTcpInvitedNode() bool {
	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if statusWhiteList != specWhiteList {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedUnitConfigV4() bool {
	specUnitConfigMap := GenerateSpecUnitConfigV4Map(m.OBTenant.Spec)
	statusUnitConfigMap := GenerateStatusUnitConfigV4Map(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		specUnitConfig := specUnitConfigMap[pool.ZoneList]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.ZoneList]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist{
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			return false
		}
	}
	return false
}
func (m *OBTenantManager) hasModifiedResourcePool() bool{
	// check add pool
	 if len(m.GetPoolsForAdd()) > 0 {
		return true
	}

	// check delete pool
	if len(m.GetPoolsForDelete()) > 0 {
		return false
	}

	// handle pool unitNum changed
	specUnitNumMap := GenerateSpecUnitNumMap(m.OBTenant.Spec)
	statusUnitNumMap := GenerateStatusUnitNumMap(m.OBTenant.Status)
	for _, pool := range m.OBTenant.Spec.Pools {
		specUnitNum := specUnitNumMap[pool.ZoneList]
		statusUnitNum := statusUnitNumMap[pool.ZoneList]
		if specUnitNum != statusUnitNum {
			return true
		}
	}
	return false
}


func (m *OBTenantManager) hasModifiedTenant() bool {
	return m.hasModifiedPriority() || m.hasModifiedLocality() || m.hasModifiedCharset()
}

func (m *OBTenantManager) hasModifiedPriority() bool {
	specPrimaryZone := GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	statusPrimaryZone := GenerateStatusPrimaryZone(m.OBTenant.Status.Pools)
	specPrimaryZoneMap := GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := GeneratePrimaryZoneMap(statusPrimaryZone)
	if reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedLocality() bool {
	specLocalityMap := GenerateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := GenerateStatusLocalityMap(m.OBTenant.Status.Pools)
	if reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		return true
	}
	return false
}

func (m *OBTenantManager) hasModifiedCharset() bool {
	specCharset := m.OBTenant.Spec.Charset
	statusCharset := m.OBTenant.Status.Charset
	if specCharset == statusCharset{
		return true
	}
	return false
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
		PrimaryZone: GenerateSpecPrimaryZone(pools),
		VariableList: m.GenerateWhiteListInVariableForm(m.OBTenant.Spec.ConnectWhiteList),
		Charset: m.OBTenant.Spec.Charset,
		PoolList: m.GenerateSpecPoolList(pools),
		Locality: m.GenerateLocality(pools),
		Collate: m.OBTenant.Spec.Collate,
	}
	if tenantSQLParam.Charset == "" {
		tenantSQLParam.Charset = tenant.Charset
	}

	return oceanbaseOperationManager.AddTenant(tenantSQLParam)
}

func (m *OBTenantManager) CreateUnitConfigV4(unitName string, unitConfig v1alpha1.UnitConfig) error {
	tenantName := m.OBTenant.Spec.TenantName
	m.Logger.Info("Create UnitConfig", "tenantName", tenantName, "unitName", unitName)
	unitModel := GenerateModelUnitConfigV4SQLParam(unitName, GenerateModelUnitConfigV4(unitConfig))
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
	unitModel := GenerateModelUnitConfigV4SQLParam(unitName, unitConfig)
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

func (m *OBTenantManager) TenantAddPool(pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Add Pool", tenantName))
	}
	err = m.CreateUnitAndPoolV4(pool)
	if err != nil {
		return err
	}

	poolStatusAdd := v1alpha1.ResourcePoolStatus{
		ZoneList:   pool.ZoneList,
		Type:       pool.Type,
		UnitNumber: pool.UnitNumber,
	}
	resourcePoolStatusList := append(m.OBTenant.Status.Pools, poolStatusAdd)
	poolList := m.GenerateStatusPoolList(resourcePoolStatusList)
	statusLocalityMap := GenerateStatusLocalityMap(resourcePoolStatusList)
	localityList := m.GenerateLocalityList(statusLocalityMap)

	tenantSQLParam := model.TenantSQLParam{
		TenantName: tenantName,
		PoolList:   poolList,
		Locality:   strings.Join(localityList, ","),
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)

	if err != nil {
		return err
	}
	m.Logger.Info("Wait For Tenant 'ALTER_TENANT' Job for addPool Finished", "tenantName", tenantName)
	for {
		_, err := oceanbaseOperationManager.GetRsJob(tenantName)
		if err != nil {
			break
		}
	}
	return nil
}

func (m *OBTenantManager) TenantDeletePool(deleteZone v1alpha1.ResourcePoolStatus) error {

	m.Logger.Info("===== debug: begin TenantDeletePool", "deleteZone", deleteZone)
	tenantName := m.OBTenant.Spec.TenantName
	poolName := m.GeneratePoolName(deleteZone.ZoneList)
	unitName := m.GenerateUnitName(deleteZone.ZoneList)

	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' -- Delete Pool ", tenantName))
	}
	var zoneList []v1alpha1.ResourcePoolStatus
	for _, zone := range m.OBTenant.Status.Pools {
		if zone.ZoneList != deleteZone.ZoneList {
			zoneList = append(zoneList, zone)
		}
	}

	statusLocalityMap := GenerateStatusLocalityMap(zoneList)
	localityList := m.GenerateLocalityList(statusLocalityMap)
	poolList := m.GenerateStatusPoolList(zoneList)
	tenantSQLParam := model.TenantSQLParam{
		TenantName: tenantName,
		Locality:   strings.Join(localityList, ","),
		PoolList: poolList,
	}
	err = oceanbaseOperationManager.SetTenant(tenantSQLParam)
	if err != nil {
		m.Logger.Error(err, "Modify Tenant failed", "tenantName", tenantName)
		return err
	}
	m.Logger.Info("Wait For Tenant 'ALTER_TENANT' Job for deletePool Finished", "tenantName", tenantName)
	for {
		_, err := oceanbaseOperationManager.GetRsJob(tenantName)
		if err != nil {
			break
		}
	}

	// delete pool
	poolExist, err := m.PoolExist(poolName)
	if err != nil {
		m.Logger.Error(err, "Check ResourcePool exist failed", "poolName", poolName)
		return err
	}
	if poolExist {
		err = oceanbaseOperationManager.DeletePool(poolName)
		if err != nil {
			return err
		}
	}

	// delete unit
	unitExist, err := m.UnitConfigV4Exist(unitName)
	if err != nil {
		m.Logger.Error(err, "Check UnitConfigV4 Exist failed", "unitName", unitName)
		return err
	}
	if unitExist {
		err = oceanbaseOperationManager.DeleteUnitConfig(unitName)
		if err != nil {
			return err
		}
	}
	m.Logger.Info("Succeed delete pool", "deleted poolName", deleteZone.ZoneList)
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

func GeneratePrimaryZoneMap(str string) map[int][]string {
	res := make(map[int][]string, 0)
	levelCuts := strings.Split(str, ";")
	for idx, levelCut := range levelCuts {
		cut := strings.Split(levelCut, ",")
		res[idx] = cut
		sort.Strings(res[idx])
	}
	return res
}

func GenerateSpecUnitConfigV4Map(spec v1alpha1.OBTenantSpec) map[string]model.UnitConfigV4 {
	var unitConfigMap = make(map[string]model.UnitConfigV4, 0)
	for _, pool := range spec.Pools {
		unitConfigMap[pool.ZoneList] = GenerateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func GenerateStatusUnitConfigV4Map(status v1alpha1.OBTenantStatus) map[string]model.UnitConfigV4 {
	var unitConfigMap = make(map[string]model.UnitConfigV4, 0)
	for _, pool := range status.Pools {
		unitConfigMap[pool.ZoneList] = GenerateModelUnitConfigV4(pool.UnitConfig)
	}
	return unitConfigMap
}

func GenerateModelUnitConfigV4(unitConfig v1alpha1.UnitConfig) model.UnitConfigV4 {
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

func GenerateModelUnitConfigV4SQLParam(unitConfigName string, unitConfig model.UnitConfigV4) model.UnitConfigV4SQLParam{
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

func GenerateSpecUnitNumMap(spec v1alpha1.OBTenantSpec) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range spec.Pools {
		unitNumMap[zone.ZoneList] = zone.UnitNumber
	}
	return unitNumMap
}

func GenerateStatusUnitNumMap(status v1alpha1.OBTenantStatus) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range status.Pools {
		unitNumMap[zone.ZoneList] = zone.UnitNumber
	}
	return unitNumMap
}

func GenerateSpecLocalityMap(zones []v1alpha1.ResourcePoolSpec) map[string]v1alpha1.LocalityType {
	localityMap := make(map[string]v1alpha1.LocalityType, 0)
	for _, zone := range zones {
		if zone.Type.Name != "" {
			switch strings.ToUpper(zone.Type.Name) {
			case tenant.TypeFull:
				localityMap[zone.ZoneList] = v1alpha1.LocalityType{
					Name:    tenant.TypeFull,
					Replica: 1,
				}
			case tenant.TypeLogonly:
				localityMap[zone.ZoneList] = v1alpha1.LocalityType{
					Name:    tenant.TypeLogonly,
					Replica: 1,
				}
			case tenant.TypeReadonly:
				var replica int
				if zone.Type.Replica == 0 {
					replica = 1
				} else {
					replica = zone.Type.Replica
				}
				localityMap[zone.ZoneList] = v1alpha1.LocalityType{
					Name:    tenant.TypeReadonly,
					Replica: replica,
				}
			}
		}
	}
	return localityMap
}

func GenerateStatusLocalityMap(pools []v1alpha1.ResourcePoolStatus) map[string]v1alpha1.LocalityType {
	localityMap := make(map[string]v1alpha1.LocalityType, 0)
	for _, pool := range pools {
		localityMap[pool.ZoneList] = pool.Type
	}
	return localityMap
}

func (m *OBTenantManager) GenerateLocalityList(localityMap map[string]v1alpha1.LocalityType) []string {
	var locality []string
	for zoneList, zoneType := range localityMap {
		if zoneType.Name != "" {
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

func GenerateSpecPrimaryZone(pools []v1alpha1.ResourcePoolSpec) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, pool := range pools {
		zones := zoneMap[pool.Priority]
		zones = append(zones, pool.ZoneList)
		zoneMap[pool.Priority] = zones
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

func GenerateStatusPrimaryZone(zones []v1alpha1.ResourcePoolStatus) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, zone := range zones {
		zones := zoneMap[zone.Priority]
		zones = append(zones, zone.ZoneList)
		zoneMap[zone.Priority] = zones
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
	specLocalityMap := GenerateSpecLocalityMap(zones)
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

func GenerateTypeMap(locality string) map[string]v1alpha1.LocalityType {
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
		}
	}
	return typeMap
}

func GeneratePriorityMap(primaryZone string) map[string]int {
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

// ---------- buildTenant function ----------

func (m *OBTenantManager) BuildTenantStatus() (*v1alpha1.OBTenantStatus ,error) {
	tenantName := m.OBTenant.Spec.TenantName
	tenantCurrentStatus := &v1alpha1.OBTenantStatus{}

	tenantExist, err := m.TenantExist(tenantName)
	if err != nil {
		return nil, err
	}
	if !tenantExist {
		return nil, errors.New(fmt.Sprintf("Tenant not exist, Tenant name: %s", tenantName))
	}
	gvTenant, err := m.GetTenantByName(tenantName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprint("Cannot Get Tenant Failed When Build Tenant Status", tenantName))
	}

	poolStatusList, err := m.BuildPoolStatusList(gvTenant)
	if err != nil {
		return nil, err
	}
	tenantCurrentStatus.Status = m.OBTenant.Status.Status
	tenantCurrentStatus.Pools = poolStatusList
	tenantCurrentStatus.OperationContext = m.OBTenant.Status.OperationContext
	tenantCurrentStatus.ConnectWhiteList, err = m.GetVariable(tenant.OBTcpInvitedNodes)
	if err != nil {
		return nil, err
	}

	tenantCurrentStatus.Charset, err = m.GetCharset()
	if err != nil {
		return nil, err
	}
	return tenantCurrentStatus, nil
}

func (m *OBTenantManager) BuildPoolStatusList(gvTenant *model.Tenant) ([]v1alpha1.ResourcePoolStatus, error) {
	var poolStatusList []v1alpha1.ResourcePoolStatus
	var err error
	var locality string
	var primaryZone string

	locality = gvTenant.Locality
	primaryZone = gvTenant.PrimaryZone
	typeMap := GenerateTypeMap(locality)
	tenantID := gvTenant.TenantID
	priorityMap := GeneratePriorityMap(primaryZone)
	unitNumMap, err := m.GenerateStatusUnitNumMap(m.OBTenant.Spec.Pools)
	if err != nil {
		return poolStatusList, err
	}
	zoneList, err := m.GenerateStatusZone(tenantID)
	if err != nil {
		return poolStatusList, err
	}
	for _, zone := range zoneList {
		var tenantCurrentStatus v1alpha1.ResourcePoolStatus
		tenantCurrentStatus.ZoneList = zone
		tenantCurrentStatus.Type = typeMap[zone]
		tenantCurrentStatus.UnitNumber = unitNumMap[zone]
		tenantCurrentStatus.Priority = priorityMap[zone]
		tenantCurrentStatus.UnitConfig, err = m.BuildUnitConfigV4FromDB(zone, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		tenantCurrentStatus.Units, err = m.BuildUnitStatusFromDB(zone, tenantID)
		if err != nil {
			return poolStatusList, err
		}
		poolStatusList = append(poolStatusList, tenantCurrentStatus)
	}
	return poolStatusList, nil
}

func (m *OBTenantManager) GenerateStatusZone(tenantID int64) ([]string, error) {
	var zoneList []string
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return zoneList, errors.Wrap(err, "Get Sql Operator Error When Generating Zone For Tenant CR Status")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return nil, errors.Wrap(err, "Get sql error when get pool list")
	}

	poolIdMap := make(map[int64]string, 0)
	for _, pool := range poolList {
		if pool.TenantID == tenantID {
			poolIdMap[pool.ResourcePoolID] = pool.Name
		}
	}
	zoneMap := make(map[string]string, 0)
	unitList, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return nil, errors.Wrap(err, "Get sql error when get unit list")
	}
	for _, unit := range unitList {
		if poolIdMap[unit.ResourcePoolID] != "" && zoneMap[unit.Zone] == "" {
			zoneMap[unit.Zone] = unit.Zone
		}
	}
	for k, _ := range zoneMap {
		zoneList = append(zoneList, k)
	}
	return zoneList, nil
}



func (m *OBTenantManager) GenerateStatusUnitNumMap(zones []v1alpha1.ResourcePoolSpec) (map[string]int, error) {
	unitNumMap := make(map[string]int, 0)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get sql error when get pool list")
	}
	for _, zone := range zones {
		poolName := m.GeneratePoolName(zone.ZoneList)
		for _, pool := range poolList {
			if pool.Name == poolName {
				unitNumMap[zone.ZoneList] = int(pool.UnitNum)
			}
		}
	}
	return unitNumMap, nil
}

func (m *OBTenantManager) BuildUnitConfigV4FromDB(zone string, tenantID int64) (v1alpha1.UnitConfig, error) {
	var unitConfig v1alpha1.UnitConfig
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	unitList, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get unit list")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get pool list")
	}
	unitConfigList, err := oceanbaseOperationManager.GetUnitConfigV4List()
	if err != nil {
		return unitConfig, errors.Wrap(err, "Get sql error when get unit config list")
	}
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) && pool.TenantID == tenantID {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						unitConfig.MaxCPU = kuberesource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						unitConfig.MinCPU = kuberesource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						unitConfig.MemorySize = *kuberesource.NewQuantity(unitConifg.MemorySize, kuberesource.DecimalSI)
						unitConfig.LogDiskSize = *kuberesource.NewQuantity(unitConifg.LogDiskSize, kuberesource.DecimalSI)
						unitConfig.MaxIops = int(unitConifg.MaxIops)
						unitConfig.MinIops = int(unitConifg.MinIops)
						unitConfig.IopsWeight = int(unitConifg.IopsWeight)
					}
				}
			}
		}
	}
	return unitConfig, nil
}

func (m *OBTenantManager) BuildUnitStatusFromDB(zone string, tenantID int64) ([]v1alpha1.UnitStatus, error) {
	var unitList []v1alpha1.UnitStatus
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return unitList, errors.Wrap(err, "Get sql error when get pool list")
	}
	var resourcePoolIDList []int64
	for _, pool := range poolList {
		if pool.TenantID == tenantID {
			resourcePoolIDList = append(resourcePoolIDList, pool.ResourcePoolID)
		}
	}
	units, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	for _, unit := range units {
		for _, poolId := range resourcePoolIDList {
			if unit.Zone == zone && poolId == unit.ResourcePoolID {
				var res v1alpha1.UnitStatus
				res.UnitId = int(unit.UnitID)
				res.ServerIP = unit.SvrIP
				res.ServerPort = int(unit.SvrPort)
				res.Status = unit.Status
				var migrateServer v1alpha1.MigrateServerStatus
				if unit.MigrateFromSvrIP.Valid {
					migrateServer.ServerIP = unit.MigrateFromSvrIP.String
				}
				if unit.MigrateFromSvrPort.Valid {
					migrateServer.ServerPort = int(unit.MigrateFromSvrPort.Int64)
				}
				res.Migrate = migrateServer
				unitList = append(unitList, res)
			}
		}
	}
	return unitList, nil
}


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
	m.Logger.Info("Check Whether The Tenant Exists", "tenantName", tenantName)
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
	m.Logger.Info("Check Whether The Resource Pool Exists", "poolName", poolName)
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
	m.Logger.Info("Check Whether The Resource Unit Exists", "unitConfigName", unitConfigName)
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
		UnitNum:  int64(pool.UnitNumber),
		ZoneList: pool.ZoneList,
	}
	return oceanbaseOperationManager.AddPool(poolSQLParam)
}

func (m *OBTenantManager) CreateUnitAndPoolV4(pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Spec.TenantName
	unitName := m.GenerateUnitName(pool.ZoneList)
	poolName := m.GeneratePoolName(pool.ZoneList)
	poolExist, err := m.PoolExist(poolName)
	if err != nil {
		m.Logger.Error(err, "Check Resource Pool Exist Failed", "tenantName", tenantName, "poolName", poolName)
		return err
	}

	unitExist, err := m.UnitConfigV4Exist(unitName)
	if err != nil {
		m.Logger.Error(err, "Check UnitConfig Exist Failed", "tenantName", tenantName, "unitName", unitName)
		return err
	}

	if !unitExist {
		err = m.CreateUnitConfigV4(unitName, pool.UnitConfig)
		if err != nil {
			m.Logger.Error(err, "Create UnitConfigV4 Failed", "tenantName", tenantName, "unitName", unitName)
			return err
		}
	}
	if !poolExist {
		err = m.CreatePool(poolName, unitName, pool)
		if err != nil {
			m.Logger.Error(err,"Create Tenant Failed", "tenantName", tenantName, "poolName", poolName)
			return err
		}
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
		return oceanbaseOperationManager.DeleteTenant(tenantName)
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
