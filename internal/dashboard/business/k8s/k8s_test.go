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

package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
)

var _ = Describe("K8s", func() {
	It("Test ListEvents", func() {
		events, err := ListEvents(&param.QueryEventParam{
			ObjectType: "Pod",
			Type:       "Normal",
			Namespace:  "kube-system",
		})
		Expect(err).ShouldNot(HaveOccurred())
		Expect(events).ShouldNot(BeNil())
	})

	It("Test ListNamespaces", func() {
		namespaces, err := ListNamespaces()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(namespaces).ShouldNot(BeNil())
	})

	It("Test ListNodes", func() {
		nodes, err := ListNodes()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(nodes).ShouldNot(BeNil())
	})

	It("Test ListStorageClasses", func() {
		scs, err := ListStorageClasses()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(scs).ShouldNot(BeNil())
	})
})
