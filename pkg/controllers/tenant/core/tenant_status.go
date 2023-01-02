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
	"context"
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"github.com/pkg/errors"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"strconv"
	"strings"
)

func (ctrl *TenantCtrl) UpdateTenantStatus(tenantStatus string) error {
	tenant := ctrl.Tenant
	tenantExecuter := resource.NewTenantResource(ctrl.Resource)
	tenantTmp, err := tenantExecuter.Get(context.TODO(), tenant.Namespace, tenant.Name)
	if err != nil {
		return err
	}
	tenantCurrent := tenantTmp.(cloudv1.Tenant)
	tenantCurrentDeepCopy := tenantCurrent.DeepCopy()
	ctrl.Tenant = *tenantCurrentDeepCopy
	tenantNew, err := ctrl.BuildTenantStatus(*tenantCurrentDeepCopy, tenantStatus)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(tenantCurrent.Status, tenantNew.Status)
	if !compareStatus {
		err = tenantExecuter.UpdateStatus(context.TODO(), tenantNew)
		if err != nil {
			return err
		}
	}
	ctrl.Tenant = tenantNew
	return nil
}

func (ctrl *TenantCtrl) BuildTenantStatus(tenant cloudv1.Tenant, tenantStatus string) (cloudv1.Tenant, error) {
	var tenantCurrentStatus cloudv1.TenantStatus
	tenantTopology, err := ctrl.BuildTenantTopology(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Status = tenantStatus
	tenantCurrentStatus.Topology = tenantTopology
	tenantCurrentStatus.ReplicaNum, tenantCurrentStatus.LogonlyReplicaNum, err = ctrl.GetReplicaNum(tenant)
	if err != nil {
		return tenant, err
	}
	tenantCurrentStatus.Charset, err = ctrl.GetCharset()
	if err != nil {
		return tenant, err
	}
	tenant.Status = tenantCurrentStatus
	return tenant, nil
}

func (ctrl *TenantCtrl) BuildTenantTopology(tenant cloudv1.Tenant) ([]cloudv1.TenantTopologyStatus, error) {
	var tenantTopologyStatusList []cloudv1.TenantTopologyStatus
	var err error
	var locality string
	var primaryZone string
	var zoneList string
	tenantList, err := ctrl.GetGvTenantList()
	if err != nil {
		return tenantTopologyStatusList, err
	}
	for _, gvTenant := range tenantList {
		if gvTenant.TenantName == tenant.Name {
			locality = gvTenant.Locality
			primaryZone = gvTenant.PrimaryZone
			zoneList = gvTenant.ZoneList
		}
	}

	typeMap := GenerateTypeMap(locality)
	priorityMap := GeneratePriorityMap(primaryZone)
	unitNumMap, err := ctrl.GenerateStatusUnitNumMap(tenant.Spec.Topology)
	if err != nil {
		return tenantTopologyStatusList, err
	}

	for _, zone := range strings.Split(zoneList, ";") {
		var tenantCurrentStatus cloudv1.TenantTopologyStatus
		tenantCurrentStatus.ZoneName = zone
		tenantCurrentStatus.Type = typeMap[zone]
		tenantCurrentStatus.UnitNumber = unitNumMap[zone]
		tenantCurrentStatus.Priority = priorityMap[zone]
		tenantCurrentStatus.ResourceUnits, err = ctrl.BuildResourceUnitFromDB(zone)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantCurrentStatus.UnitConfigs, err = ctrl.BuildUnitFromDB(zone)
		if err != nil {
			return tenantTopologyStatusList, err
		}
		tenantTopologyStatusList = append(tenantTopologyStatusList, tenantCurrentStatus)
	}
	return tenantTopologyStatusList, nil
}

func (ctrl *TenantCtrl) GetCharset() (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return "", errors.Wrap(err, "Get Sql Operator Error When Getting Charset")
	}
	charset := sqlOperator.GetCharset()
	return charset[0].Charset, nil
}

func GenerateTypeMap(locality string) map[string]string {
	typeMap := make(map[string]string, 0)
	typeList := strings.Split(locality, ", ")
	for _, type1 := range typeList {
		tmp := strings.Split(type1, "@")
		typeMap[tmp[1]] = tmp[0]
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

func (ctrl *TenantCtrl) GenerateStatusUnitNumMap(zones []cloudv1.TenantTopology) (map[string]int, error) {
	unitNumMap := make(map[string]int, 0)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return unitNumMap, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	poolList := sqlOperator.GetPoolList()
	for _, zone := range zones {
		poolName := ctrl.GeneratePoolName(ctrl.Tenant.Name, zone.ZoneName)
		for _, pool := range poolList {
			if pool.Name == poolName {
				unitNumMap[zone.ZoneName] = int(pool.UnitCount)
			}
		}
	}
	return unitNumMap, nil
}

func (ctrl *TenantCtrl) BuildResourceUnitFromDB(zone string) (cloudv1.ResourceUnit, error) {
	var resourceUnit cloudv1.ResourceUnit
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return resourceUnit, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	unitList := sqlOperator.GetUnitList()
	poolList := sqlOperator.GetPoolList()
	unitConfigList := sqlOperator.GetUnitConfigList()
	var resourcePoolIDList []int
	for _, unit := range unitList {
		if unit.Zone == zone {
			resourcePoolIDList = append(resourcePoolIDList, int(unit.ResourcePoolID))
		}
	}
	for _, pool := range poolList {
		for _, resourcePoolID := range resourcePoolIDList {
			if resourcePoolID == int(pool.ResourcePoolID) {
				for _, unitConifg := range unitConfigList {
					if unitConifg.UnitConfigID == pool.UnitConfigID {
						resourceUnit.MaxCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MaxCPU, 'f', -1, 64))
						resourceUnit.MinCPU = apiresource.MustParse(strconv.FormatFloat(unitConifg.MinCPU, 'f', -1, 64))
						resourceUnit.MaxMemory = *apiresource.NewQuantity(unitConifg.MaxMemory, apiresource.DecimalSI)
						resourceUnit.MinMemory = *apiresource.NewQuantity(unitConifg.MinMemory, apiresource.DecimalSI)
						resourceUnit.MaxDiskSize = *apiresource.NewQuantity(unitConifg.MaxDiskSize, apiresource.DecimalSI)
						resourceUnit.MaxIops = int(unitConifg.MaxIops)
						resourceUnit.MinIops = int(unitConifg.MinIops)
						resourceUnit.MaxSessionNum = int(unitConifg.MaxSessionNum)
					}
				}
			}
		}
	}
	return resourceUnit, nil
}

func (ctrl *TenantCtrl) BuildUnitFromDB(zone string) ([]cloudv1.Unit, error) {
	var unitList []cloudv1.Unit
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return unitList, errors.Wrap(err, "Get Sql Operator Error When Building Resource Unit From DB")
	}
	units := sqlOperator.GetUnitList()
	for _, unit := range units {
		if unit.Zone == zone {
			var res cloudv1.Unit
			res.UnitId = int(unit.UnitID)
			res.ServerIP = unit.SvrIP
			res.ServerPort = int(unit.SvrPort)
			res.Status = unit.Status
			var migrateServer cloudv1.MigrateServer
			migrateServer.ServerIP = unit.MigrateFromSvrIP
			migrateServer.ServerPort = int(unit.MigrateFromSvrPort)
			res.Migrate = migrateServer
			unitList = append(unitList, res)
		}
	}
	return unitList, nil
}

func (ctrl *TenantCtrl) GetReplicaNum(tenant cloudv1.Tenant) (int, int, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return 0, 0, errors.Wrap(err, "Get Sql Operator Error When Getting Replica Num  From DB")
	}
	tenantList := sqlOperator.GetTenantList()
	for _, t := range tenantList {
		if t.TenantName == tenant.Name {
			return int(t.ReplicaNum), int(t.LogonlyReplicaNum), nil
		}
	}
	return 0, 0, errors.New("No Tenant Found")
}
