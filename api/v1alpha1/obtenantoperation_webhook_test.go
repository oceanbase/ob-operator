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
	apiconsts "github.com/oceanbase/ob-operator/api/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test OBTenantOperation Webhook", Label("webhook"), Serial, func() {
	clusterName := "test-cluster-for-operation"
	tenantPrimary := "test-tenant-for-operation"
	tenantStandby := "test-tenant-for-operation2"

	It("Create cluster and tenants", func() {
		c := newOBCluster(clusterName, 1, 1)
		t := newOBTenant(tenantPrimary, clusterName)
		t2 := newOBTenant(tenantStandby, clusterName)
		t2.Spec.TenantRole = apiconsts.TenantRoleStandby
		t2.Spec.Source = &TenantSourceSpec{
			Tenant: &tenantPrimary,
		}
		Expect(k8sClient.Create(ctx, c)).Should(Succeed())
		Expect(k8sClient.Create(ctx, t)).Should(Succeed())
		Expect(k8sClient.Create(ctx, t2)).Should(Succeed())
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
})
