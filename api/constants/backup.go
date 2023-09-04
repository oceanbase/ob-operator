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

type BackupJobType string

const (
	BackupJobTypeFull    BackupJobType = "FULL"
	BackupJobTypeIncr    BackupJobType = "INC"
	BackupJobTypeClean   BackupJobType = "CLEAN"
	BackupJobTypeArchive BackupJobType = "ARCHIVE"
)

type BackupJobStatus string

const (
	BackupJobStatusRunning      BackupJobStatus = "RUNNING"
	BackupJobStatusInitializing BackupJobStatus = "INITIALIZING"
	BackupJobStatusSuccessful   BackupJobStatus = "SUCCESSFUL"
	BackupJobStatusFailed       BackupJobStatus = "FAILED"
	BackupJobStatusCanceled     BackupJobStatus = "CANCELED"
	BackupJobStatusStopped      BackupJobStatus = "STOPPED"
)

type BackupPolicyStatusType string

const (
	BackupPolicyStatusPreparing BackupPolicyStatusType = "PREPARING"
	BackupPolicyStatusPrepared  BackupPolicyStatusType = "PREPARED"
	BackupPolicyStatusRunning   BackupPolicyStatusType = "RUNNING"
	BackupPolicyStatusFailed    BackupPolicyStatusType = "FAILED"
	BackupPolicyStatusPaused    BackupPolicyStatusType = "PAUSED"
	BackupPolicyStatusStopped   BackupPolicyStatusType = "STOPPED"
)

type BackupDestination struct {
	Type BackupDestType `json:"type,omitempty"`
	Path string         `json:"path,omitempty"`
}

type BackupDestType string

const (
	BackupDestTypeOSS BackupDestType = "OSS"
	BackupDestTypeNFS BackupDestType = "NFS"
)
