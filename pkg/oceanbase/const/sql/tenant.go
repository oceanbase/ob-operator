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

package sql

const (
	GetTenantByName       = "SELECT tenant_id, tenant_name, primary_zone, locality, COALESCE(previous_locality, '') as previous_locality, status FROM oceanbase.DBA_OB_TENANTS WHERE tenant_name = ?;"
	GetPoolByName         = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.DBA_OB_RESOURCE_POOLS WHERE name = ?;"
	GetUnitConfigV4ByName = "SELECT max_cpu, min_cpu, memory_size, log_disk_size, max_iops, min_iops, iops_weight FROM oceanbase.DBA_OB_UNIT_CONFIGS WHERE name = ?;"
	GetPoolList           = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.DBA_OB_RESOURCE_POOLS;"
	GetUnitList           = "SELECT unit_id, resource_pool_id, zone, svr_ip, svr_port, migrate_from_svr_ip, migrate_from_svr_port, status FROM oceanbase.DBA_OB_UNITS;"
	GetUnitConfigV4List   = "SELECT unit_config_id, name, max_cpu, min_cpu, memory_size, max_iops, min_iops, log_disk_size, iops_weight FROM oceanbase.DBA_OB_UNIT_CONFIGS;"

	GetTenantCountByName       = "SELECT count(*) FROM oceanbase.DBA_OB_TENANTS WHERE tenant_name = ?;"
	GetPoolCountByName         = "SELECT count(*) FROM oceanbase.DBA_OB_RESOURCE_POOLS WHERE name = ?;"
	GetUnitConfigV4CountByName = "SELECT count(*) FROM oceanbase.DBA_OB_UNIT_CONFIGS WHERE name = ?;"
	GetRsJobCount              = "select count(*) from DBA_OB_TENANT_JOBS where tenant_id=? and job_status ='INPROGRESS' and job_type='ALTER_TENANT_LOCALITY'"

	GetResourceTotal = "SELECT cpu_capacity, mem_capacity, data_disk_capacity FROM oceanbase.GV$OB_SERVERS;"
	GetCharset       = "SELECT CHARSET('oceanbase') as charset;"
	GetVariableLike  = "SHOW VARIABLES LIKE ?;"
	GetRsJob         = "select job_id, job_type, job_status, tenant_id from DBA_OB_TENANT_JOBS where tenant_name=? and job_status ='INPROGRESS' and job_type='ALTER_TENANT_LOCALITY'"
	GetObVersion     = "SELECT ob_version() as version;"

	AddUnitConfigV4 = "CREATE RESOURCE UNIT IF NOT EXISTS %s max_cpu ?, memory_size ? %s;"
	AddPool         = "CREATE RESOURCE POOL IF NOT EXISTS %s UNIT=?, UNIT_NUM=?, ZONE_LIST=(?);"
	AlterPool       = "ALTER RESOURCE POOL %s UNIT=?, UNIT_NUM=?, ZONE_LIST=(?);"
	AddTenant       = "CREATE TENANT IF NOT EXISTS %s CHARSET=?, PRIMARY_ZONE=?, RESOURCE_POOL_LIST=(%s) %s %s;"

	SetTenantVariable = "ALTER TENANT %s VARIABLES %s;"
	SetUnitConfigV4   = "ALTER RESOURCE UNIT %s %s;"
	SetTenantUnitNum  = "ALTER RESOURCE TENANT %s UNIT_NUM = ?;"
	SetTenant         = "ALTER TENANT %s %s;"
	SetTenantName     = "ALTER TENANT %s RENAME GLOBAL_NAME TO ?"

	SetPool = "ALTER RESOURCE POOL %s %s"

	DeleteUnitConfig = "DROP RESOURCE UNIT %s;"
	DeletePool       = "DROP RESOURCE POOL %s;"
	DeleteTenant     = "DROP TENANT %s %s;"
)

const (
	ChangeTenantUserPassword     = "ALTER USER %s IDENTIFIED BY ?;"
	CreateTenantUserWithPwd      = "CREATE USER IF NOT EXISTS %s IDENTIFIED BY ?;"
	GrantOBROPrivilege           = "GRANT SELECT ON oceanbase.* TO %s;"
	QueryTenantAccessPointByName = "SELECT tenant_id, tenant_name, svr_ip, sql_port from oceanbase.CDB_OB_ACCESS_POINT WHERE tenant_name = ?;"
)
