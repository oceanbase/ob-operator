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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"

	"github.com/oceanbase/ob-operator/internal/resource/utils"
)

const (
	TestOBClusterName = "test"
)

var _ = Describe("OBCluster controller", Serial, func() {

	const (
		applyWait           = 5
		commonTimeout       = 30
		waitRunningTimeout  = 600
		waitUpgradeTimeout  = 1800
		waitDeletingTimeout = 1800
		interval            = 1
	)

	Context("Utils", func() {
		It("Min", func() {
			Expect(utils.Min(1, 2)).Should(Equal(1))
			Expect(utils.Min(2, 1)).Should(Equal(1))
			Expect(utils.Min(1.2, 2.1)).Should(Equal(1.2))
			Expect(utils.Min(2.1, 1.2)).Should(Equal(1.2))
		})

		It("IsZero", func() {
			Expect(utils.IsZero(0)).Should(BeTrue())
			Expect(utils.IsZero(1)).Should(BeFalse())
			Expect(utils.IsZero(0.0)).Should(BeTrue())
			Expect(utils.IsZero(1.0)).Should(BeFalse())
			Expect(utils.IsZero("")).Should(BeTrue())
			Expect(utils.IsZero("a")).Should(BeFalse())
			Expect(utils.IsZero(&v1.Secret{})).Should(BeFalse())
		})
	})

	Context("Create OBCluster", Label("long-run"), func() {
		It("Create OBCluster instance and ends with status running successfully", func() {
			// Skip("Skip all tests")
			By("By creating a new OBCluster")
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
			obcluster := newOBCluster(TestOBClusterName, 3, 1)
			logf.Log.Info("create test obcluster")
			Expect(k8sClient.Create(ctx, obcluster)).Should(Succeed())

			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			createdOBCluster := &v1alpha1.OBCluster{}

			logf.Log.Info("check obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return true
			}, commonTimeout, interval).Should(BeTrue())
			Expect(createdOBCluster.Spec.ClusterName).Should(Equal(TestOBClusterName))
			logf.Log.Info("obcluster successfully created")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return createdOBCluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully created")
		})

		It("Scale out zones of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale out obzone of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("get obcluster")
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
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully scale out obzone")
		})

		It("Scale in obzone of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale in obzone of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("get obcluster")
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
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully scale in obzone")
		})

		It("Upgrade OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By upgrade OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("get obcluster")
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
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitUpgradeTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully upgrade obzone")
		})

		It("Scale out server of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale out observer of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("check obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			for idx := 0; idx < len(obcluster.Spec.Topology); idx++ {
				obcluster.Spec.Topology[idx].Replica = 2
			}

			Eventually(func() bool {
				err := k8sClient.Update(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully scale out observer")
		})

		It("Scale in server of OBCluster instance and ends with status running successfully", func() {
			Skip("Skip all tests")
			By("By scale in observer of an OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("get obcluster")
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
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil {
					return false
				}
				return obcluster.Status.Status == clusterstatus.Running
			}, waitDeletingTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster successfully scale in observer")
		})

		It("Delete OBCluster instance successfully", func() {
			// Skip("Skip all tests")
			By("By delete OBCluster")
			ctx := context.Background()
			obclusterLookupKey := types.NamespacedName{Name: TestOBClusterName, Namespace: DefaultNamespace}
			obcluster := &v1alpha1.OBCluster{}

			logf.Log.Info("get obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			Eventually(func() bool {
				err := k8sClient.Delete(ctx, obcluster)
				return err == nil
			}, commonTimeout, interval).Should(BeTrue())
			time.Sleep(applyWait * time.Second)
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, obcluster)
				if err != nil && kubeerrors.IsNotFound(err) {
					return true
				}
				return false
			}, waitRunningTimeout, interval).Should(BeTrue())

			secrets := newClusterSecrets()
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
			logf.Log.Info("obcluster successfully deleted")
		})
	})
})
