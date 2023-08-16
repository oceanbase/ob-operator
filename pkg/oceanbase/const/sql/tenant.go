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
	GetTenantByName       = "SELECT tenant_id, tenant_name, zone_list, primary_zone, locality, status FROM oceanbase.__all_tenant WHERE tenant_name = ?"
	GetPoolByName         = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.__all_resource_pool WHERE name = ?"
	GetPoolList           = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.__all_resource_pool"
	GetUnitList           = "SELECT unit_id, resource_pool_id, zone, svr_ip, svr_port, migrate_from_svr_ip, migrate_from_svr_port, status FROM oceanbase.__all_unit"
	GetUnitConfigV4List   = "SELECT unit_config_id, name, max_cpu, min_cpu, memory_size, max_iops, min_iops, log_disk_size, iops_weight FROM oceanbase.__all_unit_config"
	GetUnitConfigV4ByName = "SELECT max_cpu, min_cpu, memory_size, log_disk_size, max_iops, min_iops, iops_weight FROM oceanbase.__all_unit_config WHERE name = ?;"

	GetResourceTotal = "SELECT cpu_total, mem_total, disk_total FROM oceanbase.__all_virtual_server_stat where zone = ?"
	GetCharset       = "SELECT CHARSET('oceanbase') as charset"
	GetVariableLike  = "SHOW VARIABLES LIKE ?"
	GetRsJob         = "select job_id, job_type, job_status, tenant_id, tenant_name from __all_rootservice_job where tenant_name=? and job_status ='INPROGRESS' and job_type='ALTER_TENANT_LOCALITY'"
	GetObVersion     = "SELECT ob_version() as version;"

	AddUnitV4 = "CREATE RESOURCE UNIT ? max_cpu ?, memory_size ??;"
	AddPool   = "CREATE RESOURCE POOL ? UNIT=?, UNIT_NUM=?, ZONE_LIST=(?);"
	AddTenant = "CREATE TENANT IF NOT EXISTS ? CHARSET=?, ZONE_LIST=(?), PRIMARY_ZONE=?, RESOURCE_POOL_LIST=(?) ?? ? "

	SetTenantVariable                 = "ALTER TENANT ? ?"
	SetUnitConfigV4_MaxCpu_MemorySize = "ALTER RESOURCE UNIT ? max_cpu ?, memory_size ??;"
	SetPoolUnitNum                    = "ALTER RESOURCE POOL ? UNIT_NUM = ? "
	SetTenantLocality                 = "ALTER TENANT ? LOCALITY = ?"
	SetTenantSQLTemplate              = "ALTER TENANT ? ? ? ? ? ?"

	DeleteUnit   = "DROP RESOURCE UNIT ?"
	DeletePool   = "DROP RESOURCE POOL ?"
	DeleteTenant = "DROP TENANT ? FORCE"
)
