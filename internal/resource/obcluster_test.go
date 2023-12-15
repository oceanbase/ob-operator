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

package resource

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
)

const (
	TestOBClusterName = "test"
)

var _ = Describe("OBCluster controller", func() {

	const (
		applyWait           = 5
		commonTimeout       = 30
		waitRunningTimeout  = 600
		waitUpgradeTimeout  = 1800
		waitDeletingTimeout = 1800
		interval            = 1
	)

	var _ = BeforeEach(func() {
		By("Create cluster secrets")
		ctx := context.Background()
		secrets := newClusterSecrets()
		for _, v := range secrets {
			Expect(k8sClient.Create(ctx, v)).Should(Succeed())
		}
		for _, v := range secrets {
			sec := &v1.Secret{}
			Eventually(func() bool {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      v.GetName(),
				}, sec) == nil
			}, commonTimeout, interval).Should(BeTrue())
		}
	})

	var _ = AfterEach(func() {
		secrets := newClusterSecrets()
		ctx := context.Background()
		for _, v := range secrets {
			Expect(k8sClient.Delete(ctx, v)).To(BeNil())
		}
		for _, v := range secrets {
			sec := &v1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      v.GetName(),
				}, sec)
				return err != nil && kubeerrors.IsNotFound(err)
			}, commonTimeout, interval).Should(BeTrue())
		}
	})

	Context("Create OBCluster", Label("long-run"), Serial, func() {
		It("Create OBCluster instance and ends with status running successfully", func() {
			By("Create test obcluster")
			ctx := context.Background()
			obcluster := newOBCluster(TestOBClusterName, 3, 1)
			Expect(k8sClient.Create(ctx, obcluster)).Should(Succeed())

			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			createdOBCluster := &v1alpha1.OBCluster{}

			By("Check obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return true
			}, commonTimeout, interval).Should(BeTrue())
			Expect(createdOBCluster.Spec.ClusterName).Should(Equal(TestOBClusterName))

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return createdOBCluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
		})

		It("Delete pod of OBCluster and recover successfully", func() {
			By("List pods and delete one of them")
			ctx := context.Background()
			podList := &v1.PodList{}
			k8sClient.List(ctx, podList, client.MatchingLabels{
				oceanbaseconst.LabelRefOBCluster: TestOBClusterName,
			})
			deleteTarget := rand.Intn(3)
			Expect(len(podList.Items)).Should(Equal(3))
			Expect(k8sClient.Delete(ctx, &podList.Items[deleteTarget])).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      podList.Items[deleteTarget].GetName(),
				}, &v1.Pod{})
				return err != nil && kubeerrors.IsNotFound(err)
			}, 100, 1).Should(BeTrue())

			By("Check # of pods to be 3")
			Eventually(func() bool {
				k8sClient.List(ctx, podList, client.MatchingLabels{
					oceanbaseconst.LabelRefOBCluster: TestOBClusterName,
				})
				return len(podList.Items) == 3
			}, 300, 1).Should(BeTrue())

			By("Wait for pod to recover")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      podList.Items[deleteTarget].GetName(),
				}, &v1.Pod{})
				return err == nil
			}, 300, 1).Should(BeTrue())

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				obcluster := &v1alpha1.OBCluster{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      TestOBClusterName,
				}, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, 300, 1).Should(BeTrue())

			By("Wait for # observers to be 3")
			Eventually(func() bool {
				observerList := &v1alpha1.OBServerList{}
				k8sClient.List(ctx, observerList, client.MatchingLabels{
					oceanbaseconst.LabelRefOBCluster: TestOBClusterName,
				})
				return len(observerList.Items) == 3
			}).Should(BeTrue())
		})

		It("Scale out zones of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale out obzone of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			By("Get obcluster and scale out zones")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			newZone := apitypes.OBZoneTopology{
				Zone:    fmt.Sprintf("zone%d", len(obcluster.Spec.Topology)),
				Replica: 1,
			}
			obcluster.Spec.Topology = append(obcluster.Spec.Topology, newZone)
			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
		})

		It("Scale in obzone of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale in obzone of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			By("Get obcluster and scale in zones")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			obcluster.Spec.Topology = obcluster.Spec.Topology[0:3]
			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
		})

		It("Upgrade OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By upgrade OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			By("Get obcluster and upgrade it")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			obcluster.Spec.OBServerTemplate.Image = UpgradeImage
			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitUpgradeTimeout, interval).Should(BeTrue())
		})

		It("Scale out server of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale out observer of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			By("Check obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			for idx := 0; idx < len(obcluster.Spec.Topology); idx++ {
				obcluster.Spec.Topology[idx].Replica = 2
			}

			By("Scale out observer")
			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
		})

		It("Scale in server of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale in observer of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			By("Get obcluster and update it")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			for idx := 0; idx < len(obcluster.Spec.Topology); idx++ {
				obcluster.Spec.Topology[idx].Replica = 1
			}

			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitDeletingTimeout, interval).Should(BeTrue())
		})

		It("Delete OBCluster instance successfully", func() {
			// Skip("Skip all tests")
			By("Get obcluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())

			By("Delete obcluster")
			Eventually(func() bool {
				err := k8sClient.Delete(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)

			By("Wait for obcluster to get deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil && kubeerrors.IsNotFound(err) {
					return true
				}
				return false
			}, waitRunningTimeout, interval).Should(BeTrue())
		})
	})
})
