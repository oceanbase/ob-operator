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
	GetBackupSetSQL = "SELECT tenant_id, bs_key, backup_type, status FROM oceanbase.CDB_OB_BACKUP_SET_FILES;"

	SetParameterTemplate      = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"
	SetBackupPasswordTemplate = "SET ENCRYPTION ON IDENTIFIED BY '${pwd}' ONLY"

	StartArchieveLogSql     = "ALTER SYSTEM ARCHIVELOG"
	GetArchieveLogStatusSql = "SELECT tenant_id, status FROM CDB_OB_BACKUP_ARCHIVELOG"

	StartBackupDatabaseSql    = "ALTER SYSTEM BACKUP DATABASE"
	StartBackupIncrementalSql = "ALTER SYSTEM BACKUP INCREMENTAL DATABASE"

	GetBackupDestSql = "select zone, svr_ip, svr_port, value from __all_virtual_sys_parameter_stat where name like 'backup_dest';"
)
