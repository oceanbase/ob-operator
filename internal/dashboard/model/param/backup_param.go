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

package param

type ScheduleBase struct {
	// Enum: Weekly, Monthly
	ScheduleType  string         `json:"scheduleType" example:"Weekly"`
	ScheduleDates []ScheduleDate `json:"scheduleDates"`
	// Description: HH:MM
	// Example: 04:00
	ScheduleTime string `json:"scheduleTime" example:"04:00"`
}

type DaysFieldBase struct {
	JobKeepDays       int `json:"jobKeepDays,omitempty" example:"5"`
	RecoveryDays      int `json:"recoveryDays,omitempty" example:"3"`
	PieceIntervalDays int `json:"pieceIntervalDays,omitempty" example:"1"`
}

type BackupPolicyBase struct {
	// Enum: NFS, OSS
	DestType    BackupDestType `json:"destType" binding:"required" example:"NFS"`
	ArchivePath string         `json:"archivePath" binding:"required"`
	BakDataPath string         `json:"bakDataPath" binding:"required"`

	ScheduleBase  `json:",inline"`
	DaysFieldBase `json:",inline"`
}

type CreateBackupPolicy struct {
	BackupPolicyBase      `json:",inline"`
	OSSAccessID           string `json:"ossAccessId,omitempty" example:"encryptedPassword"`
	OSSAccessKey          string `json:"ossAccessKey,omitempty" example:"encryptedPassword"`
	BakEncryptionPassword string `json:"bakEncryptionPassword,omitempty" example:"encryptedPassword"`
}

type ScheduleDate struct {
	// Description: 1-31 for monthly, 1-7 for weekly
	Day int `json:"day" binding:"required" example:"3"`
	// Enum: Full, Incremental
	BackupType string `json:"backupType" binding:"required" example:"Full"`
}

type UpdateBackupPolicy struct {
	// Enum: PAUSED, RUNNING
	Status string `json:"status,omitempty" example:"PAUSED"`

	ScheduleBase  `json:",inline,omitempty"`
	DaysFieldBase `json:",inline,omitempty"`
}
