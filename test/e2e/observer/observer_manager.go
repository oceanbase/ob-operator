/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	myconfig "github.com/oceanbase/ob-operator/pkg/config"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
)

var _ = ginkgo.Describe("observer test pipeline", func() {

	myconfig.ClusterName = "cn"

	var _ = ginkgo.Describe("create obcluster pipeline", func() {

		obcluster := testconverter.GetObjFromYaml("./data/obcluster-1-1.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()

		ginkgo.It("step: create obcluster", func() {
			err := testresource.TestClient.CreateObj(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create obcluster succeed",
			)
		})

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() string {
				return testresource.TestClient.GetOBStatusClusterStatus(obclusterNamespace, obclusterName, myconfig.ClusterName)
			}, OBClusterReadyimeout, TryInterval).Should(
				gomega.Equal(observerconst.ClusterReady),
				"obcluster %s ready", obclusterName,
			)
		})

	})

	var _ = ginkgo.Describe("add observer pipeline", func() {

		obcluster := testconverter.GetObjFromYaml("./data/obcluster-1-2.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()

		ginkgo.It("step: wait before apply add observer", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: add observer to 2 per zone", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
		})

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() string {
				return testresource.TestClient.GetOBStatusClusterStatus(obclusterNamespace, obclusterName, myconfig.ClusterName)
			}, OBClusterReadyimeout, TryInterval).Should(
				gomega.Equal(observerconst.ClusterReady),
				"obcluster %s ready", obclusterName,
			)
		})

	})

	var _ = ginkgo.Describe("delete observer pipeline", func() {

		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-1.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()

		ginkgo.It("step: wait before apply add observer", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: delete observer to 1 per zone", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
		})

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() string {
				return testresource.TestClient.GetOBStatusClusterStatus(obclusterNamespace, obclusterName, myconfig.ClusterName)
			}, OBClusterReadyimeout, TryInterval).Should(
				gomega.Equal(observerconst.ClusterReady),
				"obcluster %s ready", obclusterName,
			)
		})

	})
})
