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

// TODO: filter tenant
// TODO: select specific columns instead of '*' in sql

// TIPS: use parameter to set log archive dest and data backup dest

const (
	EnableArchiveLog       = "ALTER SYSTEM ARCHIVELOG TENANT = ?"
	DisableArchiveLog      = "ALTER SYSTEM NOARCHIVELOG TENANT = ?"
	QueryPieceInfo         = "SELECT * from CDB_OB_ARCHIVELOG_PIECE_FILES"
	QueryArchiveLog        = "SELECT * from CDB_OB_ARCHIVELOG"
	QueryArchiveLogSummary = "SELECT * FROM CDB_OB_ARCHIVELOG_SUMMARY"

	SetBackupPassword  = "SET ENCRYPTION ON IDENTIFIED BY ? ONLY"
	CreateBackupFull   = "ALTER SYSTEM BACKUP TENANT = ?"
	CreateBackupIncr   = "ALTER SYSTEM BACKUP INCREMENTAL TENANT = ?"
	StopBackupJob      = "ALTER SYSTEM CANCEL BACKUP TENANT = ?"
	QueryBackupJobs    = "SELECT * FROM CDB_OB_BACKUP_JOBS"
	QueryBackupHistory = "SELECT * FROM CDB_OB_BACKUP_JOB_HISTORY"

	AddCleanBackupPolicy       = "ALTER SYSTEM ADD DELETE BACKUP POLICY ? RECOVERY_WINDOW ? TENANT ?"
	RemoveCleanBackupPolicy    = "ALTER SYSTEM DROP DELETE BACKUP POLICY ? TENANT ?"
	QueryBackupCleanPolicy     = "SELECT * FROM CDB_OB_BACKUP_DELETE_POLICY"
	CancelCleanBackup          = "ALTER SYSTEM CANCEL DELETE BACKUP TENANT ?"
	CancelAllCleanBackup       = "ALTER SYSTEM CANCEL DELETE BACKUP"
	QueryBackupCleanJobs       = "SELECT * FROM CDB_OB_BACKUP_DELETE_JOBS"
	QueryBackupCleanJobHistory = "SELECT * FROM CDB_OB_BACKUP_DELETE_JOB_HISTORY"
)
