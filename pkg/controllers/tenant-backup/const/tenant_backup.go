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

package tenantBackupconst

import "time"

// tennat secret
const (
	User              = "user"
	UserSecret        = "password"
	IncrementalSecret = "incremental"
	DatabaseSecret    = "database"
)

const (
	LogAechiveDest      = "LOG_ARCHIVE_DEST"
	Path                = "path"
	Binding             = "binding"
	PieceSwitchInterval = "piece_switch_interval"
	DataBackupDest      = "data_backup_dest"
)

const (
	ArchiveLogPrepare     = "PREPARE"
	ArchiveLogBeginning   = "BEGINNING"
	ArchiveLogDoing       = "DOING"
	ArchiveLogStop        = "STOP"
	ArchiveLogStopping    = "STOPPING"
	ArchiveLogInterrupted = "INTERRUPTED"
)

const (
	TickPeriodForArchiveLogrStatusCheck = 5 * time.Second
	TickNumForArchiveLogStatusCheck     = 20
)

const (
	// full
	FullBackup     = "F"
	FullBackupType = "FULL"
	// incremental
	IncrementalBackup     = "I"
	IncrementalBackupType = "INCREMENTAL"
)

const (
	BackupOnce = "once"
)

// backup job status
const (
	BackupDoing     = "DOING"
	BackupCompleted = "COMPLETED"
	BackupFailed    = "FAILED"
)
