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
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
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

	It("Forbid to modify resources of Non-standalone cluster", func() {
		cluster := newOBCluster("test", 1, 1)
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		By("Modify resources of Non-standalone cluster")
		cluster.Spec.OBServerTemplate.Resource.Cpu = resource.MustParse("3")
		Expect(k8sClient.Update(ctx, cluster)).ShouldNot(Succeed())
		cluster = newOBCluster("test", 1, 1)
		cluster.Spec.OBServerTemplate.Resource.Memory = resource.MustParse("20Gi")
		Expect(k8sClient.Update(ctx, cluster)).ShouldNot(Succeed())

		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
	})

	It("It's OK to modify resources of standalone cluster", func() {
		cluster := newOBCluster("test", 1, 1)
		cluster.Annotations[oceanbaseconst.AnnotationsMode] = oceanbaseconst.ModeStandalone
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		By("Modify resources of standalone cluster")
		cluster.Spec.OBServerTemplate.Resource.Cpu = resource.MustParse("3")
		Expect(k8sClient.Update(ctx, cluster)).Should(Succeed())

		cluster.Spec.OBServerTemplate.Resource.Memory = resource.MustParse("20Gi")
		Expect(k8sClient.Update(ctx, cluster)).Should(Succeed())

		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
	})

	It("Validate existence of secrets", func() {
		By("Create normal cluster")
		cluster := newOBCluster("test3", 1, 1)
		cluster.Spec.UserSecrets.Monitor = ""
		cluster.Spec.UserSecrets.ProxyRO = ""
		cluster.Spec.UserSecrets.Operator = ""
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())

		cluster2 := newOBCluster("test2", 1, 1)
		cluster2.Spec.UserSecrets.Monitor = "secret-that-does-not-exist"
		cluster2.Spec.UserSecrets.ProxyRO = ""
		cluster2.Spec.UserSecrets.Operator = ""
		Expect(k8sClient.Create(ctx, cluster2)).Should(Succeed())

		cluster3 := newOBCluster("test3", 1, 1)
		cluster2.Spec.UserSecrets.Monitor = wrongKeySecret
		Expect(k8sClient.Create(ctx, cluster3)).ShouldNot(Succeed())

		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, cluster2)).Should(Succeed())
	})

	It("Validate secrets creation and fetch them", func() {
		By("Create normal cluster")
		cluster := newOBCluster("test-create-secrets", 1, 1)
		cluster.Spec.UserSecrets.Monitor = ""
		cluster.Spec.UserSecrets.ProxyRO = ""
		cluster.Spec.UserSecrets.Operator = ""
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())

		By("Check user secrets filling up situation")
		Expect(cluster.Spec.UserSecrets.Monitor).ShouldNot(BeEmpty())
		Expect(cluster.Spec.UserSecrets.ProxyRO).ShouldNot(BeEmpty())
		Expect(cluster.Spec.UserSecrets.Operator).ShouldNot(BeEmpty())
		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
	})

	It("Validate single pvc with multiple storage classes", func() {
		cluster := newOBCluster("test", 1, 1)
		cluster.Annotations[oceanbaseconst.AnnotationsSinglePVC] = "true"
		cluster.Spec.OBServerTemplate.Storage.DataStorage.StorageClass = "local-path-2"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())
	})

	It("Validate service account binding", func() {
		cluster := newOBCluster("test-cluster-sa", 1, 1)
		cluster.Spec.ServiceAccount = "non-exist"
		Expect(k8sClient.Create(ctx, cluster)).ShouldNot(Succeed())

		saName := "test-sa"
		err := k8sClient.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      saName,
				Namespace: defaultNamespace,
			},
		})
		Expect(err).Should(BeNil())
		cluster.Spec.ServiceAccount = saName
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())

		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
	})

	It("Validate memory limit", func() {
		cluster := newOBCluster("test-memory", 1, 1)
		cluster.Spec.OBServerTemplate.Resource.Memory = resource.MustParse("16Gi")
		Expect(k8sClient.Create(ctx, cluster)).Should(Succeed())
		for _, param := range cluster.Spec.Parameters {
			if param.Name == "memory_limit" {
				innerMemory := resource.MustParse(param.Value)
				expectedMemory := resource.MustParse("14745M") // floor(16 * 1024 * 0.9)
				Expect(innerMemory.Value()).Should(Equal(expectedMemory.Value()))
			}
		}
		Expect(k8sClient.Delete(ctx, cluster)).Should(Succeed())
	})
})
