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
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/oceanbase/ob-operator/api/constants"
)

var _ = Describe("Test Miscellaneous Operation", func() {
	var _ = BeforeEach(func() {
	})

	var _ = AfterEach(func() {
	})

	It("Parse Timestamp", Label("time"), func() {
		timestamp := "2023-08-25 19:13:18.961907"
		t, err := time.Parse(time.DateTime+".000000", timestamp)
		Expect(err).To(BeNil())
		GinkgoWriter.Println(t, t.UnixMilli())

		t1, err := time.Parse(time.DateTime, timestamp)
		Expect(err).To(BeNil())
		Expect(t1.Equal(t)).To(BeTrue())

		ts2 := "2023-08-25 19:13:18"
		t2, err := time.Parse(time.DateTime, ts2)
		Expect(err).To(BeNil())
		Expect(t2.Equal(t)).NotTo(BeTrue())

		GinkgoWriter.Println(t, t2, t.UnixMicro()-t2.UnixMicro(), t.Sub(t2))
	})

	It("Parse time", Label("time"), func() {
		timePast, err := time.Parse(time.DateTime, "2023-08-30 17:10:08.064041")
		Expect(err).To(BeNil())
		Expect(timePast.Before(time.Now()))
	})

	It("Crontab Parse", Label("time"), func() {
		timeNow := time.Now()
		timePast, err := time.Parse(time.DateTime, "2023-08-30 17:10:08.064041")
		Expect(err).To(BeNil())
		Expect(timePast.Before(timeNow))
		sched, err := cron.ParseStandard("*/6 * * * *")
		Expect(err).To(BeNil())
		nextTime := sched.Next(time.Now())
		printObject(nextTime)
		Expect(timeNow.Before(nextTime)).To(BeTrue())
	})

	It("Backup Interval Pattern Match", Label("pattern"), func() {
		pattern := regexp.MustCompile(`^[1-7]d$`)
		Expect(pattern.MatchString("1d")).To(BeTrue())
		Expect(pattern.MatchString("7d")).To(BeTrue())
		Expect(pattern.MatchString("8d")).To(BeFalse())
		Expect(pattern.MatchString("0d")).To(BeFalse())
		Expect(pattern.MatchString("d")).To(BeFalse())
		Expect(pattern.MatchString("dd")).To(BeFalse())
		Expect(pattern.MatchString("1")).To(BeFalse())
		Expect(pattern.MatchString("1dd")).To(BeFalse())
		Expect(pattern.MatchString("12h")).To(BeFalse())
		Expect(pattern.MatchString("1d1")).To(BeFalse())
	})

	It("Recovery Window Pattern Match", Label("pattern"), func() {
		pattern := regexp.MustCompile(`^[1-9]\d*d$`)
		Expect(pattern.MatchString("2d")).To(BeTrue())
		Expect(pattern.MatchString("10d")).To(BeTrue())
		Expect(pattern.MatchString("1d")).To(BeTrue())
		Expect(pattern.MatchString("30d")).To(BeTrue())
		Expect(pattern.MatchString("100d")).To(BeTrue())
		Expect(pattern.MatchString("1")).To(BeFalse())
		Expect(pattern.MatchString("1dd")).To(BeFalse())
		Expect(pattern.MatchString("12h")).To(BeFalse())
		Expect(pattern.MatchString("1d1")).To(BeFalse())
		Expect(pattern.MatchString("0d")).To(BeFalse())
		Expect(pattern.MatchString("d")).To(BeFalse())
		Expect(pattern.MatchString("dd")).To(BeFalse())
	})

	It("Parse label selectors", Label("label"), func() {
		_, err := labels.Parse("open.oceanbase.com/backup-job-status!=RUNNING")
		Expect(err).To(BeNil())
		_, err = labels.Parse("open.oceanbase.com/backup-job-status!=RUNNING,open.oceanbase.com/backup-job-status!=FAILED")
		Expect(err).To(BeNil())
		_, err = labels.Parse("open.oceanbase.com/backup-job-status!=RUNNING,open.oceanbase.com/backup-job-status!=FAILED,open.oceanbase.com/backup-job-status!=CANCELED")
		Expect(err).To(BeNil())
	})

	It("Parse field selectors", Label("field"), func() {
		_, err := labels.Parse("status.phase!=Running")
		Expect(err).To(BeNil())
		_, err = labels.Parse("status.phase!=Running,status.phase!=Failed")
		Expect(err).To(BeNil())
		_, err = labels.Parse("status.phase!=Running,status.phase!=Failed,status.phase!=Canceled")
		Expect(err).To(BeNil())
		_, err = fields.ParseSelector(".status.status!=" + string(constants.BackupJobStatusSuccessful) + ",.status.status!=" + string(constants.BackupJobStatusFailed) + ",.status.status!=" + string(constants.BackupJobStatusCanceled))
		Expect(err).To(BeNil())
		_, err = fields.ParseSelector("status.status!=" + string(constants.BackupJobStatusSuccessful) + ",status.status!=" + string(constants.BackupJobStatusFailed) + ",status.status!=" + string(constants.BackupJobStatusCanceled))
		Expect(err).To(BeNil())
	})

	It("Parse time window", Label("time"), func() {
		pattern := regexp.MustCompile(`^[1-9]\d*d$`)
		Expect(pattern.MatchString("2d")).To(BeTrue())
		keepWindowDays, err := strconv.Atoi(strings.TrimRight("200d", "d"))
		Expect(err).To(BeNil())
		keepWindow := time.Duration(keepWindowDays*24) * time.Hour
		printObject(keepWindow, "keepWindow")
		TwoHundredDays, err := time.ParseDuration("4800h")
		Expect(err).To(BeNil())
		Expect(keepWindow).To(Equal(TwoHundredDays))
	})
})
