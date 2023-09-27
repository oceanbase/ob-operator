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
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

var _ = Describe("Test Tenant Operation", func() {
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

var _ = Describe("Test Tenant Operation", Label("tenant-operation"), Serial, func() {
	var con *operation.OceanbaseOperationManager

	var _ = BeforeEach(func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, sysUser, "sys", sysPassword, database)
		con, err = operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		con.Logger = &logger
	})

	It("Create Unit, Resource pool and Tenants", func() {
		By("Check if tenant exists")
		tenants, err := con.ListTenantWithName(tenant)
		Expect(err).To(BeNil())
		if len(tenants) > 0 {
			Skip("tenant already exists")
		}
		By("Create unit")
		unitList, err := con.GetUnitConfigV4List()
		Expect(err).To(BeNil())
		exists := false
		for _, unit := range unitList {
			if unit.Name == "unit_test" {
				exists = true
				break
			}
		}
		if !exists {
			err = con.AddUnitConfigV4(&model.UnitConfigV4SQLParam{
				UnitConfigName: "unit_test",
				MinCPU:         2,
				MaxCPU:         2,
				MemorySize:     2048000000,
				MaxIops:        1024,
				LogDiskSize:    2048000000,
				MinIops:        1024,
			})
			Expect(err).To(BeNil())
		}
		By("Create resource pool")
		poolList, err := con.GetPoolList()
		Expect(err).To(BeNil())
		exists = false
		for _, pool := range poolList {
			if pool.Name == "pool_test1" {
				exists = true
				break
			}
		}
		if !exists {
			for _, v := range []int{1, 2, 3} {
				err = con.AddPool(model.PoolSQLParam{
					UnitNum:  1,
					PoolName: fmt.Sprintf("pool_test%d", v),
					ZoneList: fmt.Sprintf("zone%d", v),
					UnitName: "unit_test",
				})
				Expect(err).To(BeNil())
			}
		}

		By("Create tenant")
		exists, err = con.CheckTenantExistByName(tenant)
		Expect(err).To(BeNil())
		if !exists {
			err = con.AddTenant(model.TenantSQLParam{
				TenantName:   tenant,
				Charset:      "utf8mb4",
				PrimaryZone:  "zone1",
				PoolList:     []string{"pool_test1", "pool_test2", "pool_test3"},
				UnitNum:      1,
				VariableList: "ob_tcp_invited_nodes='%'",
			})
			Expect(err).To(BeNil())
		}
	})

	It("Modify password tenant's root and login root once again", func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, "", "")
		tenantCon, err := operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		tenantCon.Logger = &logger

		newPwd := "testPwd"
		err = tenantCon.ChangeTenantUserPassword("root", newPwd)
		Expect(err).To(BeNil())
		newDs := connector.NewOceanBaseDataSource(host, port, user, tenant, newPwd, "")
		newCon, err := operation.GetOceanbaseOperationManager(newDs)
		Expect(err).To(BeNil())
		newCon.Logger = &logger
		_, err = newCon.ListArchiveLogSummary()
		Expect(err).To(BeNil())
	})

	It("Delete Tenants", Label("delete_tenants"), func() {
		By("Deleting primary tenant")
		exists, err := con.CheckTenantExistByName(tenant)
		Expect(err).To(BeNil())
		if exists {
			Expect(con.DeleteTenant(tenant, true)).To(BeNil())
		}
		By("Deleting resource pools")
		for _, pool := range []string{"pool_test1", "pool_test2", "pool_test3"} {
			exists, err = con.CheckPoolExistByName(pool)
			Expect(err).To(BeNil())
			if exists {
				Expect(con.DeletePool(pool)).To(BeNil())
			}
		}
	})
})
