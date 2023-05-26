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
package tenant

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/model"
	testconverter "github.com/oceanbase/ob-operator/test/e2e/converter"
	testresource "github.com/oceanbase/ob-operator/test/e2e/resource"
	testutils "github.com/oceanbase/ob-operator/test/e2e/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("tenant manager test pipeline", func() {

	myconfig.ClusterName = "cn"

	var _ = ginkgo.Describe("create obcluster pipeline", func() {
		var obServerList []model.AllServer
		var service corev1.Service
		obcluster := testconverter.GetObjFromYaml("./data/obcluster-3-2-v314.yaml")
		obclusterNamespace := obcluster.GetNamespace()
		obclusterName := obcluster.GetName()
		statefulappName := fmt.Sprintf("sapp-%s", obcluster.GetName())
		serviceName := fmt.Sprintf("svc-%s", obcluster.GetName())

		ginkgo.It("step: create obcluster", func() {
			err := testresource.TestClient.CreateObj(obcluster)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create obcluster succeed",
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

		ginkgo.It("step: check obcluster status", func() {
			gomega.Eventually(func() string {
				return testresource.TestClient.GetOBStatusClusterStatus(obclusterNamespace, obclusterName, myconfig.ClusterName)
			}, OBClusterReadyimeout, TryInterval).Should(
				gomega.Equal(observerconst.ClusterReady),
				"obcluster %s ready", obclusterName,
			)
		})
	})

	var _ = ginkgo.Describe("create tenant pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-small-unit1-3zone-3F-priority333.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: create tenant object", func() {
			err := testresource.TestClient.CreateObj(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"create tenant object succeed",
			)
		})

		ginkgo.It("step: check tenant create succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantCreateTimeout, TryInterval).Should(
				gomega.Equal(true),
				"create tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant resource unit pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit1-3zone-3F-priority333.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant resource unit", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant resource unit every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant resource unit update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant resource unit modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantResourceUnitIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant resource unit succeed",
			)
		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant primary zone pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit1-3zone-3F-priority123.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant primary zone", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant primary zone every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant primary zone update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant primary zone modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantPrimaryZoneIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant primary zone succeed",
			)
		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant locality pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit1-3zone-2F1R-priority123.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant locality", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant locality every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant locality update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant locality modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantLocalityIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant locality succeed",
			)
		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant unit num pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit2-3zone-2F1R-priority123.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant unit num ", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant unit num every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant unit num update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant unit num modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantUnitNumIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant unit num succeed",
			)
		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant delete zone pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit2-2zone-2F-priority123.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant delete zone ", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant delete zone every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant delete zone update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant delete zone modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantResourceUnitIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant resource unit succeed",
			)

			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantPrimaryZoneIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant primary zone succeed",
			)

			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantUnitNumIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant delete zone succeed",
			)

		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

	var _ = ginkgo.Describe("modify tenant add zone pipeline", func() {

		tenant := testconverter.GetObjFromYaml("./data/tenant/tenant-large-unit2-3zone-2F1R-priority123.yaml")
		tenantName := tenant.GetName()
		tenantNamespace := tenant.GetNamespace()

		ginkgo.It("step: wait before modify tenant add zone ", func() {
			time.Sleep(ApplyWaitTime)
		})

		ginkgo.It("step: modify tenant add zone every zone", func() {
			err := testresource.TestClient.UpdateTenantInstance(tenant)
			gomega.Expect(err).To(
				gomega.BeNil(),
				"modify tenant add zone update tenant succeed",
			)
		})

		ginkgo.It("step: check tenant add zone modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantResourceUnitIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant resource unit succeed",
			)

			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantPrimaryZoneIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant primary zone succeed",
			)

			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantUnitNumIsMatched(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant add zone succeed",
			)

		})

		ginkgo.It("step: check tenant modify succeed", func() {
			gomega.Eventually(func() bool {
				return testresource.TestClient.JudgeTenantInstanceIsRunningByObj(tenantNamespace, tenantName)
			}, TenantModifyTimeout, TryInterval).Should(
				gomega.Equal(true),
				"modify tenant succeed",
			)
		})
	})

})
