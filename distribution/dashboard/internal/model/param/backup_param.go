package param

type BackupPolicyBase struct {
	// Enum: NFS, OSS
	DestType    BackupDestType `json:"destType" binding:"required"`
	ArchivePath string         `json:"archivePath"`
	BakDataPath string         `json:"bakDataPath"`

	ScheduleType  string         `json:"scheduleType" binding:"required"`
	ScheduleDates []ScheduleDate `json:"scheduleDates"`
	// Description: HH:MM
	ScheduleTime string `json:"scheduleTime,omitempty"`

	JobKeepWindow  string `json:"jobKeepWindow,omitempty"`
	RecoveryWindow string `json:"recoveryWindow,omitempty"`
	PieceInterval  string `json:"pieceInterval,omitempty"`
}

type CreateBackupPolicy struct {
	BackupPolicyBase      `json:",inline"`
	OSSAccessID           string `json:"ossAccessId,omitempty"`
	OSSAccessKey          string `json:"ossAccessKey,omitempty"`
	BakEncryptionPassword string `json:"bakEncryptionPassword,omitempty"`
}

type ScheduleDate struct {
	Day int `json:"day" binding:"required"`
	// Enum: Full, Incremental
	BackupType string `json:"backupType" binding:"required"`
}

type UpdateBackupPolicy struct {
	// Enum: Paused, Running
	Status string `json:"status" binding:"required"`

	// Enum: Weekly, Monthly
	ScheduleType  string         `json:"scheduleType,omitempty"`
	ScheduleDates []ScheduleDate `json:"scheduleDates,omitempty"`

	JobKeepWindow  string `json:"jobKeepWindow,omitempty"`
	RecoveryWindow string `json:"recoveryWindow,omitempty"`
	PieceInterval  string `json:"pieceInterval,omitempty"`
}
