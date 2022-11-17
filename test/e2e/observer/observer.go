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

	myconfig.ClusterName = "cn"

	var _ = ginkgo.Describe("create obcluster pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-1-1.yaml")
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

	var _ = ginkgo.Describe("add obzone pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-1.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", obcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", obcluster.GetName())

		ginkgo.It("step: wait before add obzone", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: add obzone from 1 to 3", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"add obzone update obcluster succeed",
			)
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
			gomega.Eventually(func() bool {
				obServerList = sqlOperator.GetOBServer()
				if len(obServerList) != 3 {
					return false
				}
				for _, observer := range obServerList {
					if observer.Status != observerconst.OBServerActive || observer.StartServiceTime <= 0 {
						return false
					}
				}
				return true
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"all observer is active",
			)
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, obcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"check observer and obcluster",
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
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-2.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", obcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", obcluster.GetName())

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
			gomega.Eventually(func() bool {
				obServerList = sqlOperator.GetOBServer()
				if len(obServerList) != 6 {
					return false
				}
				for _, observer := range obServerList {
					if observer.Status != observerconst.OBServerActive || observer.StartServiceTime <= 0 {
						return false
					}
				}
				return true
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"all observer is active",
			)
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, obcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"check observer and obcluster",
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
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-1.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", obcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", obcluster.GetName())

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
			gomega.Eventually(func() bool {
				obServerList = sqlOperator.GetOBServer()
				if len(obServerList) != 3 {
					return false
				}
				for _, observer := range obServerList {
					if observer.Status != observerconst.OBServerActive || observer.StartServiceTime <= 0 {
						return false
					}
				}
				return true
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"all observer is active",
			)
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, obcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"check observer and obcluster",
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

	})

	var _ = ginkgo.Describe("delete obzone pipeline", func() {

		var obServerList []model.AllServer
		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-1-1.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())
		rootserviceName := fmt.Sprintf("rs-%s", obcluster.GetName())
		obzoneName := fmt.Sprintf("obzone-%s", obcluster.GetName())

		ginkgo.It("step: wait before apply delete obzone", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: delete obzone from 3 to 1", func() {
			err := testresource.TestClient.UpdateOBClusterInstance(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update obcluster succeed",
			)
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
			gomega.Eventually(func() bool {
				obServerList = sqlOperator.GetOBServer()
				if len(obServerList) != 1 {
					return false
				}
				for _, observer := range obServerList {
					if observer.Status != observerconst.OBServerActive || observer.StartServiceTime <= 0 {
						return false
					}
				}
				return true
			}, OBClusterUpdateTReadyimeout, TryInterval).Should(
				gomega.Equal(true),
				"all observer is active",
			)
			status := testconverter.JudgeAllOBServerStatusByObj(obServerList, obcluster)
			gomega.Expect(status).To(
				gomega.Equal(true),
				"check observer and obcluster",
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

	})
})
