/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package ac

import (
	"context"
	"os"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
)

var _ = Describe("Role", Ordered, ContinueOnFailure, func() {
	It("Policies to CSV", func() {
		bts, err := os.ReadFile("rbac_policy.csv")
		Expect(err).To(BeNil())
		Expect(bts).ToNot(BeNil())
		csv, err := policiesToCsv()
		ginkgo.GinkgoLogr.Info("Policies to CSV", "csv", csv)
		Expect(err).To(BeNil())
		Expect(strings.TrimSpace(csv)).To(Equal(strings.TrimSpace(string(bts))))
	})

	It("Get all subjects", func() {
		subjects, err := enforcer.GetAllSubjects()
		Expect(err).To(BeNil())
		Expect(subjects).To(Equal([]string{"admin", "admin2"}))
	})

	It("Persist roles", func() {
		err := persistPolicies(context.Background(), "rbac_policy_persistent.csv")
		Expect(err).To(BeNil())
	})

	It("List roles", func() {
		roles, err := ListRoles(context.Background())
		GinkgoLogr.Info("List roles", "roles", roles)
		Expect(err).To(BeNil())
		Expect(roles).To(HaveLen(2))
		Expect(roles[0].Name).To(Equal("admin"))
		Expect(roles[0].Description).To(Equal("Super admin"))
		Expect(roles[0].Policies).To(HaveLen(1))
		Expect(roles[0].Policies[0].Domain).To(BeEquivalentTo("*"))
		Expect(roles[0].Policies[0].Object).To(BeEquivalentTo("*"))
		Expect(roles[0].Policies[0].Action).To(BeEquivalentTo("*"))

		Expect(roles[1].Name).To(Equal("admin2"))
		Expect(roles[1].Description).To(Equal("Book reader"))
		Expect(roles[1].Policies).To(HaveLen(1))
		Expect(roles[1].Policies[0].Domain).To(BeEquivalentTo("book"))
		Expect(roles[1].Policies[0].Object).To(BeEquivalentTo("*"))
		Expect(roles[1].Policies[0].Action).To(BeEquivalentTo("read"))
	})

	It("Create role", func() {
		createParam := &acmodel.CreateRoleParam{
			Name:        "test",
			Description: "test",
			Permissions: []acmodel.Policy{{
				Domain: "test",
				Object: "1",
				Action: "read",
			}, {
				Domain: "test",
				Object: "2",
				Action: "read",
			}, {
				Domain: "test2",
				Object: "*",
				Action: "*",
			}, {
				Domain: "test",
				Object: "3",
				Action: "write",
			}},
		}
		role, err := CreateRole(context.TODO(), createParam, "no-persist")
		Expect(err).To(BeNil())
		Expect(role).ToNot(BeNil())
		Expect(role.Name).To(Equal(createParam.Name))
		Expect(role.Description).To(Equal(createParam.Description))
		policyCsv, err := policiesToCsv()
		Expect(err).To(BeNil())
		Expect(strings.Contains(policyCsv, "p, test, test/1, read")).To(BeTrue())
		Expect(strings.Contains(policyCsv, "p, test, test/2, read")).To(BeTrue())
		Expect(strings.Contains(policyCsv, "p, test, test2/*, *")).To(BeTrue())
		Expect(strings.Contains(policyCsv, "p, test, test/3, write")).To(BeTrue())

		ok, err := enforcer.Enforce("test", "test/1", "read")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("test", "test/2", "read")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("test", "test/1", "write")
		Expect(err).To(BeNil())
		Expect(ok).To(BeFalse())

		ok, err = enforcer.Enforce("test", "test2/1", "read")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("test", "test/3", "read")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("test", "test/3", "write")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		Expect(err).To(BeNil())
		expectedCSV := `
p, admin, *, *, "Super admin"
p, admin2, book/*, read, "Book reader"
p, test, test/1, read, "test"
p, test, test/2, read, "test"
p, test, test2/*, *, "test"
p, test, test/3, write, "test"
g, admin, admin
g, admin, admin2
`
		actualCSV, err := policiesToCsv()
		Expect(err).To(BeNil())
		Expect(strings.TrimSpace(actualCSV)).To(Equal(strings.TrimSpace(expectedCSV)))
	})

	It("Delete role", func() {
		role, err := DeleteRole(context.Background(), "test", "no-persist")
		Expect(err).To(BeNil())

		Expect(role).ToNot(BeNil())
		Expect(role.Name).To(Equal("test"))
		Expect(role.Description).To(Equal("test"))
		Expect(role.Policies).To(HaveLen(4))

		expectedCSV := `
p, admin, *, *, "Super admin"
p, admin2, book/*, read, "Book reader"
g, admin, admin
g, admin, admin2
`
		actualCSV, err := policiesToCsv()
		Expect(err).To(BeNil())
		Expect(strings.TrimSpace(actualCSV)).To(Equal(strings.TrimSpace(expectedCSV)))
	})

	It("Key match", func() {
		Expect(keyMatch("test", "test")).To(BeTrue())
		Expect(keyMatch("test", "test2")).To(BeFalse())
		Expect(keyMatch("test", "test*")).To(BeTrue())
		Expect(keyMatch("test2", "test*")).To(BeFalse())
		Expect(keyMatch("test/", "test/1")).To(BeTrue())
		Expect(keyMatch("test/", "test/*")).To(BeTrue())
	})
})
