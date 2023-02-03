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

	GetOBServerSQL        = "SELECT id, zone, svr_ip, svr_port, inner_port, with_rootserver, with_partition, lower(status) as status, start_service_time FROM __all_server;"
	GetRestartOBServerSQL = "select id, zone, svr_ip, is_in_sync, is_offline FROM __all_virtual_clog_stat"
	AddServerSQLTemplate  = "ALTER SYSTEM ADD SERVER '${SERVER_IP}' ZONE '${ZONE_NAME}';"
	DelServerSQLTemplate  = "ALTER SYSTEM DELETE SERVER '${SERVER_IP}';"

	GetOBZoneSQL         = "SELECT zone, name, value, info FROM __all_zone WHERE name = 'status';"
	AddZoneSQLTemplate   = "ALTER SYSTEM ADD ZONE '${ZONE_NAME}';"
	StartZoneSQLTemplate = "ALTER SYSTEM START ZONE '${ZONE_NAME}';"

	StopOBZoneTemplate   = "ALTER SYSTEM STOP ZONE '${ZONE_NAME}';"
	DeleteOBZoneTemplate = "ALTER SYSTEM DELETE ZONE '${ZONE_NAME}';"

	BeginUpgradeSQL            = "ALTER SYSTEM BEGIN UPGRADE;"
	UpgradeSchemaSQL           = "ALTER SYSTEM UPGRADE VIRTUAL SCHEMA;"
	SetMinOBVersionSQLTemplate = "ALTER SYSTEM SET min_observer_version = '${VERSION}'"
	EndUpgradeSQL              = "ALTER SYSTEM END UPGRADE;"
	RunRootInspectionJobSQL    = "ALTER SYSTEM RUN JOB 'root_inspection'"
	GetRootServiceSQL          = "SELECT zone, svr_ip, svr_port, role FROM __all_virtual_core_meta_table;"
	GetClogStatSQL             = "select svr_ip, svr_port, is_offline, is_in_sync from __all_virtual_clog_stat where is_in_sync=0 and is_offline =0;"
	GetLeaderCountSQL          = "SELECT zone, leader_count FROM oceanbase.__all_virtual_server_stat"
	GetAllUnitSQL              = "SELECT unit_id, resource_pool_id, group_id, zone, svr_ip, svr_port, migrate_from_svr_ip, migrate_from_svr_port, manual_migrate, status, replica_type FROM __all_unit;"

	MajorFreezeSQL          = "ALTER SYSTEM MAJOR FREEZE;"
	GetFrozeVersionSQL      = "SELECT zone, name, value, info from oceanbase.__all_zone where name='frozen_version'"
	GetLastMergedVersionSQL = "SELECT zone, name, value, info from oceanbase.__all_zone where name='last_merged_version' and value != (select value from oceanbase.__all_zone where name='frozen_version')"
	GetObVersionSQL         = "SELECT ob_version() as version;"
	GetRSJobStatusSQL       = "SELECT job_status, progress FROM __all_rootservice_job WHERE job_type = 'DELETE_SERVER' AND svr_ip = '${DELETE_SERVER_IP}' AND svr_port = '${DELETE_SERVER_PORT}';"

	CreateUserSQLTemplate = "CREATE USER ${USER} identified by '${PASSWORD}';"

	GrantPrivilegeSQLTemplate = "GRANT ${PRIVILEGE} on ${OBJECT} to ${USER};"

	SetParameterTemplate = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"

	GetParameterTemplate  = "SELECT zone, svr_ip, svr_port, name, value, scope, edit_level FROM __ALL_VIRTUAL_SYS_PARAMETER_STAT WHERE NAME = '${NAME}'"
	ShowParameterTemplate = "SHOW PARAMETERS LIKE '${NAME}'"
)
