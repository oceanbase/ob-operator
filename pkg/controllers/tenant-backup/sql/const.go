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
	GetBackupFullJobSQLTemplate = "SELECT backup_set_id, backup_type, status FROM oceanbase.CDB_OB_BACKUP_JOB_HISTORY WHERE tenant_id =(SELECT tenant_id FROM __all_tenant WHERE tenant_name='${NAME}') AND backup_type='FULL' ORDER BY job_id DESC LIMIT 1;"
	GetBackupIncJobSQLTemplate  = "SELECT backup_set_id, backup_type, status FROM oceanbase.CDB_OB_BACKUP_JOB_HISTORY WHERE tenant_id =(SELECT tenant_id FROM __all_tenant WHERE tenant_name='${NAME}') AND backup_type='INCREMENTAL' ORDER BY job_id DESC LIMIT 1;"

	SetParameterTemplate      = "ALTER SYSTEM SET ${NAME} = '${VALUE}'"
	ShowParameterTemplate     = "SHOW PARAMETERS LIKE '${NAME}'"
	GetArchiveLogDestSQL      = "SELECT dest_no, name, value FROM oceanbase.DBA_OB_ARCHIVE_DEST;"
	GetBackupDestSQL          = "SELECT name, value FROM oceanbase.DBA_OB_BACKUP_PARAMETER WHERE name='data_backup_dest';"
	SetBackupPasswordTemplate = "SET ENCRYPTION ON IDENTIFIED BY '${pwd}' ONLY"

	GetArchiveLogSQL          = "SELECT dest_no, status, start_scn, checkpoint_scn, base_piece_id, used_piece_id FROM oceanbase.DBA_OB_ARCHIVELOG;"
	GetArchieveLogStatusSql   = "SELECT tenant_id, status FROM CDB_OB_BACKUP_ARCHIVELOG"
	StartArchiveLogSQL        = "ALTER SYSTEM ARCHIVELOG"
	StopArchiveLogSQL         = "ALTER SYSTEM NOARCHIVELOG"
	StartBackupDatabaseSql    = "ALTER SYSTEM BACKUP DATABASE"
	StartBackupIncrementalSql = "ALTER SYSTEM BACKUP INCREMENTAL DATABASE"
)
