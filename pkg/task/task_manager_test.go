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

var _ = Describe("Test TaskManager", func() {
	It("Get TaskManager", func() {
		Expect(GetTaskManager()).ShouldNot(BeNil())
	})

	It("Submit successful task", func() {
		taskManager := GetTaskManager()
		taskId := taskManager.Submit(successfulTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := taskManager.GetTaskResult(taskId)
			return err == nil && result.Status == taskstatus.Successful
		}, 300, 3).Should(BeTrue())

		Expect(taskManager.CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit failed task", func() {
		taskManager := GetTaskManager()
		taskId := taskManager.Submit(failedTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := taskManager.GetTaskResult(taskId)
			return err != nil && result.Status == taskstatus.Failed
		}, 300, 3).Should(BeTrue())

		Expect(taskManager.CleanTaskResult(taskId)).Should(Succeed())
	})

	It("Submit panicky task", func() {
		taskManager := GetTaskManager()
		taskId := taskManager.Submit(panickyTask)
		Expect(taskId).ShouldNot(BeNil())

		Eventually(func() bool {
			result, err := taskManager.GetTaskResult(taskId)
			return err != nil && result.Status == taskstatus.Failed && strings.HasPrefix(err.Error(), "Observed a panic")
		}, 300, 3).Should(BeTrue())

		Expect(taskManager.CleanTaskResult(taskId)).Should(Succeed())
	})
})
