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

package statefulapp

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
)

var _ = ginkgo.Describe("statefulapp test pipeline", func() {

	var _ = ginkgo.Describe("create default statefulapp pipeline", func() {

		defaultStatefulapp := testconverter.GetObjFromYaml("./data/statefulapp-01.yaml")
		statefulappNamespace := defaultStatefulapp.GetNamespace()
		statefulappName := defaultStatefulapp.GetName()

		ginkgo.It("step: create default statefulapp", func() {
			err := testresource.TestClient.CreateObj(defaultStatefulapp)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create default statefulapp succeed",
			)
		})

		ginkgo.It("step: wait for statefulapp created", func() {
			time.Sleep(StatefulappCreateTimeout)
		})

		ginkgo.It("step: check default pod status", func() {
			// is zone1Pod0 created
			zone1Pod0 := fmt.Sprintf("%s-test-zone1-%d", statefulappName, 0)
			_, err := testresource.TestClient.IsPodExists(statefulappNamespace, zone1Pod0)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create pod %s succeed", zone1Pod0,
			)
			// is zone1Pod0 running
			gomega.Eventually(func() corev1.PodPhase {
				return testresource.TestClient.GetPodStatus(statefulappNamespace, zone1Pod0)
			}, PodCreateTimeout, TryInterval).Should(
				gomega.Equal(corev1.PodRunning),
				"pod %s is running", zone1Pod0,
			)
			// is zone2Pod0 created
			zone2Pod0 := fmt.Sprintf("%s-test-zone2-%d", statefulappName, 0)
			_, err = testresource.TestClient.IsPodExists(statefulappNamespace, zone2Pod0)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create pod %s succeed", zone2Pod0,
			)
			// is zone2Pod0 running
			gomega.Eventually(func() corev1.PodPhase {
				return testresource.TestClient.GetPodStatus(statefulappNamespace, zone2Pod0)
			}, PodCreateTimeout, TryInterval).Should(
				gomega.Equal(corev1.PodRunning),
				"pod %s is running", zone2Pod0,
			)
			// is zone3Pod0 created
			zone3Pod0 := fmt.Sprintf("%s-test-zone3-%d", statefulappName, 0)
			_, err = testresource.TestClient.IsPodExists(statefulappNamespace, zone3Pod0)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create pod %s succeed", zone3Pod0,
			)
			// is zone3Pod0 running
			gomega.Eventually(func() corev1.PodPhase {
				return testresource.TestClient.GetPodStatus(statefulappNamespace, zone3Pod0)
			}, PodCreateTimeout, TryInterval).Should(
				gomega.Equal(corev1.PodRunning),
				"pod %s is running", zone3Pod0,
			)
		})

		ginkgo.It("step: check default statefulapp status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(statefulappNamespace, statefulappName)
			}, StatefulappReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
			)
		})

	})

	var _ = ginkgo.Describe("add pod by zone pipeline", func() {

		addPodStatefulapp := testconverter.GetObjFromYaml("./data/statefulapp-02.yaml")
		statefulappNamespace := addPodStatefulapp.GetNamespace()
		statefulappName := addPodStatefulapp.GetName()

		ginkgo.It("step: update statefulapp for add pod by zone", func() {
			err := testresource.TestClient.UpdateStatefulappInstance(addPodStatefulapp)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update statefulapp succeed",
			)
		})

		ginkgo.It("step: wait for statefulapp updated", func() {
			time.Sleep(StatefulappCreateTimeout)
		})

		ginkgo.It("step: check pod status which is added", func() {
			// is zone1Pod1 created
			zone1Pod1 := fmt.Sprintf("%s-test-zone1-%d", statefulappName, 1)
			_, err := testresource.TestClient.IsPodExists(statefulappNamespace, zone1Pod1)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create pod %s succeed", zone1Pod1,
			)
			// is zone1Pod1 running
			gomega.Eventually(func() corev1.PodPhase {
				return testresource.TestClient.GetPodStatus(statefulappNamespace, zone1Pod1)
			}, PodCreateTimeout, TryInterval).Should(
				gomega.Equal(corev1.PodRunning),
				"pod %s is running", zone1Pod1,
			)
		})

		ginkgo.It("step: check statefulapp status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(statefulappNamespace, statefulappName)
			}, StatefulappReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
			)
		})

	})

	var _ = ginkgo.Describe("delete pod by zone pipeline", func() {

		delPodStatefulapp := testconverter.GetObjFromYaml("./data/statefulapp-01.yaml")
		statefulappNamespace := delPodStatefulapp.GetNamespace()
		statefulappName := delPodStatefulapp.GetName()

		ginkgo.It("step: update statefulapp for delete pod by zone", func() {
			err := testresource.TestClient.UpdateStatefulappInstance(delPodStatefulapp)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"update statefulapp succeed",
			)
		})

		ginkgo.It("step: wait for statefulapp updated", func() {
			time.Sleep(StatefulappCreateTimeout)
		})

		ginkgo.It("step: check pod status which is deleted", func() {
			// is zone1Pod1 deleted
			zone1Pod1 := fmt.Sprintf("%s-test-zone1-%d", statefulappName, 1)
			gomega.Eventually(func() bool {
				_, err := testresource.TestClient.IsPodExists(statefulappNamespace, zone1Pod1)
				if kubeerrors.IsNotFound(err) {
					return true
				}
				return false
			}, PodDeleteTimeout, TryInterval).Should(
				gomega.Equal(true),
				"pod %s is deleted", zone1Pod1,
			)
		})

		ginkgo.It("step: check statefulapp status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(statefulappNamespace, statefulappName)
			}, StatefulappReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
			)
		})

	})

	var _ = ginkgo.Describe("pod failover pipeline", func() {

		failOverStatefulapp := testconverter.GetObjFromYaml("./data/statefulapp-01.yaml")
		statefulappNamespace := failOverStatefulapp.GetNamespace()
		statefulappName := failOverStatefulapp.GetName()

		ginkgo.It("step: delete pod for failover", func() {
			// is zone1Pod0 deleted
			zone1Pod0 := fmt.Sprintf("%s-test-zone1-%d", statefulappName, 0)
			err := testresource.TestClient.DeletePod(statefulappNamespace, zone1Pod0)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"delete pod %s succeed", zone1Pod0,
			)
		})

		ginkgo.It("step: wait for pod failover", func() {
			time.Sleep(PodDeleteTimeout)
		})

		ginkgo.It("step: check pod failover status", func() {
			// is zone1Pod0 created
			zone1Pod0 := fmt.Sprintf("%s-test-zone1-%d", statefulappName, 0)
			_, err := testresource.TestClient.IsPodExists(statefulappNamespace, zone1Pod0)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create pod %s succeed", zone1Pod0,
			)
			// is zone1Pod0 running
			gomega.Eventually(func() corev1.PodPhase {
				return testresource.TestClient.GetPodStatus(statefulappNamespace, zone1Pod0)
			}, PodFailoverTimeout, TryInterval).Should(
				gomega.Equal(corev1.PodRunning),
				"pod %s is running", zone1Pod0,
			)
		})

		ginkgo.It("step: check statefulapp status", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeStatefulappInstanceIsReadyByObj(statefulappNamespace, statefulappName)
			}, StatefulappReadyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"statefulapp %s is ready", statefulappName,
			)
		})

	})

})
