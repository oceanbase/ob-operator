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

package telemetry

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obcfg "github.com/oceanbase/ob-operator/internal/config/operator"
)

type fakeEventRecorder struct{}

func (f *fakeEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	getLogger().Printf("Event: %+v, %s, %s, %s\n", object, eventtype, reason, message)
}
func (f *fakeEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...any) {
	getLogger().Printf("Eventf: %+v, %s, %s, %s, %v\n", object, eventtype, reason, messageFmt, args)
}
func (f *fakeEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...any) {
	getLogger().Printf("AnnotatedEventf: %+v, %+v, %s, %s, %s, %v\n", object, annotations, eventtype, reason, messageFmt, args)
}

var _ = Describe("Telemetry", Label("telemetry"), Ordered, func() {
	var telemetry Recorder
	tenant := &v1alpha1.OBTenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OBTenant",
			APIVersion: "oceanbase.oceanbase.com/v1alpha1",
		},
	}

	BeforeAll(func() {
		telemetry = NewRecorder(context.TODO(), &fakeEventRecorder{})
		Expect(telemetry).ShouldNot(BeNil())
	})

	AfterAll(func() {
		By("Wait for telemetry to finish, watch the output")
		time.Sleep(2 * time.Second)
		telemetry.Done()
	})

	It("Empty pointer", func() {
		Expect(tenant).NotTo(BeNil())
		Expect(tenant.GetObjectKind().GroupVersionKind().Kind).Should(Equal("OBTenant"))
	})

	It("GetHostMetrics", func() {
		metrics := telemetry.GetHostMetrics()
		if obcfg.GetConfig().Telemetry.Disabled {
			Expect(metrics).Should(BeNil())
		} else {
			Expect(metrics).ShouldNot(BeNil())
		}
	})

	It("Event", func() {
		telemetry.Event(tenant, "event", "some reasons", "test")
	})

	It("Eventf", func() {
		telemetry.Eventf(tenant, "eventf", "some reasons", "hello %s", "world")
	})

	It("AnnotatedEventf", func() {
		annos := map[string]string{
			"hello": "world",
		}
		telemetry.AnnotatedEventf(tenant, annos, "annotatedEventf", "some reasons", "hello %s", "world")
	})

	It("Generate recorder directly", func() {
		tenant := v1alpha1.OBTenant{
			Spec: v1alpha1.OBTenantSpec{
				ClusterName:      "test",
				TenantName:       "t1",
				UnitNumber:       1,
				Charset:          "test",
				Collate:          "test",
				ConnectWhiteList: "test",
				TenantRole:       "STANDBY",
				Credentials: v1alpha1.TenantCredentials{
					Root:      "root",
					StandbyRO: "standby-ro",
				},
			}}
		telemetry.GenerateTelemetryRecord(tenant, "arbitrary", "generateTelemetryRecord", "some reason", "test", nil)
	})
})
