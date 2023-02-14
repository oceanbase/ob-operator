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
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	myconfig "github.com/oceanbase/ob-operator/pkg/config"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
	testutils "github.com/oceanbase/ob-operator/test/e2e/utils"
)

var _ = ginkgo.Describe("obcluster upgrade test pipeline", func() {

	myconfig.ClusterName = "cn"

	var _ = ginkgo.Describe("create obcluster pipeline", func() {

		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-1-v311.yaml")
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

	var _ = ginkgo.Describe("upgrade obcluster pipeline", func() {

		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-1-v314.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())

		ginkgo.It("step: get service for obcluster", func() {
			var err error
			service, err = testresource.TestClient.GetService(obclusterNamespace, serviceName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get service %s succeed", serviceName,
			)
		})

		ginkgo.It("step: check major freeze start", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			frozenVersion := sqlOperator.GetFrozenVersion()
			formerFrozenVersion := frozenVersion[0].Value
			err := sqlOperator.MajorFreeze()
			gomega.Expect(err).To(
				gomega.BeNil(),
				"Major freeze start",
			)
			gomega.Eventually(func() bool {
				frozenVersion = sqlOperator.GetFrozenVersion()
				latterFrozenVersion := frozenVersion[0].Value
				return formerFrozenVersion != latterFrozenVersion
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"Major freeze start OK",
			)
		})

		ginkgo.It("step: wait for major freeze", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			gomega.Eventually(func() bool {
				zoneLastMergedVersion := sqlOperator.GetLastMergedVersion()
				return len(zoneLastMergedVersion) == 0
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"Major freeze Succeed",
			)
		})

		ginkgo.It("step: wait before apply upgrade obcluster", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: upgrade obcluster to 3.1.4", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
		})

		ginkgo.It("step: check obcluster upgrade succeed", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			targetVersion := "3.1.4"
			gomega.Eventually(func() bool {
				version := sqlOperator.GetVersion()
				return len(version) != 0 && version[0].Version == targetVersion
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"obcluster upgrade succeed",
			)
		})

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeOBClusterInstanceIsReadyByObj(obclusterNamespace, obclusterName)
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"obcluster %s is ready", obclusterName,
			)
		})
	})
})
