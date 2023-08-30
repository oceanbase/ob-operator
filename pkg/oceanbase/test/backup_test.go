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

package test

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Backup Operation", func() {

	var con *operation.OceanbaseOperationManager

	var _ = BeforeEach(func() {
		Expect(tenant).NotTo(BeEmpty())
		Expect(host).NotTo(BeEmpty())
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, password, database)
		con, err = operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		con.Logger = &logger
	})

	var _ = AfterEach(func() {
		Expect(con).NotTo(BeNil())
		err := con.Close()
		Expect(err).To(BeNil())
	})

	It("Query Clean Policies", func() {
		cleanPolicies, err := con.QueryBackupCleanPolicy()
		Expect(err).To(BeNil())
		printSlice(cleanPolicies, "clean policies")
	})

	It("Query ArchiveLogSummary", func() {
		summaries, err := con.QueryArchiveLogSummary()
		Expect(err).To(BeNil())
		printSlice(summaries, "archive log summary")
	})

	It("Query ArchiveLog", func() {
		logs, err := con.QueryArchiveLog()
		Expect(err).To(BeNil())
		printSlice(logs, "archive log")
	})

	It("Query ArchiveLogParameters", func() {
		parameters, err := con.QueryArchiveLogParameters()
		Expect(err).To(BeNil())
		printSlice(parameters, "archive log parameters")
	})

	It("Query BackupJobs", func() {
		jobs, err := con.QueryBackupJobs()
		Expect(err).To(BeNil())
		printSlice(jobs, "backup jobs")
	})

	It("Query BackupJobHistory", func() {
		histories, err := con.QueryBackupJobHistory()
		Expect(err).To(BeNil())
		printSlice(histories, "backup job history")
	})

	It("Query BackupCleanJobs", func() {
		histories, err := con.QueryBackupCleanJobs()
		Expect(err).To(BeNil())
		printSlice(histories, "backup clean jobs")
	})

	It("Query BackupCleanJobHistory", func() {
		histories, err := con.QueryBackupCleanHistory()
		Expect(err).To(BeNil())
		printSlice(histories, "backup clean jobs")
	})

	It("Query BackupTasks", func() {
		tasks, err := con.QueryBackupTasks()
		Expect(err).To(BeNil())
		printSlice(tasks, "backup tasks")
	})

	It("Query BackupTaskHistory", func() {
		histories, err := con.QueryBackupTaskHistory()
		Expect(err).To(BeNil())
		printSlice(histories, "backup task history")
	})

	It("Query BackupJob with correct ID", func() {
		job, err := con.QueryBackupJobWithId(3)
		Expect(err).To(BeNil())
		Expect(job).NotTo(BeNil())
		printObject(job, "BackupJob")
	})

	It("Query BackupJob with incorrect ID", func() {
		job, err := con.QueryBackupJobWithId(time.Now().Unix())
		Expect(err).To(BeNil())
		Expect(job).To(BeNil())
	})

	It("Create and return full type BackupJob", func() {
		Skip("This test will create a backup job, which will take a long time")
		var t v1alpha1.BackupJobType
		timeNow := time.Now().Unix()
		if timeNow%2 == 0 {
			t = v1alpha1.BackupJobTypeFull
		} else {
			t = v1alpha1.BackupJobTypeIncr
		}

		By("Create BackupJob of type " + string(t))
		job, err := con.CreateAndReturnBackupJob(t)
		Expect(err).To(BeNil())
		printObject(job, "BackupJob of type "+string(t))

		// Query tasks at once will get empty result
		time.Sleep(time.Second)
		By(fmt.Sprintf("Query BackupJob with ID %d", job.JobId))
		tasks, err := con.QueryBackupTaskWithJobId(job.JobId)
		Expect(err).To(BeNil())
		printSlice(tasks, fmt.Sprintf("BackupTasks of Job %d", job.JobId))
	})

	It("Get Log Archive dest info", func() {
		dest, err := con.QueryArchiveLogParameters()
		Expect(err).To(BeNil())
		printSlice(dest, "Log Archive dest info")
	})

	It("Set Log Archive parameter when LogMode == ARCHIVELOG", func() {
		By("Get tenant info")
		tenants, err := con.QueryTenantWithName(tenant)
		Expect(err).To(BeNil())
		Expect(len(tenants)).To(BeEquivalentTo(1))
		tenantInfo := tenants[0]

		if tenantInfo.LogMode != "ARCHIVELOG" {
			err = con.EnableArchiveLogForTenant()
			Expect(err).To(BeNil())
		}
		err = con.SetLogArchiveConcurrency(2)
		Expect(err).To(BeNil())

		latest, err := con.QueryLatestArchiveLogJob()
		Expect(err).To(BeNil())
		if latest != nil {
			if latest.Status != "DOING" {
				err = con.SetLogArchiveDestForTenant("location=file://ob-backup/" + tenant + "/log_archive")
				Expect(err).To(BeNil())
			}
		}
	})

	It("Set Log Archive parameter when LogMode == NOARCHIVELOG", func() {
		By("Get tenant info")
		tenants, err := con.QueryTenantWithName(tenant)
		Expect(err).To(BeNil())
		Expect(len(tenants)).To(BeEquivalentTo(1))
		tenantInfo := tenants[0]
		if tenantInfo.LogMode != "NOARCHIVELOG" {
			err = con.DisableArchiveLogForTenant()
			Expect(err).To(BeNil())
		}
		err = con.SetLogArchiveConcurrency(2)
		Expect(err).To(BeNil())

		latest, err := con.QueryLatestArchiveLogJob()
		Expect(err).To(BeNil())
		if latest != nil {
			if latest.Status != "DOING" {
				err = con.SetLogArchiveDestForTenant("location=file://ob-backup/" + tenant + "/log_archive")
				Expect(err).To(BeNil())
			}
		}
	})

	It("Set Log Archive concurrency to 0", func() {
		err := con.SetLogArchiveConcurrency(0)
		Expect(err).NotTo(BeNil())
	})

	It("Configure server for backup", func() {
		By("Get tenant info")
		tenants, err := con.QueryTenantWithName(tenant)
		Expect(err).To(BeNil())
		Expect(len(tenants)).To(BeEquivalentTo(1))
		tenantInfo := tenants[0]

		By("Set Log Archive Destination")
		if tenantInfo.LogMode == "NOARCHIVELOG" {

			latest, err := con.QueryLatestArchiveLogJob()
			Expect(err).To(BeNil())
			if latest != nil {
				if latest.Status != "DOING" {
					err = con.SetLogArchiveDestForTenant("location=file:///ob-backup/" + tenant + "/log_archive")
					Expect(err).To(BeNil())
				}
			}
		}

		By("Set Log Archive Concurrency")
		err = con.SetLogArchiveConcurrency(2)
		Expect(err).To(BeNil())

		latest, err := con.QueryLatestArchiveLogJob()
		Expect(err).To(BeNil())
		if latest != nil {
			if latest.Status != "DOING" {
				By("Set Data Backup Destination")
				err = con.SetDataBackupDestForTenant("file:///ob-backup/" + tenant + "/data_backup")
				Expect(err).To(BeNil())
			}
		}
	})

	It("Stop and restart archive log", func() {
		By("Get tenant info")
		tenants, err := con.QueryTenantWithName(tenant)
		Expect(err).To(BeNil())
		Expect(len(tenants)).To(BeEquivalentTo(1))
		tenantInfo := tenants[0]

		if tenantInfo.LogMode != "ARCHIVELOG" {
			err = con.EnableArchiveLogForTenant()
			Expect(err).To(BeNil())
		}
		By("Disable Archive Log")
		err = con.DisableArchiveLogForTenant()
		Expect(err).To(BeNil())
		time.Sleep(time.Millisecond * 500)

		By("Enable Archive Log")
		err = con.EnableArchiveLogForTenant()
		Expect(err).To(BeNil())
	})

	It("Stop Backup Job", func() {
		By("Stop full job")
		_, _ = con.CreateAndReturnBackupJob(v1alpha1.BackupJobTypeFull)
		// ignore error
		time.Sleep(time.Second)
		err := con.StopBackupJobOfTenant()
		Expect(err).To(BeNil())

		By("Stop incremental job")
		_, _ = con.CreateAndReturnBackupJob(v1alpha1.BackupJobTypeIncr)
		// ignore error
		time.Sleep(time.Second)
		err = con.StopBackupJobOfTenant()
		Expect(err).To(BeNil())
	})

	It("Create two backup jobs at the same time", Label("slow"), func() {
		_, _ = con.CreateAndReturnBackupJob(v1alpha1.BackupJobTypeIncr)
		time.Sleep(time.Second)

		_, err := con.CreateAndReturnBackupJob(v1alpha1.BackupJobTypeFull)
		Expect(err).NotTo(BeNil())
		time.Sleep(time.Second)

		err = con.StopBackupJobOfTenant()
		Expect(err).To(BeNil())
	})

	It("Set ARCHIVELOG and check", func() {
		By("Get tenant info")
		tenants, err := con.QueryTenantWithName(tenant)
		Expect(err).To(BeNil())
		Expect(len(tenants)).To(BeEquivalentTo(1))
		tenantInfo := tenants[0]

		if tenantInfo.LogMode != "ARCHIVELOG" {
			err = con.EnableArchiveLogForTenant()
			Expect(err).To(BeNil())
			By("Check ARCHIVELOG")
			tenants, err = con.QueryTenantWithName(tenant)
			Expect(err).To(BeNil())
			Expect(len(tenants)).To(BeEquivalentTo(1))
			tenantInfo = tenants[0]
			Expect(tenantInfo.LogMode).To(BeEquivalentTo("ARCHIVELOG"))
			err = con.DisableArchiveLogForTenant()
			Expect(err).To(BeNil())
		} else {
			err = con.DisableArchiveLogForTenant()
			Expect(err).To(BeNil())
			By("Check NOARCHIVELOG")
			tenants, err = con.QueryTenantWithName(tenant)
			Expect(err).To(BeNil())
			Expect(len(tenants)).To(BeEquivalentTo(1))
			tenantInfo = tenants[0]
			Expect(tenantInfo.LogMode).To(BeEquivalentTo("NOARCHIVELOG"))
			err = con.EnableArchiveLogForTenant()
			Expect(err).To(BeNil())
		}
	})
})
