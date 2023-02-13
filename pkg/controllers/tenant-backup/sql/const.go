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
	GetBackupSetSQL = "SELECT tenant_id, backup_set_id, backup_type, status FROM oceanbase.CDB_OB_BACKUP_JOB_HISTORY;"

	SetParameterTemplate      = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"
	ShowParameterTemplate     = "SHOW PARAMETERS LIKE '${NAME}'"
	GetArchieveLogDestSQL     = "SELECT dest_no, name, value FROM oceanbase.DBA_OB_ARCHIVE_DEST;"
	SetBackupPasswordTemplate = "SET ENCRYPTION ON IDENTIFIED BY '${pwd}' ONLY"

	StartArchieveLogSql     = "ALTER SYSTEM ARCHIVELOG"
	GetArchieveLogStatusSql = "SELECT tenant_id, status FROM CDB_OB_BACKUP_ARCHIVELOG"

	StartBackupDatabaseSql    = "ALTER SYSTEM BACKUP DATABASE"
	StartBackupIncrementalSql = "ALTER SYSTEM BACKUP INCREMENTAL DATABASE"

	GetBackupDestSql = "SELECT zone, svr_ip, svr_port, value from __all_virtual_sys_parameter_stat where name like 'backup_dest';"
)
