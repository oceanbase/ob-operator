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

	apiconsts "github.com/oceanbase/ob-operator/api/constants"
)

var _ = Describe("Test OBTenantOperation Webhook", Label("webhook"), Serial, func() {
	clusterName := "test-cluster-for-operation"
	tenantPrimary := "test-tenant-for-operation"
	tenantStandby := "test-tenant-for-operation2"

	It("Create cluster and tenants", func() {
		c := newOBCluster(clusterName, 3, 1)
		t := newOBTenant(tenantPrimary, clusterName)
		t2 := newOBTenant(tenantStandby, clusterName)
		t2.Spec.TenantRole = apiconsts.TenantRoleStandby
		t2.Spec.Source = &TenantSourceSpec{
			Tenant: &tenantPrimary,
		}
		Expect(k8sClient.Create(ctx, c)).Should(Succeed())
		Expect(k8sClient.Create(ctx, t)).Should(Succeed())
		Expect(k8sClient.Create(ctx, t2)).Should(Succeed())

		t.Status.TenantRole = apiconsts.TenantRolePrimary
		t.Status.Pools = []ResourcePoolStatus{}
		Expect(k8sClient.Status().Update(ctx, t)).Should(Succeed())
	})

	It("Check operation types", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = "illegal-operation-type"
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
	})

	It("Check operation change password", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpChangePwd

		op.Spec.ChangePwd = &OBTenantOpChangePwdSpec{
			Tenant:    tenantPrimary,
			SecretRef: "",
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.ChangePwd.SecretRef = wrongKeySecret
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.ChangePwd = &OBTenantOpChangePwdSpec{
			Tenant:    "",
			SecretRef: defaultSecretName,
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.ChangePwd = &OBTenantOpChangePwdSpec{
			Tenant:    "tenant-not-exist",
			SecretRef: defaultSecretName,
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.ChangePwd = &OBTenantOpChangePwdSpec{
			Tenant:    tenantPrimary,
			SecretRef: "secret-not-exist",
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
	})

	It("Check operation failover", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpFailover
		op.Spec.Failover = nil
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.Failover = &OBTenantOpFailoverSpec{}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
		op.Spec.Failover = &OBTenantOpFailoverSpec{
			StandbyTenant: "tenant-not-exist",
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
	})

	It("Check operation switchover", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpSwitchover

		op.Spec.Switchover = nil
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Switchover = &OBTenantOpSwitchoverSpec{}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Switchover.PrimaryTenant = "tenant-not-exist"
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Switchover.StandbyTenant = "tenant-not-exist"
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Switchover.PrimaryTenant = tenantStandby
		op.Spec.Switchover.StandbyTenant = tenantPrimary
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Switchover.PrimaryTenant = tenantPrimary
		op.Spec.Switchover.StandbyTenant = tenantStandby
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
	})

	It("Check operation upgrade", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpUpgrade

		op.Spec.TargetTenant = nil
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		notexist := "tenant-not-exist"
		op.Spec.TargetTenant = &notexist
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.TargetTenant = &tenantPrimary
		Expect(k8sClient.Create(ctx, op)).Should(Succeed())
	})

	It("Check operation replay log", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpReplayLog
		op.Spec.ReplayUntil = &RestoreUntilConfig{
			Unlimited: true,
		}

		op.Spec.TargetTenant = nil
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		notexist := "tenant-not-exist"
		op.Spec.TargetTenant = &notexist
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.TargetTenant = &tenantPrimary
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.TargetTenant = &tenantStandby
		op.Spec.ReplayUntil = &RestoreUntilConfig{
			Unlimited: false,
		}
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())
	})

	It("Check adding resource pools", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpAddResourcePools
		op.Spec.AddResourcePools = []ResourcePoolSpec{{
			Zone: "zone1",
			Type: &LocalityType{
				Name:     "Full",
				Replica:  1,
				IsActive: true,
			},
			UnitConfig: &UnitConfig{
				MaxCPU:      resource.MustParse("1"),
				MemorySize:  resource.MustParse("5Gi"),
				MinCPU:      resource.MustParse("1"),
				MaxIops:     1024,
				MinIops:     1024,
				IopsWeight:  2,
				LogDiskSize: resource.MustParse("12Gi"),
			},
		}}

		notexist := "tenant-not-exist"
		op.Spec.TargetTenant = &notexist

		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.TargetTenant = &tenantPrimary
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Force = true
		Expect(k8sClient.Create(ctx, op)).Should(Succeed())

		// Delete resource pool
		opDel := newTenantOperation(tenantPrimary)
		opDel.Spec.Type = apiconsts.TenantOpDeleteResourcePools
		opDel.Spec.DeleteResourcePools = []string{"zone0"}
		opDel.Spec.TargetTenant = &tenantPrimary
		Expect(k8sClient.Create(ctx, opDel)).ShouldNot(Succeed())
		opDel.Spec.Force = true
		Expect(k8sClient.Create(ctx, opDel)).Should(Succeed())
	})

	It("Check modifying resource pools", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpModifyResourcePools
		op.Spec.ModifyResourcePools = []ResourcePoolSpec{{
			Zone: "zone0",
			Type: &LocalityType{
				Name:     "Full",
				Replica:  1,
				IsActive: true,
			},
			UnitConfig: &UnitConfig{
				MaxCPU:      resource.MustParse("6"),
				MemorySize:  resource.MustParse("6Gi"),
				MinCPU:      resource.MustParse("2"),
				MaxIops:     1024,
				MinIops:     1024,
				IopsWeight:  2,
				LogDiskSize: resource.MustParse("12Gi"),
			},
		}}

		op.Spec.TargetTenant = &tenantPrimary
		Expect(k8sClient.Create(ctx, op)).ShouldNot(Succeed())

		op.Spec.Force = true
		Expect(k8sClient.Create(ctx, op)).Should(Succeed())
	})

	It("Check setting connection white list", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpSetConnectWhiteList
		op.Spec.ConnectWhiteList = "%,127.0.0.1"
		op.Spec.Force = true
		Expect(k8sClient.Create(ctx, op)).Should(Succeed())
	})

	It("Check setting unit number", func() {
		op := newTenantOperation(tenantPrimary)
		op.Spec.Type = apiconsts.TenantOpSetUnitNumber
		op.Spec.UnitNumber = 2
		op.Spec.Force = true
		Expect(k8sClient.Create(ctx, op)).Should(Succeed())
	})
})
