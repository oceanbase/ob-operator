package resource

import (
	"fmt"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/const/status/obtenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ---------- task entry point ----------

func (m *OBTenantManager) CreateTenantTask() error {
	tenantName := m.OBTenant.Name
	tenantExist, _, err := m.TenantExist(tenantName)
	if err != nil {
		klog.Errorf("Check Whether The Tenant %s Exists Error: %s", tenantName, err)
		return err
	}
	if tenantExist {
		m.OBTenant.Status.Status = obtenant.Pending
		klog.Errorf("%s has existed and it is not a controlled resource of k8s, please manually delete existing tenants or change name", tenantName)
		return errors.Wrapf(err, "%s has existed and it is not a controlled resource of k8s, please manually delete existing tenants or change name", tenantName)
	}

	for _, pool := range m.OBTenant.Spec.Pools{
		err := m.CreateUnitAndPool(pool)
		if err != nil {
			return err
		}
	}

	err = m.CreateTenant()
	if err != nil {
		klog.Errorf("Create Tenant '%s' Error: %s", tenantName, err)
		return err
	}

	klog.Infof("Create Tenant '%s' OK", tenantName)
	m.OBTenant.Status.Status = obtenant.Running
	return nil
}

func (m *OBTenantManager) MaintainTenantTask() error {
	tenantName := m.OBTenant.Name

	err := m.CheckAndApplyTcpInvitedNode()
	if err != nil {
		klog.Errorf("Maintain Tenant %s Check And Set Tcp Invited Node Error: %s", tenantName, err)
		return err
	}
	err = m.CheckAndApplyUnitConfigV4()
	if err != nil {
		klog.Errorf("Maintain Tenant %s Check And Apply UnitConfigV4 Error: %s", tenantName, err)
		return err
	}
	err = m.CheckAndApplyResourcePool()
	if err != nil {
		klog.Errorf("Maintain Tenant %s Check And Apply Resource Pool Error: %s", tenantName, err)
		return err
	}
	err = m.CheckAndUpdateTenant()
	if err != nil {
		klog.Errorf("Maintain Tenant %s Check And Apply Tenant Error: %s", tenantName, err)
		return err
	}
	m.OBTenant.Status.Status = obtenant.Running
	return nil
}

func (m *OBTenantManager) hasModifiedTenantTask() bool {
	tenantName := m.OBTenant.Name

	hasModifiedTcpInvitedNode := m.hasModifiedTcpInvitedNode()
	if hasModifiedTcpInvitedNode {
		klog.Infof("Maintain Tenant %s Tcp Invited Node has modified", tenantName)
	}
	hasModifiedUnitConfigV4 := m.hasModifiedUnitConfigV4()
	if hasModifiedUnitConfigV4 {
		klog.Infof("Maintain Tenant %s UnitConfigV4 has modified", tenantName)
	}
	hasModifiedResourcePool := m.hasModifiedResourcePool()
	if hasModifiedResourcePool {
		klog.Infof("Maintain Tenant %s Resource Pool has modified", tenantName)
	}
	hasModifiedTenant := m.hasModifiedTenant()
	if hasModifiedTenant {
		klog.Infof("Maintain Tenant %s Tenant has modified", tenantName)
	}
	m.OBTenant.Status.Status = obtenant.Running
	return hasModifiedTcpInvitedNode || hasModifiedUnitConfigV4 || hasModifiedResourcePool || hasModifiedTenant
}

func (m *OBTenantManager) DeleteTenantTask() error {
	var err error
	tenantName := m.OBTenant.Name
	klog.Infof("Begin Delete Tenant '%s'", tenantName)
	err = m.DeleteTenant()
	if err != nil {
		return err
	}
	klog.Infof("Begin Delete Pool, Tenant '%s'", tenantName)
	err = m.DeletePool()
	if err != nil {
		return err
	}
	klog.Infof("Begin Delete Unit, Tenant '%s'", tenantName)
	err = m.DeleteUnit()
	if err != nil {
		return err
	}
	klog.Infof("Succeed Delete Tenant '%s'", tenantName)
	return nil
}

// ---------- Check And Update function ----------

func (m *OBTenantManager) CheckAndApplyTcpInvitedNode() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting ob_tcp_invited_nodes For Tenant ", m.OBTenant.Name))
	}

	specWhiteList := m.OBTenant.Spec.ConnectWhiteList
	statusWhiteList := m.OBTenant.Status.ConnectWhiteList

	if specWhiteList == "" {
		specWhiteList = tenant.DefaultOBTcpInvitedNodes
	}
	if statusWhiteList != specWhiteList {
		klog.Infof("found variable '%s' with specWhiteList '%s' did't match with config '%s'", tenant.OBTcpInvitedNodes, statusWhiteList, specWhiteList)
		variableList := m.GenerateVariableList(specWhiteList)
		err = oceanbaseOperationManager.SetTenantVariable(m.OBTenant.Name, variableList)
		if err != nil {
			return err
		}
		err = m.UpdateTenantStatusOBTcpInvitedNodes(specWhiteList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyUnitConfigV4() error {
	tenantName := m.OBTenant.Name
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
			klog.Infof("found pool '%s' unit config with value '%s' did't match with config '%s'", pool.ZoneList, FormatUnitV4Config(specUnitConfigMap[pool.ZoneList]), FormatUnitV4Config(statusUnitConfigMap[pool.ZoneList]))
			match = false
		}
		if !match {
			m.OBTenant.Status.Status = obtenant.Maintaining
			unitName := m.GenerateUnitName(pool.ZoneList)
			err := m.SetUnitConfigV4(unitName, specUnitConfigMap[pool.ZoneList])
			if err != nil {
				klog.Errorf("Set Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
				return err
			}
		}
	}
	return nil
}

func (m *OBTenantManager) CheckAndApplyResourcePool() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting ob_tcp_invited_nodes For Tenant ", m.OBTenant.Name))
	}

	// handle add pool
	for _, addPool := range m.GetPoolsForAdd(){
		if addPool.ZoneList != "" {
			m.OBTenant.Status.Status = obtenant.Maintaining
			err := m.TenantAddPool(addPool)
			if err != nil {
				return err
			}
		}
	}

	// handle delete pool
	poolStatuses := m.GetPoolsForDelete()
	if len(poolStatuses) > 0 {
		m.OBTenant.Status.Status = obtenant.Maintaining
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
			klog.Infof("found pool %s resource pool with unit_num value %d did't match with config %d", pool.ZoneList, statusUnitNum, specUnitNum)
			m.OBTenant.Status.Status = obtenant.Maintaining
			poolName := m.GeneratePoolName(pool.ZoneList)
			klog.Infof("set pool %s resource pool unit_num %d", pool.ZoneList, specUnitNumMap)
			err = oceanbaseOperationManager.SetPoolUnitNum(poolName, specUnitNum)
			if err != nil {
				return err
			}
		}
	}
	return nil
}



func (m *OBTenantManager) CheckAndUpdateTenant() error {
	var err error
	err = m.CheckAndSetPriority()
	if err != nil {
		return err
	}
	err = m.CheckAndSetLocality()
	if err != nil {
		return err
	}
	err = m.CheckAndUpdateCharset()
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndSetPriority() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()

	specPrimaryZone := GenerateSpecPrimaryZone(m.OBTenant.Spec.Pools)
	statusPrimaryZone := GenerateStatusPrimaryZone(m.OBTenant.Status.Pools)
	specPrimaryZoneMap := GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := GeneratePrimaryZoneMap(statusPrimaryZone)
	if reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		return nil
	}
	m.OBTenant.Status.Status = obtenant.Maintaining
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Priority ", m.OBTenant.Name))
	}
	err = oceanbaseOperationManager.SetTenant(m.OBTenant.Name, "", fmt.Sprint("PRIMARY_ZONE = '", specPrimaryZone, "'"), "", "", "")
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndSetLocality() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	specLocalityMap := GenerateSpecLocalityMap(m.OBTenant.Spec.Pools)
	statusLocalityMap := GenerateStatusLocalityMap(m.OBTenant.Status.Pools)
	if reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		return nil
	}
	locality := m.GenerateLocality(m.OBTenant.Spec.Pools)
	m.OBTenant.Status.Status = obtenant.Maintaining
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Locality ", m.OBTenant.Name))
	}
	err = oceanbaseOperationManager.SetTenant(m.OBTenant.Name, "", "", "", "", fmt.Sprint("LOCALITY = '", locality, "'"))
	if err != nil {
		return err
	}
	return nil
}

func (m *OBTenantManager) CheckAndUpdateCharset() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Params ", m.OBTenant.Name))
	}
	charset := m.OBTenant.Spec.Charset
	if charset != "" {
		charset = fmt.Sprintf("CHARSET = %s", charset)
	}
	if charset != "" {
		err = oceanbaseOperationManager.SetTenant(m.OBTenant.Name, "", "", "", charset, "")
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
		match := true
		specUnitConfig := specUnitConfigMap[pool.ZoneList]
		statusUnitConfig, statusExist := statusUnitConfigMap[pool.ZoneList]

		// If status does not exist, Continue to check UnitConfig of the next ResourcePool
		// while Add and delete a pool in the CheckAndApplyResourcePool
		if !statusExist{
			continue
		}

		if !IsUnitConfigV4Equal(specUnitConfig, statusUnitConfig) {
			klog.Infof("found pool '%s' unit config with value '%s' did't match with config '%s'", pool.ZoneList, FormatUnitV4Config(specUnitConfigMap[pool.ZoneList]), FormatUnitV4Config(statusUnitConfigMap[pool.ZoneList]))
			match = false
		}
		if !match {
			return true
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


func (m *OBTenantManager) SetTcpInvitedNode() error {
	value := m.OBTenant.Spec.ConnectWhiteList
	currentValue := m.OBTenant.Status.ConnectWhiteList
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting ob_tcp_invited_nodes For Tenant ", m.OBTenant.Name))
	}
	if value == "" {
		value = tenant.DefaultOBTcpInvitedNodes
	}
	if currentValue != value {
		klog.Infof("found variable '%s' with value '%s' did't match with config '%s'", tenant.OBTcpInvitedNodes, currentValue, value)
		variableList := m.GenerateVariableList(value)
		err = oceanbaseOperationManager.SetTenantVariable(m.OBTenant.Name, variableList)
		if err != nil {
			return err
		}
		err = m.UpdateTenantStatusOBTcpInvitedNodes(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *OBTenantManager) SetUnitConfig() error {
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


func (m *OBTenantManager) CreateTenant() error {
	tenantName := m.OBTenant.Name
	pools := m.OBTenant.Spec.Pools
	klog.Infof("Create Tenant '%s'", tenantName)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	zoneList := m.GenerateSpecZoneList(pools)
	primaryZone := GenerateSpecPrimaryZone(pools)
	poolList := m.GenerateSpecPoolList(tenantName, pools)
	variableList := m.GenerateVariableList(m.OBTenant.Spec.ConnectWhiteList)
	charset := tenant.Charset
	locality := m.GenerateLocality(pools)
	collate := m.OBTenant.Spec.Collate
	return oceanbaseOperationManager.AddTenant(tenantName, charset, strings.Join(zoneList, "','"), primaryZone, strings.Join(poolList, "','"), locality, collate, variableList)
}

func (m *OBTenantManager) CreateUnitConfigV4(unitName string, unitConfig model.UnitConfigV4) error {
	klog.Infof("Create Tenant '%s' Resource Unit '%s' ", m.OBTenant.Name, unitName)
	if unitConfig.MemorySize == 0 {
		klog.Errorf("Tenant '%s'  resource unit '%s' memorySize cannot be empty", m.OBTenant.Name, unitName)
		return errors.Errorf("Tenant '%s'  resource unit '%s' memorySize cannot be empty", m.OBTenant.Name, unitName)
	}
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Unit")
	}

	return oceanbaseOperationManager.AddUnitConfigV4(unitName, unitConfig)
}

func (m *OBTenantManager) SetUnitConfigV4(name string, unitConfig model.UnitConfigV4) error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", m.OBTenant.Name))
	}
	return oceanbaseOperationManager.SetUnitConfigV4(name, unitConfig)
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
	tenantName := m.OBTenant.Name
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Add Zone ", m.OBTenant.Name))
	}
	resourcePoolStatus := v1alpha1.ResourcePoolStatus{
		ZoneList:   pool.ZoneList,
		Type:       pool.Type,
		UnitNumber: pool.UnitNumber,
	}
	resourcePoolStatusList := m.OBTenant.Status.Pools
	resourcePoolStatusList = append(resourcePoolStatusList, resourcePoolStatus)
	err = m.CreateUnitAndPool(pool)
	if err != nil {
		return err
	}
	var localityString string
	poolList := m.GenerateStatusPoolList(tenantName, resourcePoolStatusList)
	poolListString := fmt.Sprintf("RESOURCE_POOL_LIST = ('%s')", strings.Join(poolList, "','"))
	statusLocalityMap := GenerateStatusLocalityMap(resourcePoolStatusList)
	localityList := m.GenerateLocalityList(statusLocalityMap)
	localityString = fmt.Sprintf(", LOCALITY = '%s'", strings.Join(localityList, ","))
	err = oceanbaseOperationManager.SetTenant(tenantName, "", "", poolListString, "", localityString)
	if err != nil {
		return err
	}
	klog.Infof("Wait For Tenant '%s' 'ALTER_TENANT_LOCALITY' Job Success", tenantName)
	for {
		_, err := oceanbaseOperationManager.GetRsJob(tenantName)
		if err != nil {
			break
		}
	}
	return m.UpdateTenantStatus(obtenant.Running)
}

func (m *OBTenantManager) TenantDeletePool(deleteZone v1alpha1.ResourcePoolStatus) error {
	tenantName := m.OBTenant.Name
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Delete Zone ", m.OBTenant.Name))
	}
	var zoneList []v1alpha1.ResourcePoolStatus
	for _, zone := range m.OBTenant.Status.Pools {
		if zone.ZoneList != deleteZone.ZoneList {
			zoneList = append(zoneList, zone)
		}
	}
	statusLocalityMap := GenerateStatusLocalityMap(zoneList)
	localityList := m.GenerateLocalityList(statusLocalityMap)
	localityString := fmt.Sprintf("LOCALITY = '%s'", strings.Join(localityList, ","))
	err = oceanbaseOperationManager.SetTenant(tenantName, "", "", "", "", localityString)
	if err != nil {
		klog.Errorf("Modify Tenant '%s' Locality Error : %s", tenantName, err)
		return err
	}
	klog.Infof("Wait For Tenant '%s' 'ALTER_TENANT_LOCALITY' Job Success", tenantName)
	for {
		_, err := oceanbaseOperationManager.GetRsJob(tenantName)
		if err != nil {
			break
		}
	}
	poolList := m.GenerateStatusPoolList(tenantName, zoneList)
	poolListString := fmt.Sprintf(", RESOURCE_POOL_LIST = ('%s')", strings.Join(poolList, "','"))
	err = oceanbaseOperationManager.SetTenant(tenantName, "", "", poolListString, "", "")
	if err != nil {
		klog.Errorf("Modify Tenant '%s' Resource Pool List Error : %s", tenantName, err)
		return err
	}
	poolName := m.GeneratePoolName(deleteZone.ZoneList)
	poolExist, _, err := m.PoolExist(poolName)
	if err != nil {
		klog.Errorln("Check Whether The Resource Pool Exists Error: ", err)
		return err
	}
	if poolExist {
		err = oceanbaseOperationManager.DeletePool(poolName)
		if err != nil {
			return err
		}
	}
	unitName := m.GenerateUnitName(deleteZone.ZoneList)
	unitExist, err := m.UnitConfigV4Exist(unitName)
	if err != nil {
		klog.Errorln("Check Whether The Resource Unit Exists Error: ", err)
		return err
	}
	if unitExist {
		err = oceanbaseOperationManager.DeleteUnit(unitName)
		if err != nil {
			return err
		}
	}
	klog.Infoln("Succeed delete zone  ", deleteZone.ZoneList)
	return m.UpdateTenantStatus(obtenant.Running)
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

func FormatUnitV4Config(unit model.UnitConfigV4) string {
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

func GenerateStatusLocalityMap(topology []v1alpha1.ResourcePoolStatus) map[string]v1alpha1.LocalityType {
	localityMap := make(map[string]v1alpha1.LocalityType, 0)
	for _, zone := range topology {
		localityMap[zone.ZoneList] = zone.Type
	}
	return localityMap
}

func (m *OBTenantManager) GenerateLocalityList(localityMap map[string]v1alpha1.LocalityType) []string {
	var locality []string
	for zoneName, zoneType := range localityMap {
		if zoneType.Name != "" {
			locality = append(locality, fmt.Sprint(zoneType.Name, "{", zoneType.Replica, "}@", zoneName))
		}
	}
	return locality
}



func (m *OBTenantManager) GenerateSpecZoneList(zones []v1alpha1.ResourcePoolSpec) []string {
	var zoneList []string
	for _, zone := range zones {
		zoneList = append(zoneList, zone.ZoneList)
	}
	return zoneList
}

func (m *OBTenantManager) GenerateStatusZoneList(zones []v1alpha1.ResourcePoolStatus) []string {
	var zoneList []string
	for _, zone := range zones {
		zoneList = append(zoneList, zone.ZoneList)
	}
	return zoneList
}

func (m *OBTenantManager) GenerateSpecPoolList(tenantName string, zones []v1alpha1.ResourcePoolSpec) []string {
	var poolList []string
	for _, zone := range zones {
		poolName := m.GeneratePoolName(zone.ZoneList)
		poolList = append(poolList, poolName)
	}
	return poolList
}

func (m *OBTenantManager) GenerateStatusPoolList(tenantName string, zones []v1alpha1.ResourcePoolStatus) []string {
	var poolList []string
	for _, zone := range zones {
		poolName := m.GeneratePoolName(zone.ZoneList)
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

func (m *OBTenantManager) GenerateVariableList(variable string) string {
	if variable == "" {
		return fmt.Sprintf("SET VARIABLES %s = %s", tenant.OBTcpInvitedNodes, tenant.DefaultOBTcpInvitedNodes)
	} else {
		return fmt.Sprintf("SET VARIABLES %s = '%s'", tenant.OBTcpInvitedNodes, variable)
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

// ---------- update status function ----------

func (m *OBTenantManager) UpdateTenantStatus(tenantStatus string) error {
	tenantCurrent, err := m.getObTenant()
	if err != nil {
		return err
	}
	tenantCurrentDeepCopy := tenantCurrent.DeepCopy()
	m.OBTenant = tenantCurrentDeepCopy
	tenantNew, err := m.BuildTenantStatusForVariables(tenantCurrentDeepCopy, tenantStatus)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantCurrent.Status, tenantNew.Status)
	if !compareStatus {
		err = m.PatchStatus(tenantNew, tenantCurrent)
		if err != nil {
			return err
		}
	}
	m.OBTenant = tenantNew
	return nil
}

func (m *OBTenantManager) UpdateTenantStatusOBTcpInvitedNodes(value string) error {
	tenantCurrent, err := m.getObTenant()
	if err != nil {
		return err
	}
	tenantCurrentDeepCopy := tenantCurrent.DeepCopy()
	m.OBTenant = tenantCurrentDeepCopy
	tenantNew, err := m.BuildTenantStatusForVariables(tenantCurrentDeepCopy, value)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantCurrent.Status, tenantNew.Status)
	if !compareStatus {
		err = m.PatchStatus(tenantNew, tenantCurrent)
		if err != nil {
			return err
		}
	}
	m.OBTenant = tenantNew
	return nil
}

// ---------- buildTenant function ----------

func (m *OBTenantManager) BuildTenantStatusForVariables(tenant *v1alpha1.OBTenant, value string) (*v1alpha1.OBTenant, error) {
	var tenantCurrentStatus v1alpha1.OBTenantStatus
	tenantTopology, err := m.BuildPoolStatusList(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Status = tenant.Status.Status
	tenantCurrentStatus.Pools = tenantTopology
	tenantCurrentStatus.ConnectWhiteList = value
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Charset, err = m.GetCharset()
	if err != nil {
		return tenant, err
	}
	tenant.Status = tenantCurrentStatus
	return tenant, nil
}

func (m *OBTenantManager) BuildTenantStatus(tenant *v1alpha1.OBTenant, tenantStatus string) (*v1alpha1.OBTenant, error) {
	var tenantCurrentStatus v1alpha1.OBTenantStatus
	poolStatusList, err := m.BuildPoolStatusList(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Status = tenantStatus
	tenantCurrentStatus.Pools = poolStatusList
	tenantCurrentStatus.ConnectWhiteList = tenant.Status.ConnectWhiteList

	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Charset, err = m.GetCharset()
	if err != nil {
		return tenant, err
	}
	tenant.Status = tenantCurrentStatus
	return tenant, nil
}

func (m *OBTenantManager) BuildPoolStatusList(tenant *v1alpha1.OBTenant) ([]v1alpha1.ResourcePoolStatus, error) {
	var tenantTopologyStatusList []v1alpha1.ResourcePoolStatus
	var err error
	var locality string
	var primaryZone string
	gvTenant, err := m.GetTenantByName(m.OBTenant.Name)
	if err != nil {
		return tenantTopologyStatusList, err
	}
	if err != nil {
		return tenantTopologyStatusList, errors.New(fmt.Sprint("Cannot Get Tenant For BuildPoolStatusList: ", m.OBTenant.Name))
	}
	locality = gvTenant.Locality
	primaryZone = gvTenant.PrimaryZone
	typeMap := GenerateTypeMap(locality)
	tenantID := gvTenant.TenantID
	priorityMap := GeneratePriorityMap(primaryZone)
	unitNumMap, err := m.GenerateStatusUnitNumMap(tenant.Spec.Pools)
	if err != nil {
		return tenantTopologyStatusList, err
	}
	zoneList, err := m.GenerateStatusZone(tenantID)
	if err != nil {
		return tenantTopologyStatusList, err
	}
	for _, zone := range zoneList {
		var tenantCurrentStatus v1alpha1.ResourcePoolStatus
		tenantCurrentStatus.ZoneList = zone
		tenantCurrentStatus.Type = typeMap[zone]
		tenantCurrentStatus.UnitNumber = unitNumMap[zone]
		tenantCurrentStatus.Priority = priorityMap[zone]
		tenantCurrentStatus.UnitConfig, err = m.BuilUnitConfigV4FromDB(zone, tenantID)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantCurrentStatus.Units, err = m.BuildUnitStatusFromDB(zone, tenantID)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantTopologyStatusList = append(tenantTopologyStatusList, tenantCurrentStatus)
	}
	return tenantTopologyStatusList, nil
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
				unitNumMap[zone.ZoneList] = int(pool.UnitCount)
			}
		}
	}
	return unitNumMap, nil
}

func (m *OBTenantManager) BuilUnitConfigV4FromDB(zone string, tenantID int64) (v1alpha1.UnitConfig, error) {
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
						unitConfig.MaxCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						unitConfig.MinCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						unitConfig.MemorySize = *apiresource.NewQuantity(unitConifg.MemorySize, apiresource.DecimalSI)
						unitConfig.LogDiskSize = *apiresource.NewQuantity(unitConifg.LogDiskSize, apiresource.DecimalSI)
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
				migrateServer.ServerIP = unit.MigrateFromSvrIP
				migrateServer.ServerPort = int(unit.MigrateFromSvrPort)
				res.Migrate = migrateServer
				unitList = append(unitList, res)
			}
		}
	}
	return unitList, nil
}


func (m *OBTenantManager) GetOBVersion() (string, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Get OB Version")
	}
	version, err := oceanbaseOperationManager.GetVersion()
	if err != nil {
		return "", errors.Errorf("Tenant '%s' get ob version from db failed", m.OBTenant.Name)
	}
	return version.Version, nil
}

func (m *OBTenantManager) GenerateUnitName( zoneList string) string {
	unitName := fmt.Sprintf("unitconfig_%s_%s", m.OBTenant.Name, zoneList)
	return unitName
}

func (m *OBTenantManager) GeneratePoolName( zoneList string) string {
	poolName := fmt.Sprintf("pool_%s_%s", m.OBTenant.Name, zoneList)
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


func (m *OBTenantManager) GetTenantByName(tenantName string) (*model.Tenant, error) {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
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
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
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
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
	}
	unit, err := oceanbaseOperationManager.GetUnitConfigV4ByName(unitName)
	if err != nil {
		return nil, err
	}
	return unit, nil
}

func (m *OBTenantManager) TenantExist(tenantName string) (bool, int, error) {
	klog.Infof("Check Whether The Tenant '%s' Exists", tenantName)
	tenant, err := m.GetTenantByName(m.OBTenant.Name)
	if err != nil {
		return false, 0, err
	}
	return true, int(tenant.TenantID), nil
}

func (m *OBTenantManager) PoolExist(poolName string) (bool, int, error) {
	klog.Infof("Check Whether The Resource Pool '%s' Exists", poolName)
	pool, err := m.GetPoolByName(poolName)
	if err != nil {
		return false, 0, err
	}
	return true, int(pool.ResourcePoolID), nil
}

func (m *OBTenantManager) UnitConfigV4Exist(unitConfigName string) (bool, error) {
	klog.Infof("Check Whether The Resource Unit '%s' Exists", unitConfigName)
	_, err := m.GetUnitConfigV4ByName(unitConfigName)
	if  err != nil {
		return false, nil
	}
	return true, nil
}

func (m *OBTenantManager) CreatePool(poolName, unitName string, pool v1alpha1.ResourcePoolSpec) error {
	klog.Infof("Create Tenant '%s' Resource Pool %s ", m.OBTenant.Name, poolName)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	return oceanbaseOperationManager.AddPool(poolName, unitName, pool)
}

func (m *OBTenantManager) CreateUnitAndPool(pool v1alpha1.ResourcePoolSpec) error {
	tenantName := m.OBTenant.Name
	unitName := m.GenerateUnitName(pool.ZoneList)
	poolName := m.GeneratePoolName(pool.ZoneList)
	poolExist, _, err := m.PoolExist(poolName)
	if err != nil {
		klog.Errorf("Check Tenant '%s' Whether The Resource Pool '%s' Exists Error: %s", tenantName, poolName, err)
		return err
	}

	unitExist, err := m.UnitConfigV4Exist(unitName)
	if err != nil {
		klog.Errorf("Check Tenant '%s' Whether The Resource Unit '%s' Exists Error: %s", tenantName, unitName, err)
		return err
	}

	if !unitExist {
		version, err := m.GetOBVersion()
		if err != nil {
			return err
		}
		if string(version[0]) == tenant.Version3 {
			err := m.CheckResourceEnough(tenantName, pool)
			if err != nil {
				return err
			}
		}
		err = m.CreateUnitConfigV4(unitName, GenerateModelUnitConfigV4(pool.UnitConfig))
		if err != nil {
			klog.Errorf("Create Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
			return err
		}
	}
	if !poolExist {
		err = m.CreatePool(poolName, unitName, pool)
		if err != nil {
			klog.Errorf("Create Tenant '%s' Pool '%s' Error: %s", tenantName, poolName, err)
			return err
		}
	}
	return nil
}


func (m *OBTenantManager) CheckResourceEnough(tenantName string, zone v1alpha1.ResourcePoolSpec) error {
	klog.Infof("Check Tenant '%s' Zone '%s' Reousrce ", tenantName, zone.ZoneList)
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Checking Reousrce")
	}
	resource, err := oceanbaseOperationManager.GetResourceTotal(zone.ZoneList)
	if err != nil {
		return errors.Errorf("Tenant '%s' cannot get resource", tenantName)
	}
	unitList, err := oceanbaseOperationManager.GetUnitList()
	if err != nil {
		return errors.Wrap(err, "Get sql error when get unit list")
	}
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone.ZoneList {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	poolList, err := oceanbaseOperationManager.GetPoolList()
	if err != nil {
		return errors.Wrap(err, "Get sql error when get pool list")
	}
	unitConfigList, err := oceanbaseOperationManager.GetUnitConfigV4List()
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) {
				for _, unitConfig := range unitConfigList {
					if unitConfig.UnitConfigID == pool.UnitConfigID {
						resource.CPUTotal -= unitConfig.MaxCPU
					}
				}
			}
		}
	}
	if zone.UnitConfig.MaxCPU.AsApproximateFloat64() > resource.CPUTotal {
		return errors.New(fmt.Sprintf("Tenant '%s' Zone '%s' CPU Is Not Enough: Need %f, Only %f", tenantName, zone.ZoneList, zone.UnitConfig.MaxCPU.AsApproximateFloat64(), resource.CPUTotal))
	}
	maxMem := zone.UnitConfig.MemorySize.Value()
	if err != nil {
		return err
	}
	if maxMem > resource.MemTotal {
		return errors.New(fmt.Sprintf("Tenant '%s' Zone '%s' Memory Is Not Enough: Need %d, Only %d", tenantName, zone.ZoneList, int(maxMem), int(resource.MemTotal)))
	}
	return nil
}


func (m *OBTenantManager) DeleteTenant() error {
	tenantName := m.OBTenant.Name
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Tenant ", tenantName))
	}

	tenantExist, _, err := m.TenantExist(m.OBTenant.Name)
	if err != nil {
		klog.Errorf("Check Whether The Tenant '%s' Exists Error: %s", tenantName, err)
		return err
	}
	if tenantExist {
		return oceanbaseOperationManager.DeleteTenant(tenantName)
	}
	return nil
}

func (m *OBTenantManager) DeletePool() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Pool", m.OBTenant.Name))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		poolName := m.GeneratePoolName(zone.ZoneList)
		poolExist, _, err := m.PoolExist(poolName)
		if err != nil {
			klog.Errorln("Check Whether The Resource Pool Exists Error: ", err)
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

func (m *OBTenantManager) DeleteUnit() error {
	oceanbaseOperationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Deleting Unit", m.OBTenant.Name))
	}
	for _, zone := range m.OBTenant.Spec.Pools {
		unitName := m.GenerateUnitName(zone.ZoneList)
		unitExist, err := m.UnitConfigV4Exist(unitName)
		if err != nil {
			klog.Errorln("Check Whether The Resource Unit Exists Error: ", err)
			return err
		}
		if unitExist {
			err = oceanbaseOperationManager.DeleteUnit(unitName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
