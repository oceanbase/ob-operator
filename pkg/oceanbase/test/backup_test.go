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
		time.Sleep(2 * time.Second)
		By(fmt.Sprintf("Query BackupJob with ID %d", job.JobId))
		tasks, err := con.QueryBackupTaskWithJobId(job.JobId)
		Expect(err).To(BeNil())
		printSlice(tasks, fmt.Sprintf("BackupTasks of Job %d", job.JobId))
	})
})
