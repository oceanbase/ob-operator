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
	"sort"
	"strings"

	v1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/model"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

func (ctrl *TenantCtrl) GetGvTenantList() ([]model.GvTenantList, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
	}
	tenantList := sqlOperator.GetGvTenantList()
	return tenantList, nil
}

func (ctrl *TenantCtrl) GetTenantList() ([]model.TenantList, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "Get Sql Operator Error When Getting Tenant List")
	}
	tenantList := sqlOperator.GetTenantList()
	return tenantList, nil
}

func (ctrl *TenantCtrl) TenantExist(tenantName string) (bool, int, error) {
	klog.Infoln("Check Whether The Tenant Exists")
	tenantList, err := ctrl.GetGvTenantList()
	if err != nil {
		return false, 0, err
	}
	for _, tenant := range tenantList {
		if tenant.TenantName == tenantName {
			return true, int(tenant.TenantID), nil
		}
	}
	return false, 0, nil
}

func (ctrl *TenantCtrl) PoolExist(poolName string) (bool, int, error) {
	klog.Infoln("Check Whether The Resource Pool Is Exist")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return false, 0, errors.Wrap(err, "Get Sql Operator Error When Checking Whether The Resource Pool Exists")
	}
	poolList := sqlOperator.GetPoolList()
	for _, pool := range poolList {
		if pool.Name == poolName {
			return true, int(pool.ResourcePoolID), nil
		}
	}
	return false, 0, nil
}

func (ctrl *TenantCtrl) UnitExist(name string) (error, bool) {
	klog.Infoln("Check Whether The Resource Unit Is Exist")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Checking Whether The Resource Unit Exists"), false
	}
	unitConfigList := sqlOperator.GetUnitConfigList()
	for _, unit := range unitConfigList {
		if unit.Name == name {
			return nil, true
		}
	}
	return nil, false
}

func (ctrl *TenantCtrl) GenerateUnitName(name, zoneName string) string {
	unitName := fmt.Sprintf("%s_unit_%s_1", name, zoneName)
	return unitName
}

func (ctrl *TenantCtrl) GeneratePoolName(name, zoneName string) string {
	poolName := fmt.Sprintf("%s_pool_%s_1", name, zoneName)
	return poolName
}

func (ctrl *TenantCtrl) CheckResourceEnough(zone v1.TenantTopology) error {
	klog.Infoln("Check Reousrce ", zone.ZoneName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Checking Reousrce")
	}
	resource := sqlOperator.GetResource(zone)
	unitList := sqlOperator.GetUnitList()
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone.ZoneName {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	poolList := sqlOperator.GetPoolList()
	unitConfigList := sqlOperator.GetUnitConfigList()
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						resource[0].CPUTotal -= unitConifg.MaxCPU
						resource[0].MemTotal -= unitConifg.MaxMemory
						// resource[0].DiskTotal -= unitConifg.MaxDiskSize
					}
				}
			}
		}
	}
	if zone.ResourceUnits.MaxCPU.AsApproximateFloat64() > resource[0].CPUTotal {
		return errors.New(fmt.Sprint("CPU Is Not Enough: ", resource[0].CPUTotal))
	}
	maxMem := zone.ResourceUnits.MaxMemory.Value()
	if err != nil {
		return err
	}
	if maxMem > resource[0].MemTotal {
		return errors.New(fmt.Sprint("Memory Is Not Enough: ", FormatSize(int(maxMem)), FormatSize(int(resource[0].MemTotal))))
	}
	// err, maxDiskSize := PraseSize(zone.ResourceUnits.MaxDiskSize)
	// if err != nil {
	// 	return err
	// }
	// if maxDiskSize > int(resource[0].DiskTotal) {
	// 	return errors.New(fmt.Sprint("DiskSize Is Not Enough: ", int(resource[0].DiskTotal)))
	// }
	return nil
}

func (ctrl *TenantCtrl) CreateUnit(unitName string, resourceUnit v1.ResourceUnit) error {
	klog.Infoln("Create Resource Unit ", unitName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Unit")
	}
	if resourceUnit.MaxDiskSize.Value() == 0 {
		resourceUnit.MaxDiskSize = resource.MustParse(tenantconst.MaxDiskSize)
	}
	if resourceUnit.MaxIops == 0 {
		resourceUnit.MaxIops = tenantconst.MaxIops
	}
	if resourceUnit.MinIops == 0 {
		resourceUnit.MinIops = tenantconst.MinIops
	}
	if resourceUnit.MaxSessionNum == 0 {
		resourceUnit.MaxSessionNum = tenantconst.MaxSessionNum
	}
	return sqlOperator.CreateUnit(unitName, resourceUnit)
}

func (ctrl *TenantCtrl) CreatePool(poolName, unitName string, zone v1.TenantTopology) error {
	klog.Infoln("Create Resource Pool ", poolName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	return sqlOperator.CreatePool(poolName, unitName, zone)
}

func (ctrl *TenantCtrl) CreateTenant(tenantName string, zones []v1.TenantTopology) error {
	klog.Infoln("Create Resource Tenant ", tenantName)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "Get Sql Operator Error When Creating Resource Pool")
	}
	zoneList := ctrl.GenerateSpecZoneList(zones)
	primaryZone := ctrl.GenerateSpecPrimaryZone(zones)
	poolList := ctrl.GenerateSpecPoolList(tenantName, zones)
	variableList := ctrl.GenerateVariableList(ctrl.Tenant.Spec.Variables)
	charset := tenantconst.Charset
	locality := ctrl.GenerateLocality(zones)
	if locality != "" {
		locality = fmt.Sprintf(", LOCALITY='%s'", locality)
	}
	comment := ctrl.Tenant.Spec.Comment
	if comment != "" {
		comment = fmt.Sprintf(", COMMENT ='%s'", comment)
	}
	defaultTablegroup := ctrl.Tenant.Spec.DefaultTablegroup
	if defaultTablegroup != "" {
		defaultTablegroup = fmt.Sprintf(", DEFAULT TABLEGROUP = {%s}", defaultTablegroup)
	}
	collate := ctrl.Tenant.Spec.Collate
	if collate != "" {
		collate = fmt.Sprintf(", COLLATE = %s", collate)
	}
	var logonlyReplicaNum string
	if ctrl.Tenant.Spec.LogonlyReplicaNum != 0 {
		logonlyReplicaNum = fmt.Sprintf(", LOGONLY_REPLICA_NUM = %d", ctrl.Tenant.Spec.LogonlyReplicaNum)
	}
	if ctrl.Tenant.Spec.Charset != "" {
		charset = ctrl.Tenant.Spec.Charset
	}
	return sqlOperator.CreateTenant(tenantName, charset, zoneList, primaryZone, poolList, locality, comment, defaultTablegroup, collate, logonlyReplicaNum, variableList)
}

func (ctrl *TenantCtrl) GenerateSpecZoneList(zones []v1.TenantTopology) string {
	var zoneList string
	for _, zone := range zones {
		zoneList = fmt.Sprint(zoneList, zone.ZoneName, ",")
	}
	zoneList = zoneList[0 : len(zoneList)-1]
	return zoneList
}

func (ctrl *TenantCtrl) GenerateStatusZoneList(zones []v1.TenantTopologyStatus) string {
	var zoneList string
	for _, zone := range zones {
		zoneList = fmt.Sprint(zoneList, zone.ZoneName, ",")
	}
	zoneList = zoneList[0 : len(zoneList)-1]
	return zoneList
}

func (ctrl *TenantCtrl) GenerateSpecPoolList(tenantName string, zones []v1.TenantTopology) string {
	var poolList string
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(tenantName, zone.ZoneName)
		poolList = fmt.Sprint(poolList, "'", poolName, "',")
	}
	poolList = poolList[0 : len(poolList)-1]
	return poolList
}

func (ctrl *TenantCtrl) GenerateStatusPoolList(tenantName string, zones []v1.TenantTopologyStatus) string {
	var poolList string
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(tenantName, zone.ZoneName)
		poolList = fmt.Sprint(poolList, "'", poolName, "',")
	}
	poolList = poolList[0 : len(poolList)-1]
	return poolList
}

func (ctrl *TenantCtrl) GenerateSpecPrimaryZone(zones []v1.TenantTopology) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, zone := range zones {
		zones := zoneMap[zone.Priority]
		zones = append(zones, zone.ZoneName)
		zoneMap[zone.Priority] = zones

	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		for _, zone := range zones {
			primaryZone = fmt.Sprint(primaryZone, zone, ",")
		}
		primaryZone = primaryZone[0 : len(primaryZone)-1]
		primaryZone = fmt.Sprint(primaryZone, ";")
	}
	primaryZone = primaryZone[0 : len(primaryZone)-1]
	primaryZone = fmt.Sprint(primaryZone, ";")
	return primaryZone
}

func (ctrl *TenantCtrl) GenerateStatusPrimaryZone(zones []v1.TenantTopologyStatus) string {
	var primaryZone string
	zoneMap := make(map[int][]string, 0)
	var priorityList []int
	for _, zone := range zones {
		zones := zoneMap[zone.Priority]
		zones = append(zones, zone.ZoneName)
		zoneMap[zone.Priority] = zones
	}
	for k := range zoneMap {
		priorityList = append(priorityList, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorityList)))
	for _, priority := range priorityList {
		zones := zoneMap[priority]
		for _, zone := range zones {
			primaryZone = fmt.Sprint(primaryZone, zone, ",")
		}
		primaryZone = primaryZone[0 : len(primaryZone)-1]
		primaryZone = fmt.Sprint(primaryZone, ";")
	}
	primaryZone = primaryZone[0 : len(primaryZone)-1]
	primaryZone = fmt.Sprint(primaryZone, ";")
	return primaryZone
}

func (ctrl *TenantCtrl) GenerateLocality(zones []v1.TenantTopology) string {
	var locality string
	for _, zone := range zones {
		if zone.Type != "" {
			zoneType := strings.ToUpper(string(zone.Type[0]))
			switch zoneType {
			case tenantconst.TypeF:
				zoneType = tenantconst.TypeFull
			case tenantconst.TypeL:
				zoneType = tenantconst.TypeLog
			case tenantconst.TypeR:
				if strings.Contains(zone.Type, "{") {
					zoneType = tenantconst.TypeR + zoneType[strings.Index(zoneType, "{"):]
				} else {
					zoneType = tenantconst.TypeR
				}
			}
			locality = fmt.Sprint(locality, zoneType, "@", zone.ZoneName, ",")
		}
	}
	locality = locality[0 : len(locality)-1]
	return locality
}

func (ctrl *TenantCtrl) GenerateVariableList(variables []v1.Parameter) string {
	var variableList string
	if len(variables) == 0 {
		return variableList
	}
	variableList = fmt.Sprint("SET ", variableList)
	for _, variable := range variables {
		variableList = fmt.Sprint(variableList, variable.Name, "='", variable.Value, "',")
	}
	variableList = variableList[0 : len(variableList)-1]
	return variableList
}

func FormatSize(size int) string {
	units := [...]string{"B", "K", "M", "G", "T", "P"}
	idx := 0
	size1 := float64(size)
	for idx < 5 && size1 >= 1024 {
		size1 /= 1024.0
		idx += 1
	}
	res := fmt.Sprintf("%.1f%s", size1, units[idx])
	return res
}
