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

package controller

import (
	"context"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = Describe("OBCluster controller", func() {

	const (
		createTimeout      = 30
		waitRunningTimeout = 300
		interval           = 1
	)

	Context("Create OBCluster", func() {
		It("Should successfully create OBCluster instance and ends with Status running", func() {
			By("By creating a new OBCluster")
			ctx := context.Background()
			obClusterName := "test-1-1"
			obcluster := newMinimalOBCluster(obClusterName, 1, 1)
			logf.Log.Info("create test obcluster")
			Expect(k8sClient.Create(ctx, obcluster)).Should(Succeed())

			obclusterLookupKey := types.NamespacedName{Name: obClusterName, Namespace: DefaultNamespace}
			createdOBCluster := &v1alpha1.OBCluster{}

			logf.Log.Info("check obcluster")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return true
			}, createTimeout, interval).Should(BeTrue())
			Expect(createdOBCluster.Spec.ClusterName).Should(Equal(obClusterName))
			logf.Log.Info("obcluster successfully created")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, obclusterLookupKey, createdOBCluster)
				if err != nil {
					return false
				}
				return createdOBCluster.Status.Status == clusterstatus.Running
			}, waitRunningTimeout, interval).Should(BeTrue())
			logf.Log.Info("obcluster status running")
		})
	})
})
