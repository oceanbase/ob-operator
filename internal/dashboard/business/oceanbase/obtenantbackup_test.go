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

package oceanbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

var _ = Describe("OBTenantBackup", func() {
	It("Rejects unsupported schedule types", func() {
		for _, scheduleType := range []string{"weekly", "Daily", "Weekly\n"} {
			p := param.CreateBackupPolicy{
				BackupPolicyBase: param.BackupPolicyBase{
					ScheduleBase: param.ScheduleBase{
						ScheduleType: scheduleType,
						ScheduleDates: []param.ScheduleDate{
							{Day: 1, BackupType: "Full"},
							{Day: 2, BackupType: "Incremental"},
						},
						ScheduleTime: "04:00",
					},
				},
			}

			_, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
			Expect(err).To(MatchError(ContainSubstring("Schedule type must be Weekly or Monthly")))
		}
	})

	It("Rejects invalid schedule times", func() {
		for _, scheduleTime := range []string{"", "04", "04:00:00", "24:00"} {
			p := param.CreateBackupPolicy{
				BackupPolicyBase: param.BackupPolicyBase{
					ScheduleBase: param.ScheduleBase{
						ScheduleType: "Weekly",
						ScheduleDates: []param.ScheduleDate{
							{Day: 1, BackupType: "Full"},
							{Day: 2, BackupType: "Incremental"},
						},
						ScheduleTime: scheduleTime,
					},
				},
			}

			_, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
			Expect(err).To(MatchError(ContainSubstring("Schedule time must use HH:MM in the 24-hour format")))
		}
	})

	It("Rejects a weekly schedule without an incremental backup day", func() {
		p := param.CreateBackupPolicy{
			BackupPolicyBase: param.BackupPolicyBase{
				ScheduleBase: param.ScheduleBase{
					ScheduleType: "Weekly",
					ScheduleDates: []param.ScheduleDate{{
						Day:        1,
						BackupType: "Full",
					}},
					ScheduleTime: "00:00",
				},
			},
		}

		_, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect(err).To(MatchError(ContainSubstring("At least one incremental backup day is required")))
	})

	It("Rejects a monthly schedule without an incremental backup day", func() {
		p := param.CreateBackupPolicy{
			BackupPolicyBase: param.BackupPolicyBase{
				ScheduleBase: param.ScheduleBase{
					ScheduleType: "Monthly",
					ScheduleDates: []param.ScheduleDate{{
						Day:        1,
						BackupType: "Full",
					}},
					ScheduleTime: "00:00",
				},
			},
		}

		_, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect(err).To(MatchError(ContainSubstring("At least one incremental backup day is required")))
	})

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
				DaysFieldBase: param.DaysFieldBase{
					JobKeepDays:  3,
					RecoveryDays: 7,
				},
			},
		}
		policy, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect(err).To(BeNil())
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
				DaysFieldBase: param.DaysFieldBase{
					JobKeepDays:  3,
					RecoveryDays: 7,
				},
			},
		}
		policy, err := buildBackupPolicyApiType(types.NamespacedName{Name: "t1", Namespace: "default"}, "fake-cluster", &p)
		Expect(err).To(BeNil())
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
