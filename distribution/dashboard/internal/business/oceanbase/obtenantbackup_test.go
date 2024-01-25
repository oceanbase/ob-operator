package oceanbase

import (
	"testing"

	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateOBTenantBackupPolicyWeekly(t *testing.T) {
	scheduleDates := []param.ScheduleDate{{
		Day:        1,
		BackupType: "Full",
	}, {
		Day:        2,
		BackupType: "Incremental",
	}, {
		Day:        3,
		BackupType: "Incremental",
	}, {
		Day:        4,
		BackupType: "Incremental",
	}, {
		Day:        5,
		BackupType: "Full",
	}}
	p := param.CreateBackupPolicy{
		BackupPolicyBase: param.BackupPolicyBase{
			DestType:       "NFS",
			ArchivePath:    "archive/t1",
			BakDataPath:    "backup/t1",
			ScheduleType:   "Weekly",
			ScheduleDates:  scheduleDates,
			ScheduleTime:   "04:00",
			JobKeepWindow:  "3d",
			RecoveryWindow: "7d",
		},
	}
	policy := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
	assert.Equal(t, "t1-backup-policy", policy.Name)
	assert.Equal(t, "00 04 * * 1,5", policy.Spec.DataBackup.FullCrontab)
	assert.Equal(t, "00 04 * * 2,3,4", policy.Spec.DataBackup.IncrementalCrontab)

	policyModel := buildBackupPolicyModelType(policy)
	assert.Equal(t, "t1-backup-policy", policyModel.Name)
	assert.EqualValues(t, "NFS", policyModel.DestType)
	assert.Equal(t, "archive/t1", policyModel.ArchivePath)
	assert.Equal(t, "backup/t1", policyModel.BakDataPath)
	assert.Equal(t, "Weekly", policyModel.ScheduleType)
	assert.Equal(t, scheduleDates, policyModel.ScheduleDates)
}

func TestCreateOBTenantBackupPolicyMonthly(t *testing.T) {
	scheduleDates := []param.ScheduleDate{{
		Day:        1,
		BackupType: "Full",
	}, {
		Day:        2,
		BackupType: "Incremental",
	}, {
		Day:        3,
		BackupType: "Incremental",
	}, {
		Day:        4,
		BackupType: "Incremental",
	}, {
		Day:        5,
		BackupType: "Full",
	}, {
		Day:        15,
		BackupType: "Full",
	}, {
		Day:        16,
		BackupType: "Incremental",
	}, {
		Day:        21,
		BackupType: "Full",
	}, {
		Day:        24,
		BackupType: "Incremental",
	}, {
		Day:        31,
		BackupType: "Full",
	}}
	p := param.CreateBackupPolicy{
		BackupPolicyBase: param.BackupPolicyBase{
			DestType:       "NFS",
			ArchivePath:    "archive/t1",
			BakDataPath:    "backup/t1",
			ScheduleType:   "Monthly",
			ScheduleDates:  scheduleDates,
			ScheduleTime:   "04:00",
			JobKeepWindow:  "3d",
			RecoveryWindow: "7d",
		},
	}
	policy := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
	assert.Equal(t, "t1-backup-policy", policy.Name)
	assert.Equal(t, "00 04 1,5,15,21,31 * *", policy.Spec.DataBackup.FullCrontab)
	assert.Equal(t, "00 04 2,3,4,16,24 * *", policy.Spec.DataBackup.IncrementalCrontab)

	policyModel := buildBackupPolicyModelType(policy)
	assert.Equal(t, "t1-backup-policy", policyModel.Name)
	assert.EqualValues(t, "NFS", policyModel.DestType)
	assert.Equal(t, "archive/t1", policyModel.ArchivePath)
	assert.Equal(t, "backup/t1", policyModel.BakDataPath)
	assert.Equal(t, "Monthly", policyModel.ScheduleType)
	assert.Equal(t, scheduleDates, policyModel.ScheduleDates)
}
