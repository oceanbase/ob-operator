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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Test validations", Label("validation"), func() {
	It("Test quantity parse", func() {
		overflow := resource.MustParse("15Gi")
		threshold := resource.MustParse("10Gi")
		Expect(overflow.AsApproximateFloat64() > threshold.AsApproximateFloat64()).To(BeTrue())
	})

	It("Test quantity parse 2", func() {
		overflow := resource.MustParse("16106127360")
		threshold := resource.MustParse("10Gi")
		Expect(overflow.AsApproximateFloat64() > threshold.AsApproximateFloat64()).To(BeTrue())
	})
})
