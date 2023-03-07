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
	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"reflect"
	"strings"
)

func (ctrl *TenantCtrl) CheckAndSetVariables() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Variables For Tenant ", ctrl.Tenant.Name))
	}
	tenant := sqlOperator.GetGvTenantByName(ctrl.Tenant.Name)
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
	tenantName := ctrl.Tenant.Name
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", ctrl.Tenant.Name))
	}
	specResourceUnit := GenerateSpecResourceUnitMap(ctrl.Tenant.Spec)
	statusResourceUnit := GenerateStatusResourceUnitMap(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		match := true
		if !ctrl.isUnitEqual(specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName]) {
			klog.Infof("found zone '%s' unit config with value '%s' did't match with config '%s'", zone.ZoneName, ctrl.FormatUnitConfig(specResourceUnit[zone.ZoneName]), ctrl.FormatUnitConfig(statusResourceUnit[zone.ZoneName]))
			match = false
		}
		if !match {
			err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			if err != nil {
				return err
			}
			klog.Infof("set zone '%s' unit config '%s'", zone.ZoneName, ctrl.FormatUnitConfig(specResourceUnit[zone.ZoneName]))
			unitName := ctrl.GenerateUnitName(ctrl.Tenant.Name, zone.ZoneName)
			err, unitExist := ctrl.UnitExist(unitName)
			if err != nil {
				klog.Errorf("Check Tenant '%s' Whether The Resource Unit '%s' Exists Error: %s", tenantName, unitName, err)
				return err
			}
			if !unitExist {
				err := ctrl.CheckResourceEnough(tenantName, zone)
				if err != nil {
					return err
				}
				err = ctrl.CreateUnit(unitName, zone.ResourceUnits)
				if err != nil {
					klog.Errorf("Create Tenant '%s' Unit '%s' Error: %s", tenantName, unitName, err)
					return err
				}
			} else {
				err = sqlOperator.SetUnitConfig(unitName, specResourceUnit[zone.ZoneName])
				if err != nil {
					return err
				}
			}
		}
	}
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
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
	klog.Infoln("tenantStatusReplicaList: ", tenantStatusReplicaList)
	err = ctrl.CheckAndCreateUnitAndPool(tenantName, zone)
	if err != nil {
		return err
	}
	var localityString string
	poolList := ctrl.GenerateStatusPoolList(tenantName, tenantStatusReplicaList)
	poolListString := fmt.Sprintf("RESOURCE_POOL_LIST = ('%s')", strings.Join(poolList, "','"))
	statusLocalityMap := GenerateStatusLocalityMap(tenantStatusReplicaList)
	localityList := ctrl.GenerateLocalityList(statusLocalityMap)
	localityString = fmt.Sprintf(", LOCALITY = '%s'", strings.Join(localityList, ","))
	err = sqlOperator.SetTenant(tenantName, "", "", poolListString, "", localityString, "")
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
	err = sqlOperator.SetTenant(tenantName, "", "", "", "", localityString, "")
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
	err = sqlOperator.SetTenant(tenantName, "", "", poolListString, "", "", "")
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

func (ctrl *TenantCtrl) CheckAndSetTenantParams() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get Sql Operator When Prcoessing Tenant '%s' Params ", ctrl.Tenant.Name))
	}
	charset := ctrl.Tenant.Spec.Charset
	if charset != "" {
		charset = fmt.Sprintf("CHARSET = %s", charset)
	}
	var logonlyReplicaNumString string
	if ctrl.Tenant.Spec.LogonlyReplicaNum != ctrl.Tenant.Status.LogonlyReplicaNum {
		logonlyReplicaNumString = fmt.Sprintf(", LOGONLY_REPLICA_NUM = %d", ctrl.Tenant.Spec.LogonlyReplicaNum)
	}
	if charset != "" || logonlyReplicaNumString != "" {
		err = sqlOperator.SetTenant(ctrl.Tenant.Name, "", "", "", charset, "", logonlyReplicaNumString)
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateSpecResourceUnitMap(spec v1.TenantSpec) map[string]v1.ResourceUnit {
	var resourceMap = make(map[string]v1.ResourceUnit, 0)
	for _, zone := range spec.Topology {
		if zone.ResourceUnits.MaxIops == 0 {
			zone.ResourceUnits.MaxIops = tenantconst.MaxIops
		}
		if zone.ResourceUnits.MinIops == 0 {
			zone.ResourceUnits.MinIops = tenantconst.MinIops
		}
		if zone.ResourceUnits.MaxSessionNum == 0 {
			zone.ResourceUnits.MaxSessionNum = tenantconst.MaxSessionNum
		}
		if zone.ResourceUnits.MaxDiskSize.String() == "0" {
			zone.ResourceUnits.MaxDiskSize = resource.MustParse(tenantconst.MaxDiskSize)
		}
		resourceMap[zone.ZoneName] = zone.ResourceUnits
	}
	return resourceMap
}

func GenerateStatusResourceUnitMap(status v1.TenantStatus) map[string]v1.ResourceUnit {
	var resourceMap = make(map[string]v1.ResourceUnit, 0)
	for _, zone := range status.Topology {
		resourceMap[zone.ZoneName] = zone.ResourceUnits
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

func GenerateSpecLocalityMap(zones []v1.TenantReplica) map[string]v1.TypeSpec {
	localityMap := make(map[string]v1.TypeSpec, 0)
	for _, zone := range zones {
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

func (ctrl *TenantCtrl) GenerateZoneListString() string {
	var zoneList string
	specZoneList := ctrl.GenerateSpecZoneList(ctrl.Tenant.Spec.Topology)
	statusZoneList := ctrl.GenerateStatusZoneList(ctrl.Tenant.Status.Topology)
	if !reflect.DeepEqual(specZoneList, statusZoneList) {
		zoneList = fmt.Sprintf(", ZONE_LIST = ('%s')", strings.Join(specZoneList, "','"))
	}
	return zoneList
}

func (ctrl *TenantCtrl) GeneratePrimaryZoneString() string {
	var primaryZone string
	specPrimaryZone := ctrl.GenerateSpecPrimaryZone(ctrl.Tenant.Spec.Topology)
	statusPrimaryZone := ctrl.GenerateStatusPrimaryZone(ctrl.Tenant.Status.Topology)
	if specPrimaryZone != statusPrimaryZone {
		primaryZone = fmt.Sprintf(", PRIMARY_ZONE = '%s'", specPrimaryZone)
	}
	return primaryZone
}

func (ctrl *TenantCtrl) isUnitEqual(specResourceUnit v1.ResourceUnit, statusResourceUnit v1.ResourceUnit) bool {
	if specResourceUnit.MaxCPU.Equal(statusResourceUnit.MaxCPU) &&
		specResourceUnit.MinCPU.Equal(statusResourceUnit.MinCPU) &&
		specResourceUnit.MaxMemory.Value() == statusResourceUnit.MaxMemory.Value() &&
		specResourceUnit.MinMemory.Value() == statusResourceUnit.MinMemory.Value() &&
		specResourceUnit.MaxIops == statusResourceUnit.MaxIops &&
		specResourceUnit.MinIops == statusResourceUnit.MinIops &&
		specResourceUnit.MaxDiskSize.Value() == statusResourceUnit.MaxDiskSize.Value() &&
		specResourceUnit.MaxSessionNum == statusResourceUnit.MaxSessionNum {
		return true
	} else {
		return false
	}
}

func (ctrl *TenantCtrl) FormatUnitConfig(unit v1.ResourceUnit) string {
	return fmt.Sprintf("MaxCPU: %s MinCPU:%s MaxMemory:%s MinMemory:%s MaxIops:%d MinIops:%d MaxDiskSize:%s MaxSessionNum:%d",
		unit.MaxCPU.String(), unit.MinCPU.String(), unit.MaxMemory.String(), unit.MinMemory.String(), unit.MaxIops, unit.MinIops, unit.MaxDiskSize.String(), unit.MaxSessionNum)
}
