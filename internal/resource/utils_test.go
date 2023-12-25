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

package resource

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/internal/resource/utils"
)

var _ = Describe("Test utils", func() {
	It("Min", func() {
		Expect(utils.Min(1, 2)).Should(Equal(1))
		Expect(utils.Min(2, 1)).Should(Equal(1))
		Expect(utils.Min(1.2, 2.1)).Should(Equal(1.2))
		Expect(utils.Min(2.1, 1.2)).Should(Equal(1.2))
	})

	It("IsZero", func() {
		Expect(utils.IsZero(0)).Should(BeTrue())
		Expect(utils.IsZero(1)).Should(BeFalse())
		Expect(utils.IsZero(0.0)).Should(BeTrue())
		Expect(utils.IsZero(1.0)).Should(BeFalse())
		Expect(utils.IsZero("")).Should(BeTrue())
		Expect(utils.IsZero("a")).Should(BeFalse())
		Expect(utils.IsZero(&v1.Secret{})).Should(BeFalse())
	})
})
