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
	"crypto/sha256"
	"encoding/hex"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Access Control", Ordered, ContinueOnFailure, func() {
	It("GetFilteredPolicies", func() {
		roles, err := enforcer.GetFilteredPolicy(0, "admin")
		Expect(err).To(BeNil())
		Expect(roles).To(HaveLen(1))
		GinkgoLogr.Info("roles", "roles", roles)
	})

	It("GetPolicies", func() {
		ps, err := enforcer.GetPolicy()
		Expect(err).To(BeNil())
		Expect(ps).To(HaveLen(2))
		GinkgoLogr.Info("Policies", "policies", ps)
	})

	It("GetAllRoles", func() {
		roles, err := enforcer.GetAllRoles()
		Expect(err).To(BeNil())
		Expect(roles).To(Equal([]string{"admin", "admin2"}))
	})

	It("GetAccountInfo", func() {
		roles, err := enforcer.GetRolesForUser("admin")
		Expect(err).To(BeNil())
		slices.Sort(roles)
		Expect(roles).To(Equal([]string{"admin", "admin2"}))
	})

	It("Enforce some permissions", func() {
		ok, err := enforcer.Enforce("admin", "dashboard", "view")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin", "dashboard", "edit")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin", "*", "*")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin", "book/*", "*")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin", "book/2", "*")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin2", "dashboard", "view")
		Expect(err).To(BeNil())
		Expect(ok).To(BeFalse())

		ok, err = enforcer.Enforce("admin2", "book/*", "READ")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())

		ok, err = enforcer.Enforce("admin2", "book/2", "READ")
		Expect(err).To(BeNil())
		Expect(ok).To(BeTrue())
	})

	It("Validate password", func() {
		// password is "hello"
		hash := sha256.Sum256([]byte("hello"))
		Expect(hex.EncodeToString(hash[:])).To(Equal("2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"))
	})
})
