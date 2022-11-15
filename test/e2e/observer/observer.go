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
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
	testutils "github.com/oceanbase/ob-operator/test/e2e/utils"
)

var _ = ginkgo.Describe("observer test pipeline", func() {

	var _ = ginkgo.Describe("create default obcluster pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		defaultOBcluster := testconverter.GetObjFromYaml("./data/obcluster-1-1.yaml")
		obclusterNamespace := defaultOBcluster.GetNamespace()
		obclusterName := defaultOBcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", defaultOBcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", defaultOBcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", defaultOBcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", defaultOBcluster.GetName())
		myconfig.ClusterName = "test"

		ginkgo.It("step: create default obcluster", func() {
			err := testresource.TestClient.CreateObj(defaultOBcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create default obcluster succeed",
			)
		})

		ginkgo.It("step: wait for obcluster created", func() {
			time.Sleep(OBClusterCreateTimeout)
		})

		ginkgo.It("step: check default obcluster bootstrap status", func() {
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
		ginkgo.It("step: check default obcluster status", func() {
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
			obServerList = sqlOperator.GetOBServer()
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, defaultOBcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"all observer is active",
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

	var _ = ginkgo.Describe("add observer pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		addOBcluster := testconverter.GetObjFromYaml("./data/obcluster-02.yaml")
		obclusterNamespace := addOBcluster.GetNamespace()
		obclusterName := addOBcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", addOBcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", addOBcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", addOBcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", addOBcluster.GetName())

		ginkgo.It("step: update obcluster for add observer by zone", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(addOBcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
		})

		ginkgo.It("step: wait for obcluster updated", func() {
			time.Sleep(OBClusterCreateTimeout)
		})

		ginkgo.It("step: check statefulapp status for obcluster", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(obclusterNamespace, statefulappName)
			}, StatefulappUpdateReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
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

		ginkgo.It("step: get service for obcluster", func() {
			var err error
			service, err = testresource.TestClient.GetService(obclusterNamespace, serviceName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get service %s succeed", serviceName,
			)
		})

		ginkgo.It("step: check observer status", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			obServerList = sqlOperator.GetOBServer()
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, addOBcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"all observer is active",
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

	var _ = ginkgo.Describe("delete observer pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		delOBcluster := testconverter.GetObjFromYaml("./data/obcluster-01.yaml")
		obclusterNamespace := delOBcluster.GetNamespace()
		obclusterName := delOBcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", delOBcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", delOBcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", delOBcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", delOBcluster.GetName())

		ginkgo.It("step: update obcluster for del observer by zone", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(delOBcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
		})

		ginkgo.It("step: wait for obcluster updated", func() {
			time.Sleep(OBClusterCreateTimeout)
		})

		ginkgo.It("step: check statefulapp status for obcluster", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(obclusterNamespace, statefulappName)
			}, StatefulappUpdateReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
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

		ginkgo.It("step: get service for obcluster", func() {
			var err error
			service, err = testresource.TestClient.GetService(obclusterNamespace, serviceName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get service %s succeed", serviceName,
			)
		})

		ginkgo.It("step: check observer status", func() {
			sqlOperator := testutils.NewDefaultSqlOperator(service.Spec.ClusterIP)
			obServerList = sqlOperator.GetOBServer()
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, delOBcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"all observer is active",
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
			obzone, err := testresource.TestClient.GetOBZone(obclusterNamespace, obzoneName)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"get obzone %s succeed", obzoneName,
			)
			status := testconverter.JudgeOBzoneStatusByObj(obServerList, obzone)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"obzone %s status is ok", obzoneName,
			)
		})

		ginkgo.It("step: delete default obcluster", func() {
			testresource.TestClient.DeleteObj(delOBcluster)
		})
	})
})
