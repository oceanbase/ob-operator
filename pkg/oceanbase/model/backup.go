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

package model

// OBArchiveDest 租户日志归档目标配置, match view CDB_OB_ARCHIVE_DEST
type OBArchiveDest struct {
	TenantId int64  `json:"tenant_id" db:"tenant_id"`
	DestNo   int64  `json:"dest_no" db:"dest_no"`
	Name     string `json:"name" db:"name"`
	Value    string `json:"value" db:"value"`
}

// OBArchiveLogSummary 日志归档任务汇总, match view CDB_OB_ARCHIVELOG_SUMMARY
type OBArchiveLogSummary struct {
	TenantId                  int64  `json:"tenant_id" db:"tenant_id"`
	DestId                    int64  `json:"dest_id" db:"dest_id"`
	RoundId                   int64  `json:"round_id" db:"round_id"`
	DestNo                    int64  `json:"dest_no" db:"dest_no"`
	Status                    int64  `json:"status" db:"status"`
	StartScn                  int64  `json:"start_scn" db:"start_scn"`
	StartScnDisplay           string `json:"start_scn_display" db:"start_scn_display"`
	CheckpointScn             int64  `json:"checkpoint_scn" db:"checkpoint_scn"`
	CheckpointScnDisplay      string `json:"checkpoint_scn_display" db:"checkpoint_scn_display"`
	Compatible                string `json:"compatible" db:"compatible"`
	BasePieceId               int64  `json:"base_piece_id" db:"base_piece_id"`
	UsedPieceId               int64  `json:"used_piece_id" db:"used_piece_id"`
	PieceSwitchInterval       string `json:"piece_switch_interval" db:"piece_switch_interval"`
	InputBytes                int64  `json:"input_bytes" db:"input_bytes"`
	InputBytesDisplay         string `json:"input_bytes_display" db:"input_bytes_display"`
	OutputBytes               int64  `json:"output_bytes" db:"output_bytes"`
	OutputBytesDisplay        string `json:"output_bytes_display" db:"output_bytes_display"`
	CompressionRatio          string `json:"compression_ratio" db:"compression_ratio"`
	DeletedInputBytes         int64  `json:"deleted_input_bytes" db:"deleted_input_bytes"`
	DeletedInputBytesDisplay  string `json:"deleted_input_bytes_display" db:"deleted_input_bytes_display"`
	DeletedOutputBytes        int64  `json:"deleted_output_bytes" db:"deleted_output_bytes"`
	DeletedOutputBytesDisplay string `json:"deleted_output_bytes_display" db:"deleted_output_bytes_display"`
	Comment                   string `json:"comment" db:"comment"`
	Path                      string `json:"path" db:"path"`
}

type JobCommon struct {
	TenantId         int64  `json:"tenant_id" db:"tenant_id"`
	JobId            int64  `json:"job_id" db:"job_id"`
	ExecutorTenantId int64  `json:"executor_tenant_id" db:"executor_tenant_id"`
	JobLevel         string `json:"job_level" db:"job_level"`
	StartTimestamp   string `json:"start_timestamp" db:"start_timestamp"`
	EndTimestamp     string `json:"end_timestamp" db:"end_timestamp"`
	Status           string `json:"status" db:"status"`
	Result           string `json:"result" db:"result"`
	Comment          string `json:"comment" db:"comment"`
}

// OBBackupJob 数据备份进度, match view CDB_OB_BACKUP_JOB & CDB_OB_BACKUP_JOB_HISTORY
type OBBackupJob struct {
	JobCommon `json:",inline" db:",inline"`

	BackupSetId    int64  `json:"backup_set_id" db:"backup_set_id"`
	PlusArchiveLog string `json:"plus_archive_log" db:"plus_archive_log"`
	BackupType     string `json:"backup_type" db:"backup_type"`
	EncryptionMod  string `json:"encryption_mod" db:"encryption_mod"`
	Password       string `json:"passwd" db:"passwd"`
}

type OBBackupJobHistory OBBackupJob

// OBBackupCleanPolicy 数据备份清理策略, match view CDB_OB_BACKUP_DELETE_POLICY
type OBBackupCleanPolicy struct {
	TenantId      int64  `json:"tenant_id" db:"tenant_id"`
	PolicyName    string `json:"policy_name" db:"policy_name"`
	RecoverWindow string `json:"recover_window" db:"recover_window"`
}

// OBBackupCleanJob 数据备份清理进度, match view CDB_OB_BACKUP_DELETE_JOBS & CDB_OB_BACKUP_DELETE_JOB_HISTORY
type OBBackupCleanJob struct {
	JobCommon `json:",inline" db:",inline"`

	Type             string `json:"type" db:"type"`
	Parameter        string `json:"parameter" db:"parameter"`
	TaskCount        int64  `json:"task_count" db:"task_count"`
	SuccessTaskCount int64  `json:"success_task_count" db:"success_task_count"`
}
