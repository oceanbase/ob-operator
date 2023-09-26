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
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

var _ = Describe("Test System Operation", func() {
	var con *operation.OceanbaseOperationManager

	var _ = BeforeEach(func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, sysUser, "sys", sysPassword, database)
		con, err = operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		con.Logger = &logger
	})

	var _ = AfterEach(func() {
		Expect(con).NotTo(BeNil())
		err := con.Close()
		Expect(err).To(BeNil())
	})

	It("Query Tenants and Units", func() {
		By("Query tenants with name")
		tenants, err := con.ListTenantWithName(tenant)
		Expect(err).To(BeNil())
		printSlice(tenants, "tenants with name: "+tenant)
		if len(tenants) == 0 {
			Skip("no tenant found")
		}
		By("Query units with tenant id")
		units, err := con.ListUnitsWithTenantId(tenants[0].TenantID)
		Expect(err).To(BeNil())
		printSlice(units, "units with tenant id: "+fmt.Sprint(tenants[0].TenantID))
	})
})
