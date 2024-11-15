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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) ListTenantWithName(ctx context.Context, tenantName string) ([]*model.OBTenant, error) {
	tenants := make([]*model.OBTenant, 0)
	err := m.QueryList(ctx, &tenants, sql.QueryTenantWithName, tenantName)
	if err != nil {
		m.Logger.Error(err, "Failed to query tenants")
		return nil, errors.Wrap(err, "Query tenants")
	}
	return tenants, nil
}

func (m *OceanbaseOperationManager) SelectSysTenant(ctx context.Context) (*model.OBTenant, error) {
	tenants := make([]*model.OBTenant, 0)
	err := m.QueryList(ctx, &tenants, sql.SelectSysTenant)
	if err != nil {
		return nil, errors.Wrap(err, "Select sys tenant")
	}
	if len(tenants) == 0 {
		return nil, errors.New("Empty results when selecting sys tenant")
	}
	return tenants[0], nil
}

func (m *OceanbaseOperationManager) ListUnitsWithTenantId(ctx context.Context, tenantID int64) ([]*model.OBUnit, error) {
	units := make([]*model.OBUnit, 0)
	err := m.QueryList(ctx, &units, sql.QueryUnitsWithTenantId, tenantID)
	if err != nil {
		m.Logger.Error(err, "Failed to query units")
		return nil, errors.Wrap(err, "Query units")
	}
	return units, nil
}

func (m *OceanbaseOperationManager) GetTenantByName(ctx context.Context, tenantName string) (*model.OBTenant, error) {
	tenant := &model.OBTenant{}
	err := m.QueryRow(ctx, tenant, sql.GetTenantByName, tenantName)
	if err != nil {
		return tenant, errors.Wrap(err, "Get tenantconst by tenantName")
	}
	return tenant, nil
}

func (m *OceanbaseOperationManager) GetPoolByName(ctx context.Context, poolName string) (*model.Pool, error) {
	pool := &model.Pool{}
	err := m.QueryRow(ctx, pool, sql.GetPoolByName, poolName)
	if err != nil {
		return pool, errors.Wrap(err, "Get pool by poolName")
	}
	return pool, nil
}

func (m *OceanbaseOperationManager) GetPoolList(ctx context.Context) ([]model.Pool, error) {
	var poolList []model.Pool
	err := m.QueryList(ctx, &poolList, sql.GetPoolList)
	if err != nil {
		return poolList, errors.Wrap(err, "Get get pool list")
	}
	return poolList, nil
}

func (m *OceanbaseOperationManager) GetResourceTotal(ctx context.Context, zoneName string) (*model.ResourceTotal, error) {
	resource := &model.ResourceTotal{}
	err := m.QueryRow(ctx, resource, sql.GetResourceTotal, zoneName)
	if err != nil {
		return resource, errors.Wrap(err, "Get resource by zoneName")
	}
	return resource, nil
}

func (m *OceanbaseOperationManager) GetUnitList(ctx context.Context) ([]model.Unit, error) {
	var unitList []model.Unit
	err := m.QueryList(ctx, &unitList, sql.GetUnitList)
	if err != nil {
		return unitList, errors.Wrap(err, "Get all unit list")
	}
	return unitList, nil
}

func (m *OceanbaseOperationManager) GetUnitConfigV4List(ctx context.Context) ([]model.UnitConfigV4, error) {
	var unitConfigV4List []model.UnitConfigV4
	err := m.QueryList(ctx, &unitConfigV4List, sql.GetUnitConfigV4List)
	if err != nil {
		return unitConfigV4List, errors.Wrap(err, "Get all unitConfigV4 list")
	}
	return unitConfigV4List, nil
}

func (m *OceanbaseOperationManager) GetUnitConfigV4ByName(ctx context.Context, name string) (*model.UnitConfigV4, error) {
	pool := &model.UnitConfigV4{}
	err := m.QueryRow(ctx, pool, sql.GetUnitConfigV4ByName, name)
	if err != nil {
		return pool, errors.Wrap(err, "Get unitConfigV4 By Name")
	}
	return pool, nil
}

func (m *OceanbaseOperationManager) GetCharset(ctx context.Context) (*model.Charset, error) {
	charset := &model.Charset{}
	err := m.QueryRow(ctx, charset, sql.GetCharset)
	if err != nil {
		return charset, errors.Wrap(err, "Get charset")
	}
	return charset, nil
}

func (m *OceanbaseOperationManager) GetRsJob(ctx context.Context, reJobName string) (*model.RsJob, error) {
	rsJob := &model.RsJob{}
	err := m.QueryRow(ctx, rsJob, sql.GetRsJob, reJobName)
	if err != nil {
		return rsJob, errors.Wrap(err, "Get rsJob by reJobName")
	}
	return rsJob, nil
}

// ------------ delete ------------

func (m *OceanbaseOperationManager) DeleteTenant(ctx context.Context, tenantName string, force bool) error {
	preparedSQL, params := m.preparedSQLForDeleteTenant(tenantName, force)
	err := m.ExecWithTimeout(ctx, config.TenantSqlTimeout, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Delete tenantconst by tenantName")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeletePool(ctx context.Context, poolName string) error {
	preparedSQL, params := m.preparedSQLForDeletePool(poolName)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Delete pool by poolName")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteUnitConfig(ctx context.Context, unitName string) error {
	preparedSQL, params := m.preparedSQLForDeleteUnitConfig(unitName)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Delete unit by unitName")
	}
	return nil
}

// ------------ check function ------------

func (m *OceanbaseOperationManager) CheckTenantExistByName(ctx context.Context, tenantName string) (bool, error) {
	var count int
	err := m.QueryCount(ctx, &count, sql.GetTenantCountByName, tenantName)
	if err != nil {
		return false, errors.Wrap(err, "Get tenantconst by tenantName")
	}
	return count != 0, nil
}

func (m *OceanbaseOperationManager) CheckPoolExistByName(ctx context.Context, poolName string) (bool, error) {
	var count int
	err := m.QueryCount(ctx, &count, sql.GetPoolCountByName, poolName)
	if err != nil {
		return false, errors.Wrap(err, "Check whether pool exist by poolName")
	}
	return count != 0, nil
}

func (m *OceanbaseOperationManager) CheckUnitConfigExistByName(ctx context.Context, unitConfigName string) (bool, error) {
	var count int
	err := m.QueryCount(ctx, &count, sql.GetUnitConfigV4CountByName, unitConfigName)
	if err != nil {
		return false, errors.Wrap(err, "Check whether unitconfigV4 exist by poolName")
	}
	return count != 0, nil
}

func (m *OceanbaseOperationManager) CheckRsJobExistByTenantID(ctx context.Context, tenantName int) (bool, error) {
	var count int
	err := m.QueryCount(ctx, &count, sql.GetRsJobCount, tenantName)
	if err != nil {
		return false, errors.Wrap(err, "Check whether rsJob exist by poolName")
	}
	return count != 0, nil
}

// ------------ add ------------

func (m *OceanbaseOperationManager) AddTenant(ctx context.Context, tenantSQLParam model.TenantSQLParam) error {
	preparedSQL, params := preparedSQLForAddTenant(tenantSQLParam)
	err := m.ExecWithTimeout(ctx, config.TenantSqlTimeout, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Add Tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) AddPool(ctx context.Context, pool model.PoolSQLParam) error {
	preparedSQL, params := preparedSQLForAddPool(pool)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Add pool")
	}
	return nil
}

func (m *OceanbaseOperationManager) AddUnitConfigV4(ctx context.Context, unitConfigV4 *model.UnitConfigV4SQLParam) error {
	preparedSQL, params := preparedSQLForAddUnitConfigV4(unitConfigV4)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Add UnitConfigV4")
	}
	return nil
}

// ------------ modify ------------

func (m *OceanbaseOperationManager) SetTenantVariable(ctx context.Context, tenantName, variableList string) error {
	preparedSQL, params := m.preparedSQLForSetTenantVariable(tenantName, variableList)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Set Tenant Variable")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetUnitConfigV4(ctx context.Context, unitConfigV4 *model.UnitConfigV4SQLParam) error {
	preparedSQL, params := preparedSQLForSetUnitConfigV4(unitConfigV4)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Set UnitConfig")
	}
	return nil
}

func (m *OceanbaseOperationManager) SetTenantUnitNum(ctx context.Context, tenantName string, unitNum int) error {
	preparedSQL, params := m.preparedSQLForSetTenantUnitNum(tenantName, unitNum)
	err := m.ExecWithDefaultTimeout(ctx, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Set pool UnitNum")
	}
	return nil
}

func (m *OceanbaseOperationManager) WaitTenantLocalityChangeFinished(ctx context.Context, name string, timeoutSeconds int) error {
	finished := false
	for i := 0; i < timeoutSeconds; i++ {
		tenant, err := m.GetTenantByName(ctx, name)
		if err != nil {
			m.Logger.Error(err, "Failed to get tenant info")
		}
		if tenant.PreviousLocality == "" {
			m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Tenant locality change finished", "tenant name", name)
			finished = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !finished {
		return errors.Errorf("Tenant %s locality change still not finished after %d seconds", name, timeoutSeconds)
	}
	return nil
}

func (m *OceanbaseOperationManager) SetTenant(ctx context.Context, tenantSQLParam model.TenantSQLParam) error {
	preparedSQL, params := preparedSQLForSetTenant(tenantSQLParam)
	m.Logger.V(oceanbaseconst.LogLevelTrace).Info(fmt.Sprintf("sql: %s, parms: %v", preparedSQL, params))
	err := m.ExecWithTimeout(ctx, config.TenantSqlTimeout, preparedSQL, params...)
	if err != nil {
		return errors.Wrap(err, "Set tenant")
	}
	return nil
}

// ---------- replacer sql and collect params ----------

func preparedSQLForAddUnitConfigV4(unitConfigV4 *model.UnitConfigV4SQLParam) (string, []any) {
	var optionSql string
	params := make([]any, 0)
	params = append(params, unitConfigV4.MaxCPU, unitConfigV4.MemorySize)
	if unitConfigV4.MinCPU != 0 {
		optionSql = fmt.Sprint(optionSql, ", min_cpu ?")
		params = append(params, unitConfigV4.MinCPU)
	}
	if unitConfigV4.LogDiskSize != 0 {
		optionSql = fmt.Sprint(optionSql, ", log_disk_size ?")
		params = append(params, unitConfigV4.LogDiskSize)
	}
	if unitConfigV4.MaxIops != 0 {
		optionSql = fmt.Sprint(optionSql, ", max_iops ?")
		params = append(params, unitConfigV4.MaxIops)
	}
	if unitConfigV4.MinIops != 0 {
		optionSql = fmt.Sprint(optionSql, ", min_iops ?")
		params = append(params, unitConfigV4.MinIops)
	}
	if unitConfigV4.IopsWeight != 0 {
		optionSql = fmt.Sprint(optionSql, ", iops_weight ?")
		params = append(params, unitConfigV4.IopsWeight)
	}
	return fmt.Sprintf(sql.AddUnitConfigV4, unitConfigV4.UnitConfigName, optionSql), params
}

func preparedSQLForAddPool(poolSQLParam model.PoolSQLParam) (string, []any) {
	params := make([]any, 0)
	params = append(params, poolSQLParam.UnitName, poolSQLParam.UnitNum, poolSQLParam.ZoneList)
	return fmt.Sprintf(sql.AddPool, poolSQLParam.PoolName), params
}

func preparedSQLForAddTenant(tenantSQLParam model.TenantSQLParam) (string, []any) {
	var option string
	var variableList string
	params := make([]any, 0)
	params = append(params, tenantSQLParam.Charset, tenantSQLParam.PrimaryZone)

	symbols := make([]string, 0)
	for i := 0; i < len(tenantSQLParam.PoolList); i++ {
		symbols = append(symbols, "?")
		params = append(params, tenantSQLParam.PoolList[i])
	}
	if tenantSQLParam.Locality != "" {
		option = fmt.Sprint(option, ", LOCALITY= ?")
		params = append(params, tenantSQLParam.Locality)
	}
	if tenantSQLParam.Collate != "" {
		option = fmt.Sprint(option, ", COLLATE = ?")
		params = append(params, tenantSQLParam.Collate)
	}
	variableList = fmt.Sprintf("SET VARIABLES %s", tenantSQLParam.VariableList)
	return fmt.Sprintf(sql.AddTenant, tenantSQLParam.TenantName, strings.Join(symbols, ", "), option, variableList), params
}

func preparedSQLForSetTenant(tenantSQLParam model.TenantSQLParam) (string, []any) {
	var alterItemStr string
	params := make([]any, 0)
	alterItemList := make([]string, 0)
	if tenantSQLParam.PrimaryZone != "" {
		alterItemList = append(alterItemList, "PRIMARY_ZONE=?")
		params = append(params, tenantSQLParam.PrimaryZone)
	}
	if tenantSQLParam.Charset != "" {
		alterItemList = append(alterItemList, "CHARSET=?")
		params = append(params, tenantSQLParam.Charset)
	}
	if len(tenantSQLParam.PoolList) != 0 {
		symbols := make([]string, 0)
		for i := 0; i < len(tenantSQLParam.PoolList); i++ {
			symbols = append(symbols, "?")
			params = append(params, tenantSQLParam.PoolList[i])
		}
		alterItemList = append(alterItemList, fmt.Sprintf("RESOURCE_POOL_LIST=(%s)", strings.Join(symbols, ", ")))
	}
	if tenantSQLParam.Locality != "" {
		alterItemList = append(alterItemList, "LOCALITY=?")
		params = append(params, tenantSQLParam.Locality)
	}
	alterItemStr = strings.Join(alterItemList, ",")
	return fmt.Sprintf(sql.SetTenant, tenantSQLParam.TenantName, alterItemStr), params
}

func preparedSQLForSetUnitConfigV4(unitConfigV4 *model.UnitConfigV4SQLParam) (string, []any) {
	var alterItemStr string
	params := make([]any, 0)
	alterItemList := make([]string, 0)
	if unitConfigV4.MaxCPU != 0 {
		alterItemList = append(alterItemList, "max_cpu=?")
		params = append(params, unitConfigV4.MaxCPU)
	}
	if unitConfigV4.MemorySize != 0 {
		alterItemList = append(alterItemList, "memory_size=?")
		params = append(params, unitConfigV4.MemorySize)
	}
	if unitConfigV4.MinCPU != 0 {
		alterItemList = append(alterItemList, "min_cpu=?")
		params = append(params, unitConfigV4.MinCPU)
	}
	if unitConfigV4.MinCPU != 0 {
		alterItemList = append(alterItemList, "min_cpu=?")
		params = append(params, unitConfigV4.MinCPU)
	}
	if unitConfigV4.LogDiskSize != 0 {
		alterItemList = append(alterItemList, "log_disk_size=?")
		params = append(params, unitConfigV4.LogDiskSize)
	}
	if unitConfigV4.MaxIops != 0 {
		alterItemList = append(alterItemList, "max_iops=?")
		params = append(params, unitConfigV4.MaxIops)
	}
	if unitConfigV4.MinIops != 0 {
		alterItemList = append(alterItemList, "min_iops=?")
		params = append(params, unitConfigV4.MinIops)
	}
	if unitConfigV4.IopsWeight != 0 {
		alterItemList = append(alterItemList, "iops_weight=?")
		params = append(params, unitConfigV4.IopsWeight)
	}

	alterItemStr = strings.Join(alterItemList, ",")
	return fmt.Sprintf(sql.SetUnitConfigV4, unitConfigV4.UnitConfigName, alterItemStr), params
}

func prepareSQLForAlterPool(param *model.PoolParam) (string, []any) {
	poolProperties := make([]string, 0)
	args := make([]any, 0)
	if len(param.ZoneList) > 0 {
		poolProperties = append(poolProperties, fmt.Sprintf("zone_list = ('%s')", strings.Join(param.ZoneList, "', '")))
	}
	if len(poolProperties) > 0 {
		return fmt.Sprintf(sql.SetPool, param.PoolName, strings.Join(poolProperties, ",")), args
	}
	return "", args
}

func (m *OceanbaseOperationManager) AlterPool(ctx context.Context, poolParam *model.PoolParam) error {
	sql, args := prepareSQLForAlterPool(poolParam)
	if sql != "" {
		return m.ExecWithDefaultTimeout(ctx, sql, args...)
	}
	return nil
}

func (m *OceanbaseOperationManager) preparedSQLForSetTenantVariable(tenantName, variableList string) (string, []any) {
	params := make([]any, 0)
	return fmt.Sprintf(sql.SetTenantVariable, tenantName, variableList), params
}

func (m *OceanbaseOperationManager) preparedSQLForSetTenantUnitNum(tenantNum string, unitNum int) (string, []any) {
	params := make([]any, 0)
	params = append(params, unitNum)
	return fmt.Sprintf(sql.SetTenantUnitNum, tenantNum), params
}

func (m *OceanbaseOperationManager) preparedSQLForDeleteTenant(tenantName string, force bool) (string, []any) {
	params := make([]any, 0)
	if force {
		return fmt.Sprintf(sql.DeleteTenant, tenantName, "force"), params
	}
	return fmt.Sprintf(sql.DeleteTenant, tenantName, ""), params
}

func (m *OceanbaseOperationManager) preparedSQLForDeletePool(poolName string) (string, []any) {
	params := make([]any, 0)
	return fmt.Sprintf(sql.DeletePool, poolName), params
}

func (m *OceanbaseOperationManager) preparedSQLForDeleteUnitConfig(unitConfigName string) (string, []any) {
	params := make([]any, 0)
	return fmt.Sprintf(sql.DeleteUnitConfig, unitConfigName), params
}

func (m *OceanbaseOperationManager) ChangeTenantUserPassword(ctx context.Context, username, password string) error {
	err := m.ExecWithDefaultTimeout(ctx, fmt.Sprintf(sql.ChangeTenantUserPassword, username), password)
	if err != nil {
		return errors.Wrap(err, "Change tenant user password")
	}
	return nil
}

func (m *OceanbaseOperationManager) ListTenantAccessPoints(ctx context.Context, tenantName string) ([]*model.TenantAccessPoint, error) {
	aps := make([]*model.TenantAccessPoint, 0)
	err := m.QueryList(ctx, &aps, sql.QueryTenantAccessPointByName, tenantName)
	if err != nil {
		m.Logger.Error(err, "Failed to list tenant access points")
		return nil, errors.Wrap(err, "List tenant access points")
	}
	return aps, nil
}

func (m *OceanbaseOperationManager) CreateEmptyStandbyTenant(ctx context.Context, params *model.CreateEmptyStandbyTenantParam) error {
	sqlStatement := fmt.Sprintf(sql.CreateEmptyStandbyTenant, params.TenantName, "'"+strings.Join(params.PoolList, "','")+"'")
	err := m.ExecWithTimeout(ctx, config.TenantSqlTimeout, sqlStatement, params.RestoreSource, params.PrimaryZone, params.Locality)
	if err != nil {
		m.Logger.Error(err, "Failed to create empty standby tenant")
		return errors.Wrap(err, "Create empty standby tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) SwitchTenantRole(ctx context.Context, tenant, role string) error {
	if role != "PRIMARY" && role != "STANDBY" {
		return errors.New("invalid tenant role")
	}
	err := m.ExecWithDefaultTimeout(ctx, fmt.Sprintf(sql.SwitchTenantRole, role, tenant))
	if err != nil {
		m.Logger.Error(err, "Failed to switch tenant's role")
		return err
	}
	return nil
}

func (m *OceanbaseOperationManager) ListLSDeletion(ctx context.Context, tenantId int64) ([]*model.LSInfo, error) {
	lsDeletions := make([]*model.LSInfo, 0)
	err := m.QueryList(ctx, &lsDeletions, sql.QueryLSDeletion, tenantId, tenantId)
	if err != nil {
		m.Logger.Error(err, "Failed to list ls deletion")
		return nil, errors.Wrap(err, "List ls deletion")
	}
	return lsDeletions, nil
}

func (m *OceanbaseOperationManager) ListLogStats(ctx context.Context, tenantId int64) ([]*model.LogStat, error) {
	logStats := make([]*model.LogStat, 0)
	err := m.QueryList(ctx, &logStats, sql.QueryLogStats, tenantId)
	if err != nil {
		m.Logger.Error(err, "Failed to list log stats")
		return nil, errors.Wrap(err, "List log stats")
	}
	return logStats, nil
}

func (m *OceanbaseOperationManager) UpgradeTenantWithName(ctx context.Context, tenantName string) error {
	err := m.ExecWithDefaultTimeout(ctx, fmt.Sprintf(sql.UpgradeTenantWithName, tenantName))
	if err != nil {
		return errors.Wrap(err, "Upgrade tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) ListParametersWithTenantID(ctx context.Context, tenantID int64) ([]*model.Parameter, error) {
	params := make([]*model.Parameter, 0)
	err := m.QueryList(ctx, &params, sql.ListParametersWithTenantID, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "List parameters")
	}
	return params, nil
}
