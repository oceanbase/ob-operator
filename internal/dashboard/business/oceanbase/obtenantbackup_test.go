package oceanbase

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("OBTenantBackup", func() {
	It("Test CreateOBTenantBackupPolicyWeekly", func() {
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
				DestType:    "NFS",
				ArchivePath: "archive/t1",
				BakDataPath: "backup/t1",
				ScheduleBase: param.ScheduleBase{
					ScheduleType:  "Weekly",
					ScheduleDates: scheduleDates,
					ScheduleTime:  "04:00",
				},
				JobKeepDays:  3,
				RecoveryDays: 7,
			},
		}
		policy := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect("t1-backup-policy").To(Equal(policy.Name))
		Expect("00 04 * * 1,5").To(Equal(policy.Spec.DataBackup.FullCrontab))
		Expect("00 04 * * 2,3,4").To(Equal(policy.Spec.DataBackup.IncrementalCrontab))

		policyModel := buildBackupPolicyModelType(policy)
		Expect("t1-backup-policy").To(Equal(policyModel.Name))
		Expect("NFS").To(Equal(string(policyModel.DestType)))
		Expect("archive/t1").To(Equal(policyModel.ArchivePath))
		Expect("backup/t1").To(Equal(policyModel.BakDataPath))
		Expect("Weekly").To(Equal(policyModel.ScheduleType))
		Expect(scheduleDates).To(Equal(policyModel.ScheduleDates))
	})

	It("Test CreateOBTenantBackupPolicyMonthly", func() {
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
				DestType:    "NFS",
				ArchivePath: "archive/t1",
				BakDataPath: "backup/t1",
				ScheduleBase: param.ScheduleBase{
					ScheduleType:  "Monthly",
					ScheduleDates: scheduleDates,
					ScheduleTime:  "04:00",
				},
				JobKeepDays:  3,
				RecoveryDays: 7,
			},
		}
		policy := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect("t1-backup-policy").To(Equal(policy.Name))
		Expect("00 04 1,5,15,21,31 * *").To(Equal(policy.Spec.DataBackup.FullCrontab))
		Expect("00 04 2,3,4,16,24 * *").To(Equal(policy.Spec.DataBackup.IncrementalCrontab))

		policyModel := buildBackupPolicyModelType(policy)
		Expect("t1-backup-policy").To(Equal(policyModel.Name))
		Expect("NFS").To(Equal(string(policyModel.DestType)))
		Expect("archive/t1").To(Equal(policyModel.ArchivePath))
		Expect("backup/t1").To(Equal(policyModel.BakDataPath))
		Expect("Monthly").To(Equal(policyModel.ScheduleType))
		Expect(scheduleDates).To(Equal(policyModel.ScheduleDates))
	})
})
