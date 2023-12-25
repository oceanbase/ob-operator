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
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

const poolSizeEnv = "TASK_POOL_SIZE"
const taskMemoryUsageEnv = "TASK_MEMORY_USAGE"
const taskNumEnv = "TASK_NUM"
const taskSleepEnv = "TASK_SLEEP"

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

	It("Submit 1e6 tasks", Label("long-run", "sum"), func() {
		// 1e6 unit tasks take about 2.1G memory
		for i := 0; i < 1e6; i++ {
			taskId := GetTaskManager().Submit(longrunTask)
			Expect(taskId).ShouldNot(BeNil())
		}
	})

	It("Submit tasks and clean", Label("long-run"), func() {
		var poolSize int
		var taskMemory int
		var taskNum int
		var taskSleep int

		if os.Getenv(poolSizeEnv) != "" {
			tmpPoolSize, err := strconv.Atoi(os.Getenv(poolSizeEnv))
			if err == nil {
				poolSize = tmpPoolSize
			}
		}
		if poolSize == 0 {
			poolSize = 100
		}

		if os.Getenv(taskMemoryUsageEnv) != "" {
			tmpMemoryUsage, err := strconv.Atoi(os.Getenv(taskMemoryUsageEnv))
			if err == nil {
				taskMemory = tmpMemoryUsage
			}
		}
		if taskMemory == 0 {
			taskMemory = 20
		}

		if os.Getenv(taskNumEnv) != "" {
			tmpTaskNum, err := strconv.Atoi(os.Getenv(taskNumEnv))
			if err == nil {
				taskNum = tmpTaskNum
			}
		}
		if taskNum == 0 {
			taskNum = 500
		}

		if os.Getenv(taskSleepEnv) != "" {
			tmpTaskSleep, err := strconv.Atoi(os.Getenv(taskSleepEnv))
			if err == nil {
				taskSleep = tmpTaskSleep
			}
		}
		if taskSleep == 0 {
			taskSleep = 10
		}
		defer GinkgoRecover()
		tm := &TaskManager{
			ResultMap:       sync.Map{},
			TaskResultCache: sync.Map{},
			tokens:          make(chan struct{}, poolSize),
		}
		wg := sync.WaitGroup{}
		for i := 0; i < taskNum; i++ {
			wg.Add(1)
			sleepTime := rand.Intn(taskSleep) + 1
			taskId := tm.Submit(func() tasktypes.TaskError {
				bts := make([]byte, 1<<taskMemory, 1<<taskMemory)
				bts[0] = 0
				time.Sleep(time.Duration(sleepTime) * time.Second)
				return nil
			})
			go func(i int) {
				for {
					time.Sleep(500*time.Millisecond + time.Duration(rand.Intn(1000))*time.Millisecond)
					result, err := tm.GetTaskResult(taskId)
					if err == nil && result != nil && result.Error == nil && (result.Status == taskstatus.Successful || result.Status == taskstatus.Failed) {
						tm.CleanTaskResult(taskId)
						wg.Done()
						return
					}
				}
			}(i)
		}
		wg.Wait()
	})
})
