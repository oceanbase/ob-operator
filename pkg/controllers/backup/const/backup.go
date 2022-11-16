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

package backupconst

// backup job status

const (
	BackupDoing     = "DOING"
	BackupCompleted = "COMPLETED"
	BackupFailed    = "FAILED"
)

const (
	// full
	FullBackup         = "f"
	FullBackupType     = "FULL"
	DatabaseBackupType = "D"
	// incremental
	IncrementalBackup     = "i"
	IncrementalBackupType = "INCREMENTAL"
	IncDatabaseBackupType = "I"
)

const (
	BackupOnce = "once"
)

const (
	ArchiveLogBeginning = "BEGINNING"
	ArchiveLogDoing     = "DOING"
)

const (
	BackupDestOptionName                 = "backup_dest_option"
	LogArchiveCheckpointIntervalName     = "log_archive_checkpoint_interval"
	LogArchiveCheckpointIntervalDefault  = "120s"
	RecoveryWindowName                   = "recovery_window"
	RecoveryWindowDefault                = "0"
	AutoDeleteObsoleteBackupName         = "auto_delete_obsolete_backup"
	AutoDeleteObsoleteBackupDefault      = "false"
	LogArchivePieceSwitchIntervalName    = "log_archive_piece_switch_interval"
	LogArchivePieceSwitchIntervalDefault = "0"

	DestPathName = "dest_path"

	BackupModeName                 = "backup_mode"
	BackupModeDefault              = "optional"
	BackupCompressName             = "backup_compress"
	BackupCompressDefault          = "disable"
	BackupCompressAlgorithmName    = "backup_compress_algorithm"
	BackupCompressAlgorithmDefault = "lz4_1.0"

	BackupLogArchiveOptionName    = "backup_log_archive_option"
	BackupDatabasePasswordName    = "backup_database_password"
	BackupIncrementalPasswordName = "backup_incremental_password"
)
