/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func (ctrl *TenantCtrl) CheckAndSetVariables() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Variables For Tenant ", ctrl.Tenant.Name))
	}
	tenant := sqlOperator.GetTenantByName(ctrl.Tenant.Name)
	if len(tenant) == 0 {
		return errors.New(fmt.Sprint("Cannot Get Tenant For CheckAndSetVariables: ", ctrl.Tenant.Name))
	}
	tenantID := int(tenant[0].TenantID)
	for _, variable := range ctrl.Tenant.Spec.Variables {
		currentVariables := sqlOperator.GetVariable(variable.Name, tenantID)
		match := true
		for _, currentVariable := range currentVariables {
			if currentVariable.Value != variable.Value {
				klog.Infof("found variable %s with value %s did't match with config %s", variable.Name, currentVariable.Value, variable.Value)
				match = false
				break
			}
		}
		if !match {
			klog.Infof("set variable %s = %s", variable.Name, variable.Value)
			err = sqlOperator.SetTenantVariable(ctrl.Tenant.Name, variable.Name, variable.Value)
			if err != nil {
				return err
			}
			err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ctrl *TenantCtrl) CheckAndSetUnitConfig() error {
	version, err := ctrl.GetOBVersion()
	if err != nil {
		return err
	}
	switch string(version[0]) {
	case tenantconst.Version3:
		return ctrl.CheckAndSetUnitV3Config()
	case tenantconst.Version4:
		return ctrl.CheckAndSetUnitV4Config()
	}
	return errors.New("no match version for check and set unit config")
}

func (ctrl *TenantCtrl) CheckAndSetUnitV3Config() error {
	tenantName := ctrl.Tenant.Name
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", ctrl.Tenant.Name))
	}
	specResourceUnit := GenerateSpecResourceUnitV3Map(ctrl.Tenant.Spec)
	statusResourceUnit := GenerateStatusResourceUnitV3Map(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		match := true
		if !ctrl.isUnitV3Equal(specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName]) {
			klog.Infof("found zone '%s' unit config with value '%s' did't match with config '%s'", zone.ZoneName, ctrl.FormatUnitV3Config(specResourceUnit[zone.ZoneName]), ctrl.FormatUnitV3Config(statusResourceUnit[zone.ZoneName]))
			match = false
		}
		if !match {
			err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			if err != nil {
				return err
			}
			klog.Infof("set zone '%s' unit config '%s'", zone.ZoneName, ctrl.FormatUnitV3Config(specResourceUnit[zone.ZoneName]))
			unitName := ctrl.GenerateUnitName(ctrl.Tenant.Name, zone.ZoneName)
			err, unitExist := ctrl.UnitExist(unitName)
			if err != nil {
				klog.Errorf("Check Tenant '%s' Whether The Resource Unit '%s' Exists Error: %s", tenantName, unitName, err)
				return err
			}
			if !unitExist {
				err = ctrl.CreateUnit(unitName, zone.ResourceUnits)
				if err != nil {
					klog.Errorf("Create Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
					return err
				}
			} else {
				err = sqlOperator.SetUnitConfigV3(unitName, specResourceUnit[zone.ZoneName])
				if err != nil {
					return err
				}
			}
		}
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) CheckAndSetUnitV4Config() error {
	tenantName := ctrl.Tenant.Name
	specResourceUnit := GenerateSpecResourceUnitV4Map(ctrl.Tenant.Spec)
	statusResourceUnit := GenerateStatusResourceUnitV4Map(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		match := true
		if !ctrl.isUnitV4Equal(specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName]) {
			klog.Infof("found zone '%s' unit config with value '%s' did't match with config '%s'", zone.ZoneName, ctrl.FormatUnitV4Config(specResourceUnit[zone.ZoneName]), ctrl.FormatUnitV4Config(statusResourceUnit[zone.ZoneName]))
			match = false
		}
		if !match {
			err := ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			if err != nil {
				return err
			}
			unitName := ctrl.GenerateUnitName(ctrl.Tenant.Name, zone.ZoneName)
			err, unitExist := ctrl.UnitExist(unitName)
			if err != nil {
				klog.Errorf("Check Tenant '%s' Whether The Resource Unit '%s' Exists Error: %s", tenantName, unitName, err)
				return err
			}
			if !unitExist {
				err = ctrl.CreateUnit(unitName, zone.ResourceUnits)
				if err != nil {
					klog.Errorf("Create Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
					return err
				}
			} else {
				err = ctrl.SetUnitV4(unitName, specResourceUnit[zone.ZoneName])
				if err != nil {
					klog.Errorf("Set Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
					return err
				}
			}
		}
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) SetUnitV4(name string, resourceUnit model.ResourceUnitV4) error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", ctrl.Tenant.Name))
	}
	var option string
	if resourceUnit.MinCPU.Value() != 0 {
		option = fmt.Sprint(option, ", min_cpu ", resourceUnit.MinCPU.Value())
	}
	if resourceUnit.LogDiskSize.Value() != 0 {
		option = fmt.Sprint(option, ", log_disk_size ", resourceUnit.LogDiskSize.Value())
	}
	if resourceUnit.MaxIops != 0 {
		option = fmt.Sprint(option, ", max_iops ", resourceUnit.MaxIops)
	}
	if resourceUnit.MinIops != 0 {
		option = fmt.Sprint(option, ", min_iops ", resourceUnit.MinIops)
	}
	if resourceUnit.IopsWeight != 0 {
		option = fmt.Sprint(option, ", iops_weight ", resourceUnit.IopsWeight)
	}
	return sqlOperator.SetUnitConfigV4(name, resourceUnit, option)
}

func (ctrl *TenantCtrl) CheckAndSetResourcePool() error {
	tenantName := ctrl.Tenant.Name
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Resource Pool For Tenant ", ctrl.Tenant.Name))
	}
	specUnitNumMap := GenerateSpecUnitNumMap(ctrl.Tenant.Spec)
	statusUnitNumMap := GenerateStatusUnitNumMap(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		if specUnitNumMap[zone.ZoneName] != statusUnitNumMap[zone.ZoneName] {
			klog.Infof("found zone %s resource pool with unit_num value %d did't match with config %d", zone.ZoneName, statusUnitNumMap[zone.ZoneName], statusUnitNumMap[zone.ZoneName])
			err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			if err != nil {
				return err
			}
			poolName := ctrl.GeneratePoolName(ctrl.Tenant.Name, zone.ZoneName)
			poolExist, _, err := ctrl.PoolExist(poolName)
			if err != nil {
				klog.Errorf("Check Tenant '%s' Whether The Resource Pool '%s' Exists Error: %s", tenantName, poolName, err)
				return err
			}
			if !poolExist {
				unitName := ctrl.GenerateUnitName(tenantName, zone.ZoneName)
				err = ctrl.CreatePool(poolName, unitName, zone)
				if err != nil {
					klog.Errorf("Create Tenant '%s' Pool '%s' Error: %s", tenantName, poolName, err)
					return err
				}
			} else {
				klog.Infof("set zone %s resource pool unit_num %d", zone.ZoneName, specUnitNumMap[zone.ZoneName])
				err = sqlOperator.SetPoolUnitNum(poolName, specUnitNumMap[zone.ZoneName])
				if err != nil {
					return err
				}
			}
		}
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) CheckAndSetTenant() error {
	var err error
	addZone := ctrl.GetZoneForAdd()
	if addZone.ZoneName != "" {
		err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
		if err != nil {
			return err
		}
		return ctrl.TenantAddZone(addZone)
	}
	deleteZone := ctrl.GetZoneForDelete()
	if deleteZone.ZoneName != "" {
		err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
		if err != nil {
			return err
		}
		return ctrl.TenantDeleteZone(deleteZone)
	}
	err = ctrl.CheckAndSetPriority()
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetLocality()
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetTenantParams()
	if err != nil {
		return err
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) GetZoneForAdd() v1.TenantReplica {
	var zone v1.TenantReplica
	for _, specZone := range ctrl.Tenant.Spec.Topology {
		exist := false
		for _, statusZone := range ctrl.Tenant.Status.Topology {
			if statusZone.ZoneName == specZone.ZoneName {
				exist = true
			}
		}
		if !exist {
			zone = specZone
			break
		}
	}
	return zone
}

func (ctrl *TenantCtrl) GetZoneForDelete() v1.TenantReplicaStatus {
	var zone v1.TenantReplicaStatus
	for _, statusZone := range ctrl.Tenant.Status.Topology {
		exist := false
		for _, specZone := range ctrl.Tenant.Spec.Topology {
			if statusZone.ZoneName == specZone.ZoneName {
				exist = true
			}
		}
		if !exist {
			zone = statusZone
			break
		}
	}
	return zone
}

func (ctrl *TenantCtrl) TenantAddZone(zone v1.TenantReplica) error {
	tenantName := ctrl.Tenant.Name
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Add Zone ", ctrl.Tenant.Name))
	}
	tenantStatusReplica := v1.TenantReplicaStatus{
		ZoneName:   zone.ZoneName,
		Type:       zone.Type,
		UnitNumber: zone.UnitNumber,
	}
	tenantStatusReplicaList := ctrl.Tenant.Status.Topology
	tenantStatusReplicaList = append(tenantStatusReplicaList, tenantStatusReplica)
	err = ctrl.CheckAndCreateUnitAndPool(zone)
	if err != nil {
		return err
	}
	var localityString string
	poolList := ctrl.GenerateStatusPoolList(tenantName, tenantStatusReplicaList)
	poolListString := fmt.Sprintf("RESOURCE_POOL_LIST = ('%s')", strings.Join(poolList, "','"))
	statusLocalityMap := GenerateStatusLocalityMap(tenantStatusReplicaList)
	localityList := ctrl.GenerateLocalityList(statusLocalityMap)
	localityString = fmt.Sprintf(", LOCALITY = '%s'", strings.Join(localityList, ","))
	err = sqlOperator.SetTenant(tenantName, "", "", poolListString, "", localityString)
	if err != nil {
		return err
	}
	klog.Infof("Wait For Tenant '%s' 'ALTER_TENANT_LOCALITY' Job Success", tenantName)
	for {
		jobList := sqlOperator.GetInprogressJob(tenantName)
		if len(jobList) == 0 {
			break
		}
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) TenantDeleteZone(deleteZone v1.TenantReplicaStatus) error {
	tenantName := ctrl.Tenant.Name
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Delete Zone ", ctrl.Tenant.Name))
	}
	var zoneList []v1.TenantReplicaStatus
	for _, zone := range ctrl.Tenant.Status.Topology {
		if zone.ZoneName != deleteZone.ZoneName {
			zoneList = append(zoneList, zone)
		}
	}
	statusLocalityMap := GenerateStatusLocalityMap(zoneList)
	localityList := ctrl.GenerateLocalityList(statusLocalityMap)
	localityString := fmt.Sprintf("LOCALITY = '%s'", strings.Join(localityList, ","))
	err = sqlOperator.SetTenant(tenantName, "", "", "", "", localityString)
	if err != nil {
		klog.Errorf("Modify Tenant '%s' Locality Error : %s", tenantName, err)
		return err
	}
	klog.Infof("Wait For Tenant '%s' 'ALTER_TENANT_LOCALITY' Job Success", tenantName)
	for {
		jobList := sqlOperator.GetInprogressJob(tenantName)
		if len(jobList) == 0 {
			break
		}
	}
	poolList := ctrl.GenerateStatusPoolList(tenantName, zoneList)
	poolListString := fmt.Sprintf(", RESOURCE_POOL_LIST = ('%s')", strings.Join(poolList, "','"))
	err = sqlOperator.SetTenant(tenantName, "", "", poolListString, "", "")
	if err != nil {
		klog.Errorf("Modify Tenant '%s' Resource Pool List Error : %s", tenantName, err)
		return err
	}
	poolName := ctrl.GeneratePoolName(tenantName, deleteZone.ZoneName)
	poolExist, _, err := ctrl.PoolExist(poolName)
	if err != nil {
		klog.Errorln("Check Whether The Resource Pool Exists Error: ", err)
		return err
	}
	if poolExist {
		err = sqlOperator.DeletePool(poolName)
		if err != nil {
			return err
		}
	}
	unitName := ctrl.GenerateUnitName(tenantName, deleteZone.ZoneName)
	err, unitExist := ctrl.UnitExist(unitName)
	if err != nil {
		klog.Errorln("Check Whether The Resource Unit Exists Error: ", err)
		return err
	}
	if unitExist {
		err = sqlOperator.DeleteUnit(unitName)
		if err != nil {
			return err
		}
	}
	klog.Infoln("Succeed delete zone  ", deleteZone.ZoneName)
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) CheckAndSetPriority() error {
	specPrimaryZone := ctrl.GenerateSpecPrimaryZone(ctrl.Tenant.Spec.Topology)
	statusPrimaryZone := ctrl.GenerateStatusPrimaryZone(ctrl.Tenant.Status.Topology)
	specPrimaryZoneMap := ctrl.GeneratePrimaryZoneMap(specPrimaryZone)
	statusPrimaryZoneMap := ctrl.GeneratePrimaryZoneMap(statusPrimaryZone)
	if reflect.DeepEqual(specPrimaryZoneMap, statusPrimaryZoneMap) {
		return nil
	}
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Priority ", ctrl.Tenant.Name))
	}
	err = sqlOperator.SetTenant(ctrl.Tenant.Name, "", fmt.Sprint("PRIMARY_ZONE = '", specPrimaryZone, "'"), "", "", "")
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *TenantCtrl) GeneratePrimaryZoneMap(str string) map[int][]string {
	res := make(map[int][]string, 0)
	levelCuts := strings.Split(str, ";")
	for idx, levelCut := range levelCuts {
		cut := strings.Split(levelCut, ",")
		res[idx] = cut
		sort.Strings(res[idx])
	}
	return res
}

func (ctrl *TenantCtrl) CheckAndSetLocality() error {
	specLocalityMap := GenerateSpecLocalityMap(ctrl.Tenant.Spec)
	statusLocalityMap := GenerateStatusLocalityMap(ctrl.Tenant.Status.Topology)
	if reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		return nil
	}
	locality := ctrl.GenerateLocality(ctrl.Tenant.Spec.Topology)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Locality ", ctrl.Tenant.Name))
	}
	err = sqlOperator.SetTenant(ctrl.Tenant.Name, "", "", "", "", fmt.Sprint("LOCALITY = '", locality, "'"))
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *TenantCtrl) CheckAndSetTenantParams() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Params ", ctrl.Tenant.Name))
	}
	charset := ctrl.Tenant.Spec.Charset
	if charset != "" {
		charset = fmt.Sprintf("CHARSET = %s", charset)
	}
	if charset != "" {
		err = sqlOperator.SetTenant(ctrl.Tenant.Name, "", "", "", charset, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateSpecResourceUnitV3Map(spec v1.TenantSpec) map[string]model.ResourceUnitV3 {
	var resourceMap = make(map[string]model.ResourceUnitV3, 0)
	for _, zone := range spec.Topology {
		var resourceUnit model.ResourceUnitV3
		resourceUnit.MaxCPU = zone.ResourceUnits.MaxCPU
		resourceUnit.MinCPU = zone.ResourceUnits.MinCPU
		resourceUnit.MemorySize = zone.ResourceUnits.MemorySize
		if zone.ResourceUnits.MaxIops == 0 {
			resourceUnit.MaxIops = tenantconst.MaxIops
		}
		if zone.ResourceUnits.MinIops == 0 {
			resourceUnit.MinIops = tenantconst.MinIops
		}
		if zone.ResourceUnits.MaxSessionNum == 0 {
			resourceUnit.MaxSessionNum = tenantconst.MaxSessionNum
		}
		if zone.ResourceUnits.MaxDiskSize.String() == "0" {
			resourceUnit.MaxDiskSize = resource.MustParse(tenantconst.MaxDiskSize)
		}
		resourceMap[zone.ZoneName] = resourceUnit
	}
	return resourceMap
}

func GenerateSpecResourceUnitV4Map(spec v1.TenantSpec) map[string]model.ResourceUnitV4 {
	var resourceMap = make(map[string]model.ResourceUnitV4, 0)
	for _, zone := range spec.Topology {
		var resourceUnit model.ResourceUnitV4
		resourceUnit.MaxCPU = zone.ResourceUnits.MaxCPU
		resourceUnit.MinCPU = zone.ResourceUnits.MinCPU
		resourceUnit.MaxIops = zone.ResourceUnits.MaxIops
		resourceUnit.MinIops = zone.ResourceUnits.MinIops
		resourceUnit.IopsWeight = zone.ResourceUnits.IopsWeight
		resourceUnit.MemorySize = zone.ResourceUnits.MemorySize
		resourceUnit.LogDiskSize = zone.ResourceUnits.LogDiskSize
		resourceMap[zone.ZoneName] = resourceUnit
	}
	return resourceMap
}

func GenerateStatusResourceUnitV3Map(status v1.TenantStatus) map[string]model.ResourceUnitV3 {
	var resourceMap = make(map[string]model.ResourceUnitV3, 0)
	for _, zone := range status.Topology {
		var resourceUnit model.ResourceUnitV3
		resourceUnit.MaxCPU = zone.ResourceUnits.MaxCPU
		resourceUnit.MinCPU = zone.ResourceUnits.MinCPU
		resourceUnit.MemorySize = zone.ResourceUnits.MemorySize
		resourceUnit.MaxIops = zone.ResourceUnits.MaxIops
		resourceUnit.MinIops = zone.ResourceUnits.MinIops
		resourceUnit.MaxSessionNum = zone.ResourceUnits.MaxSessionNum
		resourceUnit.MaxDiskSize = zone.ResourceUnits.MaxDiskSize
		resourceMap[zone.ZoneName] = resourceUnit
	}
	return resourceMap
}

func GenerateStatusResourceUnitV4Map(status v1.TenantStatus) map[string]model.ResourceUnitV4 {
	var resourceMap = make(map[string]model.ResourceUnitV4, 0)
	for _, zone := range status.Topology {
		var resourceUnit model.ResourceUnitV4
		resourceUnit.MaxCPU = zone.ResourceUnits.MaxCPU
		resourceUnit.MinCPU = zone.ResourceUnits.MinCPU
		resourceUnit.MaxIops = zone.ResourceUnits.MaxIops
		resourceUnit.MinIops = zone.ResourceUnits.MinIops
		resourceUnit.IopsWeight = zone.ResourceUnits.IopsWeight
		resourceUnit.MemorySize = zone.ResourceUnits.MemorySize
		resourceUnit.LogDiskSize = zone.ResourceUnits.LogDiskSize
		resourceMap[zone.ZoneName] = resourceUnit
	}
	return resourceMap
}

func GenerateSpecUnitNumMap(spec v1.TenantSpec) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range spec.Topology {
		unitNumMap[zone.ZoneName] = zone.UnitNumber
	}
	return unitNumMap
}

func GenerateStatusUnitNumMap(status v1.TenantStatus) map[string]int {
	var unitNumMap = make(map[string]int, 0)
	for _, zone := range status.Topology {
		unitNumMap[zone.ZoneName] = zone.UnitNumber
	}
	return unitNumMap
}

func GenerateSpecLocalityMap(spec v1.TenantSpec) map[string]v1.TypeSpec {
	localityMap := make(map[string]v1.TypeSpec, 0)
	for _, zone := range spec.Topology {
		if zone.Type.Name != "" {
			switch strings.ToUpper(zone.Type.Name) {
			case tenantconst.TypeFull:
				localityMap[zone.ZoneName] = v1.TypeSpec{
					Name:    tenantconst.TypeFull,
					Replica: 1,
				}
			case tenantconst.TypeLogonly:
				localityMap[zone.ZoneName] = v1.TypeSpec{
					Name:    tenantconst.TypeLogonly,
					Replica: 1,
				}
			case tenantconst.TypeReadonly:
				var replica int
				if zone.Type.Replica == 0 {
					replica = 1
				} else {
					replica = zone.Type.Replica
				}
				localityMap[zone.ZoneName] = v1.TypeSpec{
					Name:    tenantconst.TypeReadonly,
					Replica: replica,
				}
			}
		}
	}
	return localityMap
}

func GenerateStatusLocalityMap(topology []v1.TenantReplicaStatus) map[string]v1.TypeSpec {
	localityMap := make(map[string]v1.TypeSpec, 0)
	for _, zone := range topology {
		localityMap[zone.ZoneName] = zone.Type
	}
	return localityMap
}

func (ctrl *TenantCtrl) GenerateLocalityList(localityMap map[string]v1.TypeSpec) []string {
	var locality []string
	for zoneName, zoneType := range localityMap {
		if zoneType.Name != "" {
			locality = append(locality, fmt.Sprint(zoneType.Name, "{", zoneType.Replica, "}@", zoneName))
		}
	}
	return locality
}

func (ctrl *TenantCtrl) isUnitV3Equal(specResourceUnit model.ResourceUnitV3, statusResourceUnit model.ResourceUnitV3) bool {
	if specResourceUnit.MaxCPU.Equal(statusResourceUnit.MaxCPU) &&
		specResourceUnit.MinCPU.Equal(statusResourceUnit.MinCPU) &&
		specResourceUnit.MemorySize.Value() == statusResourceUnit.MemorySize.Value() &&
		specResourceUnit.MaxIops == statusResourceUnit.MaxIops &&
		specResourceUnit.MinIops == statusResourceUnit.MinIops &&
		specResourceUnit.MaxDiskSize.Value() == statusResourceUnit.MaxDiskSize.Value() &&
		specResourceUnit.MaxSessionNum == statusResourceUnit.MaxSessionNum {
		return true
	} else {
		return false
	}
}

func (ctrl *TenantCtrl) isUnitV4Equal(specResourceUnit model.ResourceUnitV4, statusResourceUnit model.ResourceUnitV4) bool {
	if specResourceUnit.MaxCPU.Equal(statusResourceUnit.MaxCPU) &&
		specResourceUnit.MemorySize.Value() == statusResourceUnit.MemorySize.Value() {
		if (specResourceUnit.MinIops != 0 && specResourceUnit.MinIops != statusResourceUnit.MinIops) ||
			(specResourceUnit.MaxIops != 0 && specResourceUnit.MaxIops != statusResourceUnit.MaxIops) ||
			(specResourceUnit.MinCPU.Value() != 0 && specResourceUnit.MinCPU.Value() != statusResourceUnit.MinCPU.Value()) ||
			(specResourceUnit.LogDiskSize.Value() != 0 && specResourceUnit.LogDiskSize.Value() != statusResourceUnit.LogDiskSize.Value()) ||
			(specResourceUnit.IopsWeight != 0 && specResourceUnit.IopsWeight != statusResourceUnit.IopsWeight) {
			return false
		}
		return true
	}
	return false
}

func (ctrl *TenantCtrl) FormatUnitV3Config(unit model.ResourceUnitV3) string {
	return fmt.Sprintf("MaxCPU: %s MinCPU:%s MemorySize:%s MaxIops:%d MinIops:%d MaxDiskSize:%s MaxSessionNum:%d",
		unit.MaxCPU.String(), unit.MinCPU.String(), unit.MemorySize.String(), unit.MaxIops, unit.MinIops, unit.MaxDiskSize.String(), unit.MaxSessionNum)
}

func (ctrl *TenantCtrl) FormatUnitV4Config(unit model.ResourceUnitV4) string {
	return fmt.Sprintf("MaxCPU: %s MinCPU:%s MemorySize:%s MaxIops:%d MinIops:%d IopsWeight:%d LogDiskSize:%s",
		unit.MaxCPU.String(), unit.MinCPU.String(), unit.MemorySize.String(), unit.MaxIops, unit.MinIops, unit.IopsWeight, unit.LogDiskSize.String())
}
