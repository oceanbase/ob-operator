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
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

var _ = Describe("Test Backup Operation", Label("system"), func() {

	var con *operation.OceanbaseOperationManager

	var _ = BeforeEach(func() {
		Expect(host).NotTo(BeEmpty())
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

	It("List units with tenant ID", func() {
		units, err := con.ListUnitsWithTenantId(1)
		Expect(err).To(BeNil())
		printSlice(units, "list units with tenant ID")
	})

	It("List units with server IP", func() {
		units, err := con.ListUnitsWithServerIP("10.42.0.99")
		Expect(err).To(BeNil())
		printSlice(units, "list units with server IP")
	})
})
