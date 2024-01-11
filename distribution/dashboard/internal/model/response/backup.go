package response

import "github.com/oceanbase/oceanbase-dashboard/internal/model/param"

type BackupPolicy struct {
	param.BackupPolicyBase `json:",inline"`

	Status              string     `json:"status"`
	OSSAccessSecret     string     `json:"ossAccessSecret,omitempty"`
	BakEncryptionSecret string     `json:"bakEncryptionSecret,omitempty"`
	LatestFullBackup    *BackupJob `json:"latestFullBackup,omitempty"`
	LatestIncrBackup    *BackupJob `json:"latestIncrBackup,omitempty"`
	LatestArchiveJob    *BackupJob `json:"latestArchiveJob,omitempty"`
	LatestCleanJob      *BackupJob `json:"latestCleanJob,omitempty"`
}

type BackupJob struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Enum: FULL, INCR, ARCHIVE, CLEAN
	Type             string `json:"type"`
	TenantName       string `json:"tenantName"`
	BackupPolicyName string `json:"backupPolicyName"`
	Path             string `json:"path"`      // Empty for Clean job
	StartTime        string `json:"startTime"` // Start time of the backup job, StartScnDisplay for ARCHIVE job
	EndTime          string `json:"endTime"`   // End time of the backup job, empty for ARCHIVE job
	Status           string `json:"status"`
	StatusInDatabase string `json:"statusInDatabase"`
	EncryptionSecret string `json:"encryptionSecret,omitempty"`
}
