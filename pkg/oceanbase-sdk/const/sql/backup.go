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

// TIPS: use parameter to set log archive dest and data backup dest

const jobCommonFields = "job_id, executor_tenant_id, job_level, start_timestamp, end_timestamp, status, result, comment"
const backupJobFields = jobCommonFields + ", backup_set_id, plus_archivelog, backup_type, encryption_mode, passwd"
const backupTaskFields = "job_id, backup_set_id, start_timestamp, end_timestamp, status, result, comment, task_id, incarnation, start_scn, end_scn, user_ls_start_scn, encryption_mode, passwd, input_bytes, output_bytes, output_rate_bytes, extra_meta_bytes, tablet_count, finish_tablet_count, macro_block_count, finish_macro_block_count, file_count, meta_turn_id, data_turn_id, path"
const cleanJobFields = jobCommonFields + ", type, parameter, task_count, success_task_count"
const logArchiveJobFields = "dest_id, round_id, dest_no, status, start_scn, start_scn_display, checkpoint_scn, IFNULL(checkpoint_scn_display, '-') as checkpoint_scn_display, compatible, base_piece_id, used_piece_id, piece_switch_interval, input_bytes, input_bytes_display, output_bytes, output_bytes_display, compression_ratio, deleted_input_bytes, deleted_input_bytes_display, deleted_output_bytes, deleted_output_bytes_display, comment, path"
const logArchivePieceFileFields = "dest_id, round_id, piece_id, incarnation, dest_no, status, start_scn, start_scn_display, checkpoint_scn, IFNULL(checkpoint_scn_display, '-') as checkpoint_scn_display, max_scn, end_scn, end_scn_display, compatible, unit_size, compression, input_bytes, input_bytes_display, output_bytes, output_bytes_display, compression_ratio, file_status, path"

const (
	EnableArchiveLog           = "ALTER SYSTEM ARCHIVELOG"
	DisableArchiveLog          = "ALTER SYSTEM NOARCHIVELOG"
	QueryPieceInfo             = "SELECT " + logArchivePieceFileFields + " FROM DBA_OB_ARCHIVELOG_PIECE_FILES"
	QueryArchiveLog            = "SELECT " + logArchiveJobFields + " FROM DBA_OB_ARCHIVELOG"
	QueryArchiveLogSummary     = "SELECT " + logArchiveJobFields + " FROM DBA_OB_ARCHIVELOG_SUMMARY"
	QueryArchiveLogDestConfigs = "SELECT dest_no, name, value FROM DBA_OB_ARCHIVE_DEST"

	SetBackupPassword      = "SET ENCRYPTION ON IDENTIFIED BY ? ONLY"
	CreateBackupFull       = "ALTER SYSTEM BACKUP DATABASE"
	CreateBackupIncr       = "ALTER SYSTEM BACKUP INCREMENTAL DATABASE"
	StopBackupJob          = "ALTER SYSTEM CANCEL BACKUP"
	QueryBackupJobs        = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOBS"
	QueryBackupHistory     = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOB_HISTORY"
	QueryBackupTasks       = "SELECT " + backupTaskFields + " FROM DBA_OB_BACKUP_TASKS"
	QueryBackupTaskHistory = "SELECT " + backupTaskFields + " FROM DBA_OB_BACKUP_TASK_HISTORY"
	QueryBackupParameter   = "SELECT name, value FROM DBA_OB_BACKUP_PARAMETER"

	AddCleanBackupPolicy       = "ALTER SYSTEM ADD DELETE BACKUP POLICY ? RECOVERY_WINDOW ?"
	RemoveCleanBackupPolicy    = "ALTER SYSTEM DROP DELETE BACKUP POLICY ?"
	QueryBackupCleanPolicy     = "SELECT policy_name, recovery_window FROM DBA_OB_BACKUP_DELETE_POLICY"
	CancelCleanBackup          = "ALTER SYSTEM CANCEL DELETE BACKUP"
	CancelAllCleanBackup       = "ALTER SYSTEM CANCEL DELETE BACKUP"
	QueryBackupCleanJobs       = "SELECT " + cleanJobFields + " FROM DBA_OB_BACKUP_DELETE_JOBS"
	QueryBackupCleanJobHistory = "SELECT " + cleanJobFields + " FROM DBA_OB_BACKUP_DELETE_JOB_HISTORY"

	QueryLatestBackupJobOfType               = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOBS WHERE backup_type = ? order by job_id DESC LIMIT 1"
	QueryLatestBackupJobHistoryOfType        = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOB_HISTORY WHERE backup_type = ? order by job_id DESC LIMIT 1"
	QueryLatestBackupJobOfTypeAndPath        = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOBS WHERE backup_type = ? and path = ? order by job_id DESC LIMIT 1"
	QueryLatestBackupJobHistoryOfTypeAndPath = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOB_HISTORY WHERE backup_type = ? and path = ? order by job_id DESC LIMIT 1"
	QueryLatestRunningBackupJob              = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOBS order by job_id DESC LIMIT 1"
	QueryLatestCleanJob                      = "SELECT " + cleanJobFields + " FROM DBA_OB_BACKUP_DELETE_JOBS ORDER BY job_id DESC LIMIT 1"
	QueryLatestCleanJobHistory               = "SELECT " + cleanJobFields + " FROM DBA_OB_BACKUP_DELETE_JOB_HISTORY ORDER BY job_id DESC LIMIT 1"
	QueryLatestArchiveLogJob                 = "SELECT " + logArchiveJobFields + " FROM DBA_OB_ARCHIVELOG_SUMMARY ORDER BY round_id DESC LIMIT 1"

	QueryBackupJobWithId            = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOBS WHERE job_id = ?"
	QueryBackupHistoryWithId        = "SELECT " + backupJobFields + " FROM DBA_OB_BACKUP_JOB_HISTORY WHERE job_id = ?"
	QueryBackupTaskWithJobId        = "SELECT " + backupTaskFields + " FROM DBA_OB_BACKUP_TASKS WHERE job_id = ?"
	QueryBackupTaskHistoryWithJobId = "SELECT " + backupTaskFields + " FROM DBA_OB_BACKUP_TASK_HISTORY WHERE job_id = ?"
)
