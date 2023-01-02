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
	"strings"

	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func (ctrl *TenantCtrl) CheckAndSetVariables() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Variables For Tenant ", ctrl.Tenant.Name))
	}
	tenantList := sqlOperator.GetGvTenantList()
	var tenantID int
	for _, tenant := range tenantList {
		if tenant.TenantName == ctrl.Tenant.Name {
			tenantID = int(tenant.TenantID)
		}
	}
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
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Unit Config For Tenant ", ctrl.Tenant.Name))
	}
	specResourceUnit := GenerateSpecResourceUnitMap(ctrl.Tenant.Spec)
	statusResourceUnit := GenerateStatusResourceUnitMap(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		match := true
		if !ctrl.isUnitEqual(specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName]) {
			klog.Infof("found zone %s unit config with value %s did't match with config %s", zone.ZoneName, specResourceUnit[zone.ZoneName], statusResourceUnit[zone.ZoneName])
			match = false
		}
		if !match {
			klog.Infof("set zone %s unit config %s", zone.ZoneName, specResourceUnit[zone.ZoneName])
			unitName := ctrl.GenerateUnitName(ctrl.Tenant.Name, zone.ZoneName)
			err = sqlOperator.SetUnitConfig(unitName, specResourceUnit[zone.ZoneName])
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

func (ctrl *TenantCtrl) CheckAndSetResourcePool() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Resource Pool For Tenant ", ctrl.Tenant.Name))
	}
	specUnitNumMap := GenerateSpecUnitNumMap(ctrl.Tenant.Spec)
	statusUnitNumMap := GenerateStatusUnitNumMap(ctrl.Tenant.Status)
	for _, zone := range ctrl.Tenant.Spec.Topology {
		if specUnitNumMap[zone.ZoneName] != statusUnitNumMap[zone.ZoneName] {
			klog.Infof("found zone %s resource pool with unit_num value %s did't match with config %s", zone.ZoneName, statusUnitNumMap[zone.ZoneName], statusUnitNumMap[zone.ZoneName])
			klog.Infof("set zone %s resource pool unit_num %s", zone.ZoneName, specUnitNumMap[zone.ZoneName])
			poolName := ctrl.GeneratePoolName(ctrl.Tenant.Name, zone.ZoneName)
			err = sqlOperator.SetPoolUnitNum(poolName, specUnitNumMap[zone.ZoneName])
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

func (ctrl *TenantCtrl) CheckAndSetTenant() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, fmt.Sprint("Get Sql Operator When Checking And Setting Tenant ", ctrl.Tenant.Name))
	}
	if len(ctrl.Tenant.Spec.Topology) > len(ctrl.Tenant.Status.Topology) {
		// RESOURCE_POOL_LIST
		poolList := ctrl.GeneratePoolList()
		err = sqlOperator.SetTenant(ctrl.Tenant.Name, "", "", poolList, "", "", "")
		if err != nil {
			return err
		}
	}
	// LOCALITY
	specLocalityMap := GenerateSpecLocalityMap(ctrl.Tenant.Spec)
	statusLocalityMap := GenerateStatusLocalityMap(ctrl.Tenant.Status)
	if !reflect.DeepEqual(specLocalityMap, statusLocalityMap) {
		for zoneName, zoneType := range specLocalityMap {
			if statusLocalityMap[zoneName] == "" || statusLocalityMap[zoneName] != specLocalityMap[zoneName] {
				statusLocalityMap[zoneName] = zoneType
				locality := GenerateLocality(statusLocalityMap)
				err = sqlOperator.SetTenantLocality(ctrl.Tenant.Name, locality)
				if err != nil {
					return err
				}
				return ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			}
		}
		for zoneName := range statusLocalityMap {
			if specLocalityMap[zoneName] == "" {
				statusLocalityMap[zoneName] = ""
				locality := GenerateLocality(statusLocalityMap)
				err = sqlOperator.SetTenantLocality(ctrl.Tenant.Name, locality)
				if err != nil {
					return err
				}
				return ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
			}
		}
	}
	// RESOURCE_POOL_LIST
	poolList := ctrl.GeneratePoolList()
	// PRIMARY_ZONE
	primaryZone := ctrl.GeneratePrimaryZone()
	// ZONE_LIST
	zoneList := ctrl.GenerateZoneList()
	// CHARSET
	charset := ctrl.Tenant.Spec.Charset
	if charset != "" {
		charset = fmt.Sprintf(", CHARSET = %s", charset)
	}
	// DEFAULT TABLEGROUP
	defaultTablegroup := ctrl.Tenant.Spec.DefaultTablegroup
	if defaultTablegroup != "" {
		defaultTablegroup = fmt.Sprintf(", DEFAULT TABLEGROUP = {%s}", defaultTablegroup)
	}
	// LOGONLY_REPLICA_NUM
	var logonlyReplicaNum string
	if ctrl.Tenant.Spec.LogonlyReplicaNum != ctrl.Tenant.Status.LogonlyReplicaNum {
		logonlyReplicaNum = fmt.Sprintf(", LOGONLY_REPLICA_NUM = %d", ctrl.Tenant.Spec.LogonlyReplicaNum)
	}
	if zoneList != "" || primaryZone != "" || poolList != "" || charset != "" || defaultTablegroup != "" || logonlyReplicaNum != "" {
		err = sqlOperator.SetTenant(ctrl.Tenant.Name, zoneList, primaryZone, poolList, charset, defaultTablegroup, logonlyReplicaNum)
		if err != nil {
			return err
		}
		err = ctrl.UpdateTenantStatus(tenantconst.TenantModifying)
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

func GenerateSpecLocalityMap(spec v1.TenantSpec) map[string]string {
	localityMap := make(map[string]string, 0)
	for _, zone := range spec.Topology {
		if zone.Type != "" {
			switch strings.ToUpper(string(zone.Type[0])) {
			case tenantconst.TypeF:
				localityMap[zone.ZoneName] = tenantconst.TypeFull
			case tenantconst.TypeL:
				localityMap[zone.ZoneName] = tenantconst.TypeLog
			case tenantconst.TypeR:
				if strings.Contains(zone.Type, "{") {
					localityMap[zone.ZoneName] = tenantconst.TypeR + zone.Type[strings.Index(zone.Type, "{"):]
				} else {
					localityMap[zone.ZoneName] = zone.Type + "{1}"
				}
			}
		}
	}
	return localityMap
}

func GenerateStatusLocalityMap(status v1.TenantStatus) map[string]string {
	localityMap := make(map[string]string, 0)
	for _, zone := range status.Topology {
		if zone.Type != "" {
			tmp := strings.Split(zone.Type, "{")[0]
			switch tmp {
			case tenantconst.TypeFull:
				localityMap[zone.ZoneName] = tenantconst.TypeFull
			case tenantconst.TypeLog:
				localityMap[zone.ZoneName] = tenantconst.TypeLog
			case tenantconst.TypeReadonly:
				localityMap[zone.ZoneName] = zone.Type
			}
		}
	}
	return localityMap
}

func GenerateLocality(localityMap map[string]string) string {
	var locality string
	for zoneName, zoneType := range localityMap {
		if zoneType == "" {
			continue
		}
		switch strings.ToUpper(string(zoneType[0])) {
		case tenantconst.TypeF:
			zoneType = tenantconst.TypeF
		case tenantconst.TypeL:
			zoneType = tenantconst.TypeL
		case tenantconst.TypeR:
			if strings.Contains(zoneType, "{") {
				zoneType = tenantconst.TypeR + zoneType[strings.Index(zoneType, "{"):]
			} else {
				zoneType = tenantconst.TypeR
			}
		}
		locality = fmt.Sprint(locality, zoneType, "@", zoneName, ",")
	}
	locality = locality[0 : len(locality)-1]
	return locality
}

func (ctrl *TenantCtrl) GenerateZoneList() string {
	var zoneList string
	specZoneList := ctrl.GenerateSpecZoneList(ctrl.Tenant.Spec.Topology)
	statusZoneList := ctrl.GenerateStatusZoneList(ctrl.Tenant.Status.Topology)
	if specZoneList != statusZoneList {
		zoneList = fmt.Sprintf(", ZONE_LIST = ('%s')", specZoneList)
	}
	return zoneList
}

func (ctrl *TenantCtrl) GeneratePrimaryZone() string {
	var primaryZone string
	specPrimaryZone := ctrl.GenerateSpecPrimaryZone(ctrl.Tenant.Spec.Topology)
	statusPrimaryZone := ctrl.GenerateStatusPrimaryZone(ctrl.Tenant.Status.Topology)
	if specPrimaryZone != statusPrimaryZone {
		primaryZone = fmt.Sprintf(", PRIMARY_ZONE = '%s'", specPrimaryZone)
	}
	return primaryZone
}

func (ctrl *TenantCtrl) GeneratePoolList() string {
	var poolList string
	specPoolList := ctrl.GenerateSpecPoolList(ctrl.Tenant.Name, ctrl.Tenant.Spec.Topology)
	statusPoolList := ctrl.GenerateStatusPoolList(ctrl.Tenant.Name, ctrl.Tenant.Status.Topology)
	if specPoolList != statusPoolList {
		poolList = fmt.Sprintf(", RESOURCE_POOL_LIST = (%s)", specPoolList)
	}
	return poolList
}
