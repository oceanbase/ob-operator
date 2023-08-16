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

package operation

import (
	"fmt"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/tenant"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
)

func (m *OceanbaseOperationManager) GetTenantByName(tenantName string) (*model.Tenant, error) {
	tenant := &model.Tenant{}
	err := m.QueryRow(tenant, sql.GetTenantByName, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when get tenantconst by tenantName")
		return tenant, errors.Wrap(err, "Get tenantconst by tenantName")
	}
	return tenant, nil
}

func (m *OceanbaseOperationManager) GetPoolByName(poolName string) (*model.Pool, error) {
	pool := &model.Pool{}
	err := m.QueryRow(pool, sql.GetPoolByName, poolName)
	if err != nil {
		m.Logger.Error(err, "Got exception when get pool by poolName")
		return pool, errors.Wrap(err, "Get pool by poolName")
	}
	return pool, nil
}

func (m *OceanbaseOperationManager) GetPoolList() ([]model.Pool, error) {
	var poolList []model.Pool
	err := m.QueryList(&poolList, sql.GetPoolList)
	if err != nil {
		m.Logger.Error(err, "Got exception when get pool list")
		return poolList, errors.Wrap(err, "Get get pool list")
	}
	return poolList, nil
}

func (m *OceanbaseOperationManager) GetResourceTotal(zoneName string) (*model.ResourceTotal, error) {
	resource := &model.ResourceTotal{}
	err := m.QueryRow(resource, sql.GetResourceTotal, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when get resource by zoneName")
		return resource, errors.Wrap(err, "Get resource by zoneName")
	}
	return resource, nil
}

func (m *OceanbaseOperationManager) GetUnitList() ([]model.Unit, error) {
	var unitList []model.Unit
	err := m.QueryList(&unitList, sql.GetUnitList)
	if err != nil {
		m.Logger.Error(err, "Got exception when get all unit list")
		return unitList, errors.Wrap(err, "Get all unit list")
	}
	return unitList, nil
}

func (m *OceanbaseOperationManager) GetUnitConfigV4List() ([]model.UnitConfigV4, error) {
	var unitConfigV4List []model.UnitConfigV4
	err := m.QueryList(&unitConfigV4List, sql.GetUnitConfigV4List)
	if err != nil {
		m.Logger.Error(err, "Got exception when get all unitConfigV4 list")
		return unitConfigV4List, errors.Wrap(err, "Get all unitConfigV4 list")
	}
	return unitConfigV4List, nil
}

func (m *OceanbaseOperationManager) GetUnitConfigV4ByName(name string) (*model.UnitConfigV4, error) {
	pool := &model.UnitConfigV4{}
	err := m.QueryRow(pool, sql.GetUnitConfigV4ByName, name)
	if err != nil {
		m.Logger.Error(err, "Got exception when get unitConfigV4 by name")
		return pool, errors.Wrap(err, "Get unitConfigV4 By Name")
	}
	return pool, nil
}

func (m *OceanbaseOperationManager) GetCharset() (*model.Charset, error) {
	charset := &model.Charset{}
	err := m.QueryRow(charset, sql.GetCharset)
	if err != nil {
		m.Logger.Error(err, "Got exception when get charset")
		return charset, errors.Wrap(err, "Get charset")
	}
	return charset, nil
}

func (m *OceanbaseOperationManager) GetVariable(name string) ([]model.Variable, error) {
	var variableList []model.Variable
	err := m.QueryRow(&variableList, sql.GetVariableLike, name)
	if err != nil {
		m.Logger.Error(err, "Got exception when get variableList")
		return variableList, errors.Wrap(err, "Get variableList")
	}
	return variableList, nil
}

func (m *OceanbaseOperationManager) GetRsJob(reJobName string) (*model.RsJob, error) {
	rsJob := &model.RsJob{}
	err := m.QueryRow(rsJob, sql.GetRsJob, reJobName)
	if err != nil {
		m.Logger.Error(err, "Got exception when get rsJob by reJobName")
		return rsJob, errors.Wrap(err, "Get rsJob by reJobName")
	}
	return rsJob, nil
}

func (m *OceanbaseOperationManager) GetVersion() (*model.OBVersion, error) {
	version := &model.OBVersion{}
	err := m.QueryRow(version, sql.GetObVersion)
	if err != nil {
		m.Logger.Error(err, "Got exception when get version")
		return version, errors.Wrap(err, "Get version")
	}
	return version, nil
}

// ------------ delete ------------

func (m *OceanbaseOperationManager) DeleteTenant(tenantName string) error {
	err := m.ExecWithDefaultTimeout(sql.DeleteTenant, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete tenantconst by tenantName")
		return errors.Wrap(err, "Delete tenantconst by tenantName")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeletePool(poolName string) error {
	err := m.ExecWithDefaultTimeout(sql.DeletePool, poolName)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete pool by poolName")
		return errors.Wrap(err, "Delete pool by poolName")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteUnit(unitName string) error {
	err := m.ExecWithDefaultTimeout(sql.DeleteUnit, unitName)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete unit by unitName")
		return errors.Wrap(err, "Delete unit by unitName")
	}
	return nil
}

// ------------ add ------------

func (m *OceanbaseOperationManager) AddTenant(tenantName, charset, zoneList, primaryZone, poolList, locality, collate, variableList string) error {
	var option string
	if charset == "" {
		charset = tenant.Charset
	}
	if locality != "" {
		option = fmt.Sprint(option,  ", LOCALITY='%s' ", locality)
	}
	if collate != "" {
		option = fmt.Sprint(option,  ", COLLATE = %s ", collate)
	}

	err := m.ExecWithDefaultTimeout(sql.AddTenant, tenantName, charset, zoneList, primaryZone, poolList, option, variableList)
	if err != nil {
		m.Logger.Error(err, "Got exception when add Tenant")
		return errors.Wrap(err, "Add Tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) AddPool(poolName, unitName string, pool v1alpha1.ResourcePoolSpec) error {
	err := m.ExecWithDefaultTimeout(sql.AddPool, poolName, unitName, pool.UnitNumber, pool.ZoneList)
	if err != nil {
		m.Logger.Error(err, "Got exception when add pool")
		return errors.Wrap(err, "Add pool")
	}
	return nil
}

func (m *OceanbaseOperationManager) AddUnitConfigV4(unitConfigName string, unitConfigV4 model.UnitConfigV4) error {

	var option string
	if unitConfigV4.MinCPU != 0 {
		option = fmt.Sprint(option, ", min_cpu ", unitConfigV4.MinCPU)
	}
	if unitConfigV4.LogDiskSize != 0 {
		option = fmt.Sprint(option, ", log_disk_size ", unitConfigV4.LogDiskSize)
	}
	if unitConfigV4.MaxIops != 0 {
		option = fmt.Sprint(option, ", max_iops ", unitConfigV4.MaxIops)
	}
	if unitConfigV4.MinIops != 0 {
		option = fmt.Sprint(option, ", min_iops ", unitConfigV4.MinIops)
	}
	if unitConfigV4.IopsWeight != 0 {
		option = fmt.Sprint(option, ", iops_weight ", unitConfigV4.IopsWeight)
	}

	err := m.ExecWithDefaultTimeout(sql.AddUnitV4, unitConfigName, unitConfigV4.MaxCPU, unitConfigV4.MemorySize, option)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

// ------------ modify ------------

func (m *OceanbaseOperationManager) SetTenantVariable(tenantName, variableList string) error {
	err := m.ExecWithDefaultTimeout(sql.SetTenantVariable, tenantName, variableList)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetUnitConfigV4(unitConfigName string,  unitConfigV4 model.UnitConfigV4) error {

	var option string
	if unitConfigV4.MinCPU != 0 {
		option = fmt.Sprint(option, ", min_cpu ", unitConfigV4.MinCPU)
	}
	if unitConfigV4.LogDiskSize != 0 {
		option = fmt.Sprint(option, ", log_disk_size ", unitConfigV4.LogDiskSize)
	}
	if unitConfigV4.MaxIops != 0 {
		option = fmt.Sprint(option, ", max_iops ", unitConfigV4.MaxIops)
	}
	if unitConfigV4.MinIops != 0 {
		option = fmt.Sprint(option, ", min_iops ", unitConfigV4.MinIops)
	}
	if unitConfigV4.IopsWeight != 0 {
		option = fmt.Sprint(option, ", iops_weight ", unitConfigV4.IopsWeight)
	}

	err := m.ExecWithDefaultTimeout(sql.SetUnitConfigV4_MaxCpu_MemorySize, unitConfigName, unitConfigV4.MaxCPU, unitConfigV4.MemorySize, option)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetPoolUnitNum(poolName string, unitNum int) error {
	err := m.ExecWithDefaultTimeout(sql.SetPoolUnitNum, poolName, unitNum)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetTenantLocality(tenantName, locality string) error {
	err := m.ExecWithDefaultTimeout(sql.SetTenantLocality, tenantName, locality)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetTenant(tenantName, zoneList, primaryZone, poolList, charset, locality string) error {
	err := m.ExecWithDefaultTimeout(sql.SetTenantSQLTemplate, tenantName, zoneList, primaryZone, poolList, charset, locality)
	if err != nil {
		m.Logger.Error(err, "Got exception when add UnitV4")
		return errors.Wrap(err, "Add server")
	}
	return nil
}
