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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

var _ = Describe("OBResourceRescueWebhook", func() {
	It("Validate create", func() {
		rescue := newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())
	})

	It("Validate wrong types", func() {
		rescue := newOBResourceRescue()
		rescue.Spec.Type = "wrong"
		Expect(k8sClient.Create(ctx, rescue)).ShouldNot(Succeed())
	})

	It("Validate update", func() {
		rescue := newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.Type = "reset"
		rescue.Spec.TargetStatus = "running"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())
	})

	It("Validate target status field when type is reset", func() {
		rescue := newOBResourceRescue()
		rescue.Spec.Type = "reset"
		Expect(k8sClient.Create(ctx, rescue)).ShouldNot(Succeed())
		rescue.Spec.TargetStatus = "Running"
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())
	})

	It("Validate empty kind, resName, and type", func() {
		rescue := newOBResourceRescue()
		rescue.Spec.TargetKind = ""
		Expect(k8sClient.Create(ctx, rescue)).ShouldNot(Succeed())
		rescue.Spec.TargetKind = "OBCluster"
		rescue.Spec.TargetResName = ""
		Expect(k8sClient.Create(ctx, rescue)).ShouldNot(Succeed())
		rescue.Spec.TargetResName = "test"
		rescue.Spec.Type = ""
		Expect(k8sClient.Create(ctx, rescue)).ShouldNot(Succeed())
		rescue.Spec.Type = "delete"
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())
	})

	It("Validate forbidding to update a resource", func() {
		rescue := newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.Type = "reset"
		rescue.Spec.TargetStatus = "working"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())

		rescue = newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.TargetKind = "OBTenant"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())

		rescue = newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.TargetResName = "test2"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())

		rescue = newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.Namespace = "test232"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())

		rescue = newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.TargetGV = "oceanbase.oceanbase.com/v2"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())

		rescue = newOBResourceRescue()
		Expect(k8sClient.Create(ctx, rescue)).Should(Succeed())
		rescue.Spec.TargetStatus = "failed"
		Expect(k8sClient.Update(ctx, rescue)).ShouldNot(Succeed())
		Expect(k8sClient.Delete(ctx, rescue)).Should(Succeed())
	})
})

func newOBResourceRescue() *OBResourceRescue {
	return &OBResourceRescue{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.String(10),
			Namespace: defaultNamespace,
		},
		Spec: OBResourceRescueSpec{
			TargetKind:    "OBCluster",
			TargetResName: "test",
			Type:          "delete",
		},
		Status: OBResourceRescueStatus{
			Status: "Successful",
		},
	}
}
