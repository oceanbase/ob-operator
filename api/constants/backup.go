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

package constants

import "github.com/oceanbase/ob-operator/api/types"

const (
	BackupJobTypeFull    types.BackupJobType = "FULL"
	BackupJobTypeIncr    types.BackupJobType = "INC"
	BackupJobTypeClean   types.BackupJobType = "CLEAN"
	BackupJobTypeArchive types.BackupJobType = "ARCHIVE"
)

const (
	BackupJobStatusRunning      types.BackupJobStatus = "RUNNING"
	BackupJobStatusInitializing types.BackupJobStatus = "INITIALIZING"
	BackupJobStatusSuccessful   types.BackupJobStatus = "SUCCESSFUL"
	BackupJobStatusFailed       types.BackupJobStatus = "FAILED"
	BackupJobStatusCanceled     types.BackupJobStatus = "CANCELED"
	BackupJobStatusStopped      types.BackupJobStatus = "STOPPED"
	BackupJobStatusSuspend      types.BackupJobStatus = "SUSPEND"
)

const (
	BackupPolicyStatusPreparing   types.BackupPolicyStatusType = "PREPARING"
	BackupPolicyStatusPrepared    types.BackupPolicyStatusType = "PREPARED"
	BackupPolicyStatusRunning     types.BackupPolicyStatusType = "RUNNING"
	BackupPolicyStatusFailed      types.BackupPolicyStatusType = "FAILED"
	BackupPolicyStatusPausing     types.BackupPolicyStatusType = "PAUSING"
	BackupPolicyStatusPaused      types.BackupPolicyStatusType = "PAUSED"
	BackupPolicyStatusStopped     types.BackupPolicyStatusType = "STOPPED"
	BackupPolicyStatusResuming    types.BackupPolicyStatusType = "RESUMING"
	BackupPolicyStatusDeleting    types.BackupPolicyStatusType = "DELETING"
	BackupPolicyStatusMaintaining types.BackupPolicyStatusType = "MAINTAINING"
)

const (
	BackupDestTypeOSS types.BackupDestType = "OSS"
	BackupDestTypeNFS types.BackupDestType = "NFS"
)

const (
	LogArchiveDestStateEnable types.LogArchiveDestState = "ENABLE"
	LogArchiveDestStateDefer  types.LogArchiveDestState = "DEFER"
)

const (
	ArchiveBindingOptional  types.ArchiveBinding = "Optional"
	ArchiveBindingMandatory types.ArchiveBinding = "Mandatory"
)
