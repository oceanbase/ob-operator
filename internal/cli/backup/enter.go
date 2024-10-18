/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package backup

// Flagsets and Flags for backup policy management
const (
	// Flagsets
	FLAGSET_DAYS_FIELD = "days-field-flags"
	FLAGSET_SCHEDULE   = "schedule-flags"
	FLAGSET_ACCESS     = "access-flags"

	// Base flags
	FLAG_NAMESPACE     = "namespace"
	FLAG_NAME          = "name"
	FLAG_SCHEDULE_TYPE = "schedule-type"
	FLAG_DEST_TYPE     = "dest-type"
	FLAG_ARCHIVE_PATH  = "archive-path"
	FLAG_BAK_DATA_PATH = "bak-data-path"
	FLAG_STATUS        = "status"

	// Day field flags
	FLAG_JOB_KEEP_DAYS       = "job-keep-days"
	FLAG_RECOVERY_DAYS       = "recovery-days"
	FLAG_PIECE_INTERVAL_DAYS = "piece-interval-days"
	FLAG_SCHEDULE_TIME       = "schedule-time"

	// Schedule flags
	FLAG_INCREMENTAL = "inc"
	FLAG_FULL        = "full"
	// Access flags
	FLAG_OSS_ACCESS_ID           = "oss-access-id"
	FLAG_OSS_ACCESS_KEY          = "oss-access-key"
	FLAG_BAK_ENCRYPTION_PASSWORD = "bak-encryption-password"
)

// Default values for backup policy management
const (
	// Default values for int and string flags
	DEFAULT_NAMESPACE           = "default"
	DEFAULT_JOB_KEEP_DAYS       = 7
	DEFAULT_RECOVERY_DAYS       = 30
	DEFAULT_PIECE_INTERVAL_DAYS = 1
	DEFAULT_DEST_TYPE           = "NFS"
	DEFAULT_SCHEDULE_TYPE       = "weekly"
	DEFAULT_BACKUP_TYPE         = "full"
	DEFAULT_STATUS              = "RUNNING"
)
