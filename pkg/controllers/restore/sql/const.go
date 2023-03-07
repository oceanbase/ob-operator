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
	SetParameterTemplate = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"

	GetRestoreSetCurrentSql  = "SELECT job_id, backup_cluster_id, backup_cluster_name, tenant_name, backup_tenant_name, status, restore_finish_timestamp FROM oceanbase.CDB_OB_RESTORE_PROGRESS order by job_id desc;"
	GetRestoreSetHistorySql  = "SELECT job_id, backup_cluster_id, backup_cluster_name, tenant_name, backup_tenant_name, status, restore_finish_timestamp FROM oceanbase.CDB_OB_RESTORE_HISTORY order by job_id desc;"
	GetRestoreConcurrencySql = "select value  from __all_virtual_sys_parameter_stat where name like 'restore_concurrency';"

	CreateResourceUnitSql = "CREATE RESOURCE UNIT ${unit_name} max_cpu ${max_cpu}, max_memory '${max_memory}', max_iops ${max_iops},max_disk_size '${max_disk_size}', max_session_num ${max_session_num}, MIN_CPU=${min_cpu}, MIN_MEMORY= '${min_memory}', MIN_IOPS=${min_iops};"
	CreateResourcePoolSql = "CREATE RESOURCE POOL ${pool_name} UNIT='${unit_name}', UNIT_NUM=${unit_num}, ZONE_LIST=(${zone_list});"

	SetDecryptionTemplate = "SET DECRYPTION IDENTIFIED BY '%s'"
	DoRestoreSql          = "ALTER SYSTEM RESTORE ${dest_tenant} FROM ${source_tenant} at '${dest_path}' UNTIL '${time}' WITH 'backup_cluster_name=${backup_cluster_name}&backup_cluster_id=${backup_cluster_id}&pool_list=${pool_list}&${restore_option}';"
)
