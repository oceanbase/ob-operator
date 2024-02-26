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
