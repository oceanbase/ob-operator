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
	"log"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	myconfig "github.com/oceanbase/ob-operator/pkg/config"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
	testutils "github.com/oceanbase/ob-operator/test/e2e/utils"
)

var _ = ginkgo.Describe("obcluster upgrade test pipeline", func() {

	myconfig.ClusterName = "cn"

	var _ = ginkgo.Describe("create obcluster pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-v311.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", obcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", obcluster.GetName())

		ginkgo.It("step: create obcluster", func() {
			err := testresource.TestClient.CreateObj(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create obcluster succeed",
			)
		})

		ginkgo.It("step: check obcluster bootstrap status", func() {
			gomega.Eventually(func() string {
				return testresource.TestClient.GetOBClusterStatus(obclusterNamespace, obclusterName)
			}, OBClusterBootstrapTimeout, TryInterval).Should(
				gomega.Equal(observerconst.TopologyReady),
				"obcluster %s bootstrap ready", obclusterName,
			)
		})

		ginkgo.It("step: check statefulapp status for obcluster", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(obclusterNamespace, statefulappName)
			}, OBClusterReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
			)
		})

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeOBClusterInstanceIsReadyByObj(obclusterNamespace, obclusterName)
			}, OBClusterReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"obcluster %s is ready", obclusterName,
			)
		})

		ginkgo.It("step: check service status for obcluster", func() {
			var err error
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeServicefForOBClusterIsReadyByObj(obclusterNamespace, serviceName)
			}, OBClusterReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"service %s is ready", serviceName,
			)
			service, err = testresource.TestClient.GetService(obclusterNamespace, serviceName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get service %s succeed", serviceName,
			)
		})

		ginkgo.It("step: check observer status", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			gomega.Eventually(func() bool {
				obServerList = sqlOperator.GetOBServer()
				if len(obServerList) == 0 {
					return false
				}
				for _, observer := range obServerList {
					if observer.Status != observerconst.OBServerActive || observer.StartServiceTime <= 0 {
						return false
					}
				}
				return true
			}, OBClusterReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"all observer is active",
			)
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, obcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"check observer and obcluster matched",
			)
		})

		ginkgo.It("step: check rootservice status", func() {
			rootservice, err := testresource.TestClient.GetRootService(obclusterNamespace, rootserviceName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get rootservice %s succeed", rootserviceName,
			)
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			rsList := sqlOperator.GetRootService()
			status := testconverter.JudgeRootserviceStatusByObj(rsList, rootservice)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"rootservice %s status is ok", rootserviceName,
			)
		})

		ginkgo.It("step: check obzone status", func() {

			gomega.Eventually(func() bool {
				_, err := testresource.TestClient.GetOBZone(obclusterNamespace, obzoneName)
				return err == nil
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"get obzone %s succeed", obzoneName,
			)

			obzone, _ := testresource.TestClient.GetOBZone(obclusterNamespace, obzoneName)

			status := testconverter.JudgeOBzoneStatusByObj(obServerList, obzone)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"obzone %s status is ok", obzoneName,
			)
		})

	})

	var _ = ginkgo.Describe("upgrade obcluster pipeline", func() {

		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-v314.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()

		ginkgo.It("step: check major freeze start", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			gomega.Eventually(func() bool {
				frozenVersion := sqlOperator.GetFrozenVersion()
				log.Println("frozenVersion: ", frozenVersion)
				formerFrozenVersion := frozenVersion[0].Value

				err := sqlOperator.MajorFreeze()
				gomega.Expect(err).To(
					gomega.BeNil(),
					"Major freeze start",
				)
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
				currentVersion := sqlOperator.GetVersion()[0].Version
				log.Println("currentVersion: ", currentVersion)
				return currentVersion == targetVersion
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
