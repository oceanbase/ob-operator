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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func flow1() *tasktypes.TaskFlow {
	return nil
}

func flow2() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{},
	}
}

var _ = Describe("Test Task Registry", Serial, func() {

	It("Test Register", func() {
		Expect(GetRegistry()).ShouldNot(BeNil())
		GetRegistry().Register("flow1", flow1)
		GetRegistry().Register("flow2", flow2)
	})

	It("Test Get", func() {
		f1, err := GetRegistry().Get("flow1")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(f1).Should(BeNil())

		f2, err := GetRegistry().Get("flow2")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(f2).ShouldNot(BeNil())

		_, err = GetRegistry().Get("flow3")
		Expect(err).Should(HaveOccurred())
	})
})
