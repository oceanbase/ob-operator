/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	oceanbasev1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
)

var _ = Describe("OBTenantVariable Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		obtenantvariable := &oceanbasev1alpha1.OBTenantVariable{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind OBTenantVariable")
			err := k8sClient.Get(ctx, typeNamespacedName, obtenantvariable)
			if err != nil && errors.IsNotFound(err) {
				resource := &oceanbasev1alpha1.OBTenantVariable{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: oceanbasev1alpha1.OBTenantVariableSpec{OBTenant: "test-tenant", Variable: &apitypes.Variable{Name: "ob_query_timeout", Value: "10000000"}},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &oceanbasev1alpha1.OBTenantVariable{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance OBTenantVariable")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("reports the missing tenant dependency without claiming reconciliation succeeded", func() {
			By("Reconciling the created resource")
			controllerReconciler := &OBTenantVariableReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(10),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())
			current := &oceanbasev1alpha1.OBTenantVariable{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, current)).To(Succeed())
			Expect(current.Status.Status).NotTo(Equal("matched"))
		})
	})
})
