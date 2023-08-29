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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Miscellaneous Operation", func() {
	var _ = BeforeEach(func() {
	})

	var _ = AfterEach(func() {
	})

	It("Parse Timestamp", func() {
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
})
