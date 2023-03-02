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

package sql

const (
	DatabaseOb = "oceanbase"
)

const (
	GetTenantSQL   = "SELECT tenant_id, tenant_name, zone_list, primary_zone, locality, status FROM oceanbase.__all_tenant WHERE tenant_name = '${NAME}'"
	GetPoolSQL     = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.__all_resource_pool WHERE name = '${NAME}'"
	GetPoolListSQL = "SELECT resource_pool_id, name, unit_count, unit_config_id, zone_list, tenant_id FROM oceanbase.__all_resource_pool"

	GetUnitListSQL         = "SELECT unit_id, resource_pool_id, zone, svr_ip, svr_port, migrate_from_svr_ip, migrate_from_svr_port, status FROM oceanbase.__all_unit"
	GetUnitConfigV3ListSQL = "SELECT unit_config_id, name, max_cpu, min_cpu, max_memory, min_memory, max_iops, min_iops, max_disk_size, max_session_num FROM oceanbase.__all_unit_config"
	GetUnitConfigV4ListSQL = "SELECT unit_config_id, name, max_cpu, min_cpu, memory_size, max_iops, min_iops, log_disk_size, iops_weight FROM oceanbase.__all_unit_config"
	GetUnitConfigSQL       = "SELECT unit_config_id, name FROM oceanbase.__all_unit_config WHERE name = '${NAME}'"
	GetUnitConfigV4SQL     = "SELECT max_cpu, min_cpu, memory_size, log_disk_size, max_iops, min_iops, iops_weight FROM oceanbase.__all_unit_config WHERE name = '${NAME}';"
	GetResourceSQLTemplate = "SELECT cpu_total, mem_total, disk_total FROM oceanbase.__all_virtual_server_stat where zone = '${ZONE_NAME}'"

	GetCharsetSQL           = "SELECT CHARSET('oceanbase')"
	CreateUnitV3SQLTemplate = "CREATE RESOURCE UNIT ${UNIT_NAME} max_cpu ${MAX_CPU}, max_memory '${MAX_MEMORY}', max_iops ${MAX_IOPS}, max_disk_size '${MAX_DISK_SIZE}', max_session_num ${MAX_SESSION_NUM}, MIN_CPU=${MIN_CPU}, MIN_MEMORY='${MIN_MEMORY}', MIN_IOPS=${MIN_IOPS};"
	CreateUnitV4SQLTemplate = "CREATE RESOURCE UNIT ${UNIT_NAME} max_cpu ${MAX_CPU}, memory_size ${MEMORY_SIZE}${OPTION};"
	CreatePoolSQLTemplate   = "CREATE RESOURCE POOL ${POOL_NAME} UNIT=${UNIT_NAME}, UNIT_NUM=${UNIT_NUM}, ZONE_LIST=('${ZONE_NAME}');"
	CreateTenantSQLTemplate = "CREATE TENANT IF NOT EXISTS ${TENANT_NAME} CHARSET='${CHARSET}', ZONE_LIST=('${ZONE_LIST}'), PRIMARY_ZONE='${PRIMARY_ZONE}', RESOURCE_POOL_LIST=('${RESOURCE_POOL_LIST}') ${LOCALITY}${COLLATE} ${VARIABLE_LIST} "

	GetVariableSQLTemplate       = "SELECT tenant_id, zone, name, value FROM __all_virtual_sys_variable WHERE name = '${NAME}' and tenant_id = ${TENANT_ID}"
	SetTenantVariableSQLTemplate = "ALTER TENANT ${TENANT_NAME} SET VARIABLES ${NAME} = ${VALUE}"
	SetUnitConfigV3SQLTemplate   = "ALTER RESOURCE UNIT ${UNIT_NAME} max_cpu ${MAX_CPU}, max_memory '${MAX_MEMORY}', max_iops ${MAX_IOPS}, max_disk_size '${MAX_DISK_SIZE}', max_session_num ${MAX_SESSION_NUM}, MIN_CPU=${MIN_CPU}, MIN_MEMORY='${MIN_MEMORY}', MIN_IOPS=${MIN_IOPS};"
	SetUnitConfigV4SQLTemplate   = "ALTER RESOURCE UNIT ${UNIT_NAME} max_cpu ${MAX_CPU}, memory_size ${MEMORY_SIZE}${OPTION};"

	SetPoolUnitNumSQLTemplate    = "ALTER RESOURCE POOL ${POOL_NAME} UNIT_NUM = ${UNIT_NUM} "
	SetTenantLocalitySQLTemplate = "ALTER TENANT ${TENANT_NAME} LOCALITY = '${LOCALITY}'"
	SetTenantPoolListSQLTemplate = "ALTER TENANT ${TENANT_NAME}  RESOURCE_POOL_LIST = '${POOL_LIST}'"
	// SetTenantSQLTemplate         = "ALTER TENANT ${TENANT_NAME} ${ZONE_LIST} ${PRIMARY_ZONE} ${CHARSET} ${LOGONLY_REPLICA_NUM}"
	SetTenantSQLTemplate = "ALTER TENANT ${TENANT_NAME} ${ZONE_LIST}${PRIMARY_ZONE}${RESOURCE_POOL_LIST}${CHARSET}${LOCALITY}${LOGONLY_REPLICA_NUM}"

	GetInprogressJobSQLTemplate = "select job_id, job_type, job_status, tenant_id, tenant_name from __all_rootservice_job where tenant_name='${NAME}' and job_status ='INPROGRESS' and job_type='ALTER_TENANT_LOCALITY'"
	DeleteUnitSQLTemplate       = "DROP RESOURCE UNIT ${NAME}"
	DeletePoolSQLTemplate       = "DROP RESOURCE POOL ${NAME}"
	DeleteTenantSQLTemplate     = "DROP TENANT ${NAME} FORCE"

	GetObVersionSQL = "SELECT ob_version() as version;"
)
