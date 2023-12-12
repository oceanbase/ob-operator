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

var _ = Describe("Test OBTenantBackupPolicy Webhook", Label("webhook"), Serial, func() {

	clusterName := "test-cluster-for-obtbp"
	tenantName := "test-tenant-for-obtbp"
	policyName := "test-policy"

	It("Create cluster and tenant", func() {
		c := newOBCluster(clusterName, 1, 1)
		t := newOBTenant(tenantName, clusterName)
		Expect(k8sClient.Create(ctx, c)).Should(Succeed())
		Expect(k8sClient.Create(ctx, t)).Should(Succeed())
	})

	It("Check existence of cluster", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.ObClusterName = ""
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.ObClusterName = "cluster-that-does-not-exist"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("TenantName and TenantCRName are both empty", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.TenantName = ""
		p.Spec.TenantCRName = ""
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("TenantCRName and TenantSecret are both empty", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.TenantCRName = ""
		p.Spec.TenantSecret = ""
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check existence of EncryptionSecret", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)

		p.Spec.DataBackup.EncryptionSecret = ""
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.EncryptionSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.EncryptionSecret = wrongKeySecret
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check Archive binding type", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Binding = ""
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Binding = "Binding1"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check destination types", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Destination.Type = "123"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p = newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.DataBackup.Destination.Type = "123"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check oss secret 1", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Destination.Type = apiconsts.BackupDestTypeOSS

		p.Spec.LogArchive.Destination.OSSAccessSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.OSSAccessSecret = wrongKeySecret
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check oss secret 2", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Destination.Type = apiconsts.BackupDestTypeOSS

		p.Spec.DataBackup.Destination.OSSAccessSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.OSSAccessSecret = wrongKeySecret
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check oss secret 2", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Destination.Type = apiconsts.BackupDestTypeOSS

		p.Spec.DataBackup.Destination.OSSAccessSecret = "secret-not-exist"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.OSSAccessSecret = wrongKeySecret
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check oss path", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.Destination.Path = "oss://bucket/backup?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss://bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss://operator-backup-data/backup-t1?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss://bucket?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss://bucket/backup?host="
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "soss://bucket/backup?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss:///bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "oss://?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.Destination.Path = "bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check oss path 2", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.DataBackup.Destination.Path = "oss://bucket/backup?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss://bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss://operator-backup-data/backup-t1?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss://bucket?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss://bucket/backup?host="
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "soss://bucket/backup?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss:///bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "oss://?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.Destination.Path = "bucket/backup/?host=oss-cn-hangzhou.aliyuncs.com"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check Crontab", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.DataBackup.FullCrontab = "* * *"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.FullCrontab = "* 1 2 4 5 6 7 8 9 10"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})
	It("Check Crontab 2", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.DataBackup.IncrementalCrontab = "* * *"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataBackup.IncrementalCrontab = "* 1 2 4 5 6 7 8 9 10"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check intervals", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.JobKeepWindow = "1sec"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.JobKeepWindow = "1h"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check intervals 2", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.DataClean.RecoveryWindow = "1sec"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.DataClean.RecoveryWindow = "1h"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})

	It("Check intervals 3", func() {
		p := newBackupPolicy(policyName, tenantName, clusterName)
		p.Spec.LogArchive.SwitchPieceInterval = "1sec"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
		p.Spec.LogArchive.SwitchPieceInterval = "1h"
		Expect(k8sClient.Create(ctx, p)).ShouldNot(Succeed())
	})
})
