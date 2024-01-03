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
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	apitypes "github.com/oceanbase/ob-operator/api/types"
)

var _ = Describe("Test OBCluster Webhook", Label("webhook"), func() {
	It("Check existence of secrets", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.UserSecrets.Root = "secret-that-does-not-exist"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check topology", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.Topology = nil
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check storage class", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass = "storage-class-that-does-not-exist"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check memory size", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.OBServerTemplate.Resource.Memory = resource.MustParse("5Gi")
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check Data disk size", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.OBServerTemplate.Storage.DataStorage.Size = resource.MustParse("5Gi")
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check Log disk size", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.OBServerTemplate.Storage.LogStorage.Size = resource.MustParse("5Gi")
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check redo log disk size", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.OBServerTemplate.Storage.RedoLogStorage.Size = resource.MustParse("5Gi")
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check overflowed memory limit", func() {
		cluster := newOBCluster("test", 3, 1)

		cluster.Spec.Parameters = []apitypes.Parameter{{
			Name:  "memory_limit",
			Value: "100Gi",
		}}
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
		cluster.Spec.Parameters[0].Value = "123dsf123gdf"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
		cluster.Spec.Parameters[0].Value = "107374182400"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Check overflowed data file max size", func() {
		cluster := newOBCluster("test", 3, 1)
		cluster.Spec.Parameters = []apitypes.Parameter{{
			Name:  "datafile_maxsize",
			Value: "100Gi",
		}}
		cluster.Spec.Parameters[0].Value = "123dsf123gdf"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
		cluster.Spec.Parameters[0].Value = "107374182400"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Create OBCluster successfully", func() {
		cluster := newOBCluster("test", 1, 1)
		By("Create OBCluster normally")
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		Eventually(func() bool {
			c := &OBCluster{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: defaultNamespace,
				Name:      "test",
			}, c)).Should(Succeed())
			return c != nil
		}, 300, 1).Should(BeTrue())

		By("Delete OBCluster normally")
		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
		Eventually(func() bool {
			c := &OBCluster{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Namespace: defaultNamespace,
				Name:      "test",
			}, c)
			return err != nil && kubeerrors.IsNotFound(err)
		}, 300, 1).Should(BeTrue())
	})
})
