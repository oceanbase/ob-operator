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

package task

import (
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func successfulTask() tasktypes.TaskError {
	time.Sleep(time.Second)
	return nil
}

func failedTask() tasktypes.TaskError {
	return errors.New("failed task")
}

func panickyTask() tasktypes.TaskError {
	panic("panicky task")
}

func longrunTask() tasktypes.TaskError {
	time.Sleep(3 * time.Second)
	return nil
}

var _ = Describe("Test TaskManager", Serial, func() {
	It("Get TaskManager", func() {
		Expect(GetTaskManager()).ShouldNot(BeNil())
	})

	It("Submit successful task", func() {
		taskId := GetTaskManager().Submit(successfulTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := GetTaskManager().GetTaskResult(taskId)
			return err == nil && result != nil && result.Status == taskstatus.Successful
		}, 3, 1).Should(BeTrue())

		Expect(GetTaskManager().CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit failed task", func() {
		taskId := GetTaskManager().Submit(failedTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			defer GinkgoRecover()
			result, err := GetTaskManager().GetTaskResult(taskId)
			return err == nil && result != nil && result.Status == taskstatus.Failed
		}, 3, 1).Should(BeTrue())

		Expect(GetTaskManager().CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit panicky task", func() {
		taskId := GetTaskManager().Submit(panickyTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := GetTaskManager().GetTaskResult(taskId)
			return err == nil && result != nil && result.Status == taskstatus.Failed && strings.HasPrefix(result.Error.Error(), "Observed a panic")
		}, 3, 1).Should(BeTrue())

		Expect(GetTaskManager().CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit long run task", func() {
		taskId := GetTaskManager().Submit(longrunTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := GetTaskManager().GetTaskResult(taskId)
			return err == nil && result != nil && result.Status == taskstatus.Successful
		}, 8, 1).Should(BeTrue())

		Expect(GetTaskManager().CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit 1e6 tasks", Label("long-run"), func() {
		// 1e6 unit tasks take about 0.9G memory
		for i := 0; i < 1e6; i++ {
			taskId := GetTaskManager().Submit(longrunTask)
			Expect(taskId).ShouldNot(BeNil())
		}
	})
})
