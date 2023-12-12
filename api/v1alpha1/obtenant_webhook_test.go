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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	apiconsts "github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
)

var _ = Describe("Test OBTenant Webhook", Label("webhook"), Serial, func() {

	clusterName := "test-cluster-name"
	tenantName := "test-tenant-name"

	It("Create a base cluster", func() {
		cluster := newOBCluster(clusterName, 1, 1)
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: defaultNamespace,
				Name:      clusterName,
			}, &OBCluster{})
			return err == nil
		}, 300, 1).Should(BeTrue())
	})

	It("Check positive unit number", func() {
		t := newOBTenant(tenantName, clusterName)
		t.Spec.UnitNumber = -2
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.UnitNumber = 0
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check legality of tenant name", func() {
		t := newOBTenant(tenantName, clusterName)
		t.Spec.TenantName = "illegal-tenant-name"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.TenantName = "0tenant1"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.TenantName = rand.String(129)
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.TenantName = ""
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check tenant roles", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.TenantRole = "hello"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())

		t.Spec.TenantRole = "standby1"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check existence of cluster", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.ClusterName = "cluster-not-exist"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check zones match or not", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.Pools[0].Zone = "zone_not_exist"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check existence of user secrets", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.Credentials.Root = "secret-not-exist1"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Credentials.StandbyRO = "secret-not-exist2"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Credentials.Root = wrongKeySecret
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check standby without a source", func() {
		t := newOBTenant(tenantName, clusterName)
		t.Spec.TenantRole = "Standby"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Source = &TenantSourceSpec{}
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		primaryTenantName := "tenant-not-exist"
		t.Spec.Source.Tenant = &primaryTenantName
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check standby with restore until without a limit key", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.Source = &TenantSourceSpec{}
		t.Spec.Source.Restore = &RestoreSourceSpec{
			Until: RestoreUntilConfig{
				Unlimited: false,
			},
		}
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})

	It("Check restoring w/o OSS access secret", func() {
		t := newOBTenant(tenantName, clusterName)

		t.Spec.Source = &TenantSourceSpec{}
		t.Spec.Source.Restore = &RestoreSourceSpec{
			Until: RestoreUntilConfig{
				Unlimited: true,
			},
		}
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())

		t.Spec.Source.Restore.ArchiveSource = &apitypes.BackupDestination{}
		t.Spec.Source.Restore.ArchiveSource.Type = apiconsts.BackupDestTypeOSS
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Source.Restore.ArchiveSource.OSSAccessSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Source.Restore.ArchiveSource.OSSAccessSecret = defaultSecretName
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())

		t.Spec.Source.Restore.BakDataSource = &apitypes.BackupDestination{}
		t.Spec.Source.Restore.BakDataSource.Type = apiconsts.BackupDestTypeOSS
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Source.Restore.BakDataSource.OSSAccessSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
		t.Spec.Source.Restore.BakDataSource.OSSAccessSecret = defaultSecretName
		Expect(k8sClient.Create(ctx, t)).ShouldNot(Succeed())
	})
})
