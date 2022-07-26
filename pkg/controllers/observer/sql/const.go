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
	SetTimeoutSQL                   = "SET ob_query_timeout = 600000000"
	SetServerOfflineTimeSQLTemplate = "ALTER SYSTEM SET server_permanent_offline_time=${OFFLINE_TIME};"

	GetOBServerSQL       = "SELECT id, zone, svr_ip, svr_port, inner_port, with_rootserver, with_partition, status, start_service_time FROM __all_server;"
	AddServerSQLTemplate = "ALTER SYSTEM ADD SERVER '${SERVER_IP}' ZONE '${ZONE_NAME}';"
	DelServerSQLTemplate = "ALTER SYSTEM DELETE SERVER '${SERVER_IP}';"

	GetOBZoneSQL         = "SELECT zone, name, value, info FROM __all_zone WHERE name = 'status';"
	AddZoneSQLTemplate   = "ALTER SYSTEM ADD ZONE '${ZONE_NAME}';"
	StartZoneSQLTemplate = "ALTER SYSTEM START ZONE '${ZONE_NAME}';"

	GetRootServiceSQL = "SELECT zone, svr_ip, svr_port, role, partition_id, partition_cnt FROM __all_virtual_core_meta_table;"

	GetRSJobStatusSQL = "SELECT job_status, return_code, progress FROM __all_rootservice_job WHERE job_type = 'DELETE_SERVER' AND svr_ip = '${DELETE_SERVER_IP}' AND svr_port = '${DELETE_SERVER_PORT}';"

	CreateUserSQLTemplate = "CREATE USER ${USER} identified by '${PASSWORD}';"

	GrantPrivilegeSQLTemplate = "GRANT ${PRIVILEGE} on ${OBJECT} to ${USER};"

	SetParameterTemplate = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"

	GetParameterTemplate = "SELECT zone, svr_ip, svr_port, name, value, scope, edit_level FROM __ALL_VIRTUAL_SYS_PARAMETER_STAT WHERE NAME = '${NAME}'"
)
