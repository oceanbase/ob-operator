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
	"strings"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

var _ = Describe("Test Restore Operation", Serial, Label("restore"), func() {
	var con *operation.OceanbaseOperationManager
	var standbyName string
	var _ = BeforeEach(func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, sysUser, "sys", sysPassword, database)
		con, err = operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		con.Logger = &logger
		standbyName = tenant + "_standby"
	})

	var _ = AfterEach(func() {
		Expect(con).NotTo(BeNil())
		err := con.Close()
		Expect(err).To(BeNil())
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

	It("Write some data to tenant", Label("prepare_data"), func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, "", "")
		tenantCon, err := operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		tenantCon.Logger = &logger

		By("Write some data to tenant")
		err = tenantCon.ExecWithDefaultTimeout("create table if not exists test.test (id int, name varchar(20))")
		Expect(err).To(BeNil())
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (1, 'test')")
		Expect(err).To(BeNil())
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (2, 'test')")
		Expect(err).To(BeNil())
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (3, 'test'), (4, 'test'), (5, 'test')")
		Expect(err).To(BeNil())

		Expect(tenantCon.Close()).To(BeNil())
	})

	It("Backup primary tenant", Label("prepare_backup"), func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, "", database)
		tenantCon, err := operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		tenantCon.Logger = &logger
		backupDest := "file:///ob-backup/" + tenant + "/backup"
		archiveDest := "location=file:///ob-backup/" + tenant + "/archive"

		By("Backup primary tenant")
		err = tenantCon.SetLogArchiveDestForTenant(archiveDest)
		Expect(err).To(BeNil())

		By("Set archive log retention")
		err = tenantCon.EnableArchiveLogForTenant()
		Expect(err).To(BeNil())

		By("Wait for archive doing")
		for {
			time.Sleep(5 * time.Second)
			latestArchive, err := tenantCon.GetLatestArchiveLogJob()
			Expect(err).To(BeNil())
			if latestArchive != nil && latestArchive.Status == "DOING" {
				break
			}
		}

		By("Set backup destination")
		err = tenantCon.SetDataBackupDestForTenant(backupDest)
		Expect(err).To(BeNil())

		By("Create full backup job")
		err = tenantCon.CreateBackupFull()
		Expect(err).To(BeNil())

		By("Wait for backup done")
		for {
			time.Sleep(3 * time.Second)
			backupJob, err := tenantCon.GetLatestBackupJobOfTypeAndPath(constants.BackupJobTypeFull, backupDest)
			Expect(err).To(BeNil())
			if backupJob != nil && backupJob.Status == "COMPLETED" {
				break
			}
		}
		By("Finish backup of primary tenant")
	})

	It("Write some data to tenant", Label("prepare_data2"), func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, "", "")
		tenantCon, err := operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		tenantCon.Logger = &logger

		By("Write some data to tenant")
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (101, 'test_after')")
		Expect(err).To(BeNil())
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (102, 'test_after')")
		Expect(err).To(BeNil())
		err = tenantCon.ExecWithDefaultTimeout("insert into test.test values (103, 'test_after')")
		Expect(err).To(BeNil())
		By("Wait for a moment")
		time.Sleep(10 * time.Second)
		Expect(tenantCon.Close()).To(BeNil())
	})

	It("Checking backup progress", Label("check_backup"), func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, user, tenant, "", database)
		tenantCon, err := operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		tenantCon.Logger = &logger
		backupDest := "file:///ob-backup/" + tenant + "/backup"

		By("Wait for backup done")
		for {
			time.Sleep(3 * time.Second)
			backupJob, err := tenantCon.GetLatestBackupJobOfTypeAndPath(constants.BackupJobTypeFull, backupDest)
			Expect(err).To(BeNil())
			if backupJob != nil && backupJob.Status == "COMPLETED" {
				break
			}
		}
		By("Finish backup of primary tenant")
	})

	It("Wait for 3 mins", func() {
		// Avoid the case that the standby tenant is created too fast
		// ERROR 4018 (HY000): No enough log to restore
		time.Sleep(1 * time.Minute)
	})

	It("Restore standby tenant", Label("restore_standby"), func() {
		By("Check target tenant's existence")
		exists, err := con.CheckTenantExistByName(standbyName)
		Expect(err).To(BeNil())
		if exists {
			Skip("Target standby tenant exists")
		}

		By("Create resource pool")
		poolList, err := con.GetPoolList()
		Expect(err).To(BeNil())
		exists = false
		for _, pool := range poolList {
			if pool.Name == "pool_test_standby1" {
				exists = true
				break
			}
		}
		if !exists {
			for _, v := range []int{1, 2, 3} {
				err = con.AddPool(model.PoolSQLParam{
					UnitNum:  1,
					PoolName: fmt.Sprintf("pool_test_standby%d", v),
					ZoneList: fmt.Sprintf("zone%d", v),
					UnitName: "unit_test",
				})
				Expect(err).To(BeNil())
			}
		}

		By("Trigger restoration of standby tenant")
		backupDest := "file:///ob-backup/" + tenant + "/backup"
		archiveDest := "file:///ob-backup/" + tenant + "/archive"
		err = con.StartRestoreUnlimited(standbyName, strings.Join([]string{backupDest, archiveDest}, ","), "pool_list=pool_test_standby1,pool_test_standby2,pool_test_standby3")
		Expect(err).To(BeNil())
	})

	It("Query restore progress", Label("query_restore"), func() {

		By("Check restoration progress")
		for {
			time.Sleep(5 * time.Second)
			restoreJob, err := con.GetLatestRestoreProgressOfTenant(standbyName)
			Expect(err).To(BeNil())
			printObject(restoreJob, "restoreJob")
			if restoreJob != nil && restoreJob.Status == "SUCCESS" {
				break
			}
			if restoreJob == nil {
				restoreHistory, err := con.GetLatestRestoreHistoryOfTenant(standbyName)
				Expect(err).To(BeNil())
				if restoreHistory != nil {
					printObject(restoreHistory, "restoreHistory")
					if restoreHistory.Status == "SUCCESS" {
						break
					}
				}
			}
		}
		By("Restore finished")
	})

	It("Cancel restoring", Label("cancel_restore"), func() {
		Skip("")
		Expect(con.CancelCleanBackup()).To(BeNil())
	})

	It("Replay", Label("replay"), func() {
		// Not repeatable
		err := con.ReplayStandbyLog(standbyName, "UNLIMITED")
		Expect(err).To(BeNil())
		time.Sleep(3 * time.Second)
	})

	It("Activate", Label("activate"), func() {
		// Repeatable
		err := con.ActivateStandby(standbyName)
		Expect(err).To(BeNil())
		time.Sleep(3 * time.Second)
	})

	It("Delete Tenants", Label("delete_tenants"), func() {
		By("Deleting primary tenant")
		exists, err := con.CheckTenantExistByName(tenant)
		Expect(err).To(BeNil())
		if exists {
			Expect(con.DeleteTenant(tenant, true)).To(BeNil())
		}
		By("Deleting standby tenants")
		exists, err = con.CheckTenantExistByName(standbyName)
		Expect(err).To(BeNil())
		if exists {
			Expect(con.DeleteTenant(standbyName, true)).To(BeNil())
		}
		By("Deleting resource pools")
		for _, pool := range []string{"pool_test1", "pool_test2", "pool_test3", "pool_test_standby1", "pool_test_standby2", "pool_test_standby3"} {
			exists, err = con.CheckPoolExistByName(pool)
			Expect(err).To(BeNil())
			if exists {
				Expect(con.DeletePool(pool)).To(BeNil())
			}
		}
	})
})

var _ = Describe("Test canceling restore", Serial, Label("canceling"), func() {
	var con *operation.OceanbaseOperationManager
	var standbyName string
	var _ = BeforeEach(func() {
		var err error
		logger := logr.Discard()
		ds := connector.NewOceanBaseDataSource(host, port, sysUser, "sys", sysPassword, database)
		con, err = operation.GetOceanbaseOperationManager(ds)
		Expect(err).To(BeNil())
		con.Logger = &logger
		standbyName = tenant + "_standby"
	})
	It("Create units", func() {
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
				MemorySize:     2147483648,
				MaxIops:        1024,
				LogDiskSize:    2147483648,
				MinIops:        1024,
			})
			Expect(err).To(BeNil())
		}

	})
	It("Start and cancel the restore", func() {
		By("Check target tenant's existence")
		exists, err := con.CheckTenantExistByName(standbyName)
		Expect(err).To(BeNil())
		if exists {
			Skip("Target standby tenant exists")
		}

		By("Create resource pool")
		poolList, err := con.GetPoolList()
		Expect(err).To(BeNil())
		exists = false
		for _, pool := range poolList {
			if pool.Name == "pool_test_standby1" {
				exists = true
				break
			}
		}
		if !exists {
			for _, v := range []int{1, 2, 3} {
				err = con.AddPool(model.PoolSQLParam{
					UnitNum:  1,
					PoolName: fmt.Sprintf("pool_test_standby%d", v),
					ZoneList: fmt.Sprintf("zone%d", v),
					UnitName: "unit_test",
				})
				Expect(err).To(BeNil())
			}
		}

		By("Trigger restoration of standby tenant")
		backupDest := "file:///ob-backup/" + tenant + "/data_backup_custom1"
		archiveDest := "file:///ob-backup/" + tenant + "/log_archive_custom1"
		err = con.StartRestoreUnlimited(standbyName, strings.Join([]string{backupDest, archiveDest}, ","), "pool_list=pool_test_standby1,pool_test_standby2,pool_test_standby3")
		Expect(err).To(BeNil())
		time.Sleep(5 * time.Second)

		By("Cancel restoration of tenant")
		err = con.CancelRestoreOfTenant(standbyName)
		Expect(err).To(BeNil())
	})

	It("Delete Tenants", Label("delete_tenants"), func() {
		By("Deleting primary tenant")
		exists, err := con.CheckTenantExistByName(tenant)
		Expect(err).To(BeNil())
		if exists {
			Expect(con.DeleteTenant(tenant, true)).To(BeNil())
		}

		By("Deleting standby tenants")
		exists, err = con.CheckTenantExistByName(standbyName)
		Expect(err).To(BeNil())
		if exists {
			Expect(con.DeleteTenant(standbyName, true)).To(BeNil())
		}

		By("Deleting resource pools")
		for _, pool := range []string{"pool_test1", "pool_test2", "pool_test3", "pool_test_standby1", "pool_test_standby2", "pool_test_standby3"} {
			exists, err = con.CheckPoolExistByName(pool)
			Expect(err).To(BeNil())
			if exists {
				Expect(con.DeletePool(pool)).To(BeNil())
			}
		}
	})
})
