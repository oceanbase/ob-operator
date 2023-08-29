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

// OBArchiveDest 租户日志归档目标配置, match view DBA_OB_ARCHIVE_DEST
type OBArchiveDest struct {
	TenantId int64  `json:"tenant_id" db:"tenant_id"`
	DestNo   int64  `json:"dest_no" db:"dest_no"`
	Name     string `json:"name" db:"name"`
	Value    string `json:"value" db:"value"`
}

// OBArchiveLogSummary matches view DBA_OB_ARCHIVELOG_SUMMARY
type OBArchiveLogSummary struct {
	TenantId                  int64  `json:"tenant_id" db:"tenant_id"`
	DestId                    int64  `json:"dest_id" db:"dest_id"`
	RoundId                   int64  `json:"round_id" db:"round_id"`
	DestNo                    int64  `json:"dest_no" db:"dest_no"`
	Status                    string `json:"status" db:"status"`
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

// OBArchiveLogJob is equal to OBArchiveLogSummary, but match view DBA_OB_ARCHIVELOG_JOBS
type OBArchiveLogJob OBArchiveLogSummary

type JobCommon struct {
	TenantId         int64   `json:"tenant_id" db:"tenant_id"`
	JobId            int64   `json:"job_id" db:"job_id"`
	ExecutorTenantId int64   `json:"executor_tenant_id" db:"executor_tenant_id"`
	JobLevel         string  `json:"job_level" db:"job_level"`
	StartTimestamp   string  `json:"start_timestamp" db:"start_timestamp"`
	EndTimestamp     *string `json:"end_timestamp,omitempty" db:"end_timestamp"`
	Status           string  `json:"status" db:"status"`
	Result           string  `json:"result" db:"result"`
	Comment          string  `json:"comment" db:"comment"`
}

// OBBackupJob matches view DBA_OB_BACKUP_JOBS & DBA_OB_BACKUP_JOB_HISTORY
type OBBackupJob struct {
	JobCommon `json:",inline" db:",inline"`

	BackupSetId    int64  `json:"backup_set_id" db:"backup_set_id"`
	PlusArchiveLog string `json:"plus_archivelog" db:"plus_archivelog"`
	BackupType     string `json:"backup_type" db:"backup_type"`
	EncryptionMode string `json:"encryption_mode" db:"encryption_mode"`
	Password       string `json:"passwd" db:"passwd"`
}

type OBBackupJobHistory OBBackupJob

// OBBackupCleanPolicy matches view DBA_OB_BACKUP_DELETE_POLICY
type OBBackupCleanPolicy struct {
	TenantId       int64  `json:"tenant_id" db:"tenant_id"`
	PolicyName     string `json:"policy_name" db:"policy_name"`
	RecoveryWindow string `json:"recovery_window" db:"recovery_window"`
}

// OBBackupCleanJob matches view DBA_OB_BACKUP_DELETE_JOBS & DBA_OB_BACKUP_DELETE_JOB_HISTORY
type OBBackupCleanJob struct {
	JobCommon `json:",inline" db:",inline"`

	Type             string `json:"type" db:"type"`
	Parameter        string `json:"parameter" db:"parameter"`
	TaskCount        int64  `json:"task_count" db:"task_count"`
	SuccessTaskCount int64  `json:"success_task_count" db:"success_task_count"`
}

// OBBackupTask belonging to OBBackupJob, matches DBA_OB_BACKUP_TASKS
type OBBackupTask struct {
	TenantId       int64   `json:"tenant_id" db:"tenant_id"`
	JobId          int64   `json:"job_id" db:"job_id"`
	BackupSetId    int64   `json:"backup_set_id" db:"backup_set_id"`
	StartTimestamp string  `json:"start_timestamp" db:"start_timestamp"`
	EndTimestamp   *string `json:"end_timestamp,omitempty" db:"end_timestamp"`
	Status         string  `json:"status" db:"status"`
	Result         string  `json:"result" db:"result"`
	Comment        string  `json:"comment" db:"comment"`

	TaskId                int64  `json:"task_id" db:"task_id"`
	Incarnation           int64  `json:"incarnation" db:"incarnation"`
	StartScn              int64  `json:"start_scn" db:"start_scn"`
	EndScn                int64  `json:"end_scn" db:"end_scn"`
	UserLsStartScn        int64  `json:"user_ls_start_scn" db:"user_ls_start_scn"`
	EncryptionMode        string `json:"encryption_mode" db:"encryption_mode"`
	Password              string `json:"passwd" db:"passwd"`
	InputBytes            int64  `json:"input_bytes" db:"input_bytes"`
	OutputBytes           int64  `json:"output_bytes" db:"output_bytes"`
	OutputRateBytes       string `json:"output_rate_bytes" db:"output_rate_bytes"`
	ExtraMetaBytes        int64  `json:"extra_meta_bytes" db:"extra_meta_bytes"`
	TabletCount           int64  `json:"tablet_count" db:"tablet_count"`
	FinishTabletCount     int64  `json:"finish_tablet_count" db:"finish_tablet_count"`
	MacroBlockCount       int64  `json:"macro_block_count" db:"macro_block_count"`
	FinishMacroBlockCount int64  `json:"finish_macro_block_count" db:"finish_macro_block_count"`
	FileCount             int64  `json:"file_count" db:"file_count"`
	MetaTurnId            int64  `json:"meta_turn_id" db:"meta_turn_id"`
	DataTurnId            int64  `json:"data_turn_id" db:"data_turn_id"`
	Path                  string `json:"path" db:"path"`
}

// OBArchiveLogPieceFile matches DBA_OB_ARCHIVELOG_PIECE_FILES
type OBArchiveLogPieceFile struct {
	DestId               int64  `json:"dest_id" db:"dest_id"`
	RoundId              int64  `json:"round_id" db:"round_id"`
	PieceId              int64  `json:"piece_id" db:"piece_id"`
	Incarnation          int64  `json:"incarnation" db:"incarnation"`
	DestNo               int64  `json:"dest_no" db:"dest_no"`
	Status               string `json:"status" db:"status"`
	StartScn             int64  `json:"start_scn" db:"start_scn"`
	StartScnDisplay      string `json:"start_scn_display" db:"start_scn_display"`
	CheckpointScn        int64  `json:"checkpoint_scn" db:"checkpoint_scn"`
	CheckpointScnDisplay string `json:"checkpoint_scn_display" db:"checkpoint_scn_display"`
	MaxScn               int64  `json:"max_scn" db:"max_scn"`
	EndScn               int64  `json:"end_scn" db:"end_scn"`
	EndScnDisplay        string `json:"end_scn_display" db:"end_scn_display"`
	Compatible           string `json:"compatible" db:"compatible"`
	UnitSize             int64  `json:"unit_size" db:"unit_size"`
	Compression          string `json:"compression" db:"compression"`
	InputBytes           int64  `json:"input_bytes" db:"input_bytes"`
	InputBytesDisplay    string `json:"input_bytes_display" db:"input_bytes_display"`
	OutputBytes          int64  `json:"output_bytes" db:"output_bytes"`
	OutputBytesDisplay   string `json:"output_bytes_display" db:"output_bytes_display"`
	CompressionRatio     string `json:"compression_ratio" db:"compression_ratio"`
	FileStatus           int64  `json:"file_status" db:"file_status"`
	Path                 string `json:"path" db:"path"`
}
