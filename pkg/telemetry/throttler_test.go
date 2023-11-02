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
	"fmt"
	"io"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/pkg/telemetry/models"
)

var _ = Describe("Telemetry throttler", Ordered, Label("throttler"), func() {
	var throttler *throttler

	BeforeAll(func() {
		throttler = getThrottler()
		Expect(throttler).ShouldNot(BeNil())
	})

	AfterAll(func() {
		throttler.close()
	})

	It("Send telemetry record", func() {
		res, err := throttler.sendTelemetryRecord(&models.TelemetryRecord{
			IpHashes:     []string{},
			Timestamp:    time.Now().Unix(),
			Message:      "dev",
			ResourceType: "dev",
			EventType:    "test",
			Resource:     nil,
			Extra:        nil,
		})
		Expect(err).ShouldNot(HaveOccurred())
		bts, err := io.ReadAll(res.Body)
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Printf("%s\n", string(bts))
	})

	It("Send telemetry record", func() {
		res, err := throttler.sendTelemetryRecord(&models.TelemetryRecord{
			IpHashes:     []string{},
			Timestamp:    time.Now().Unix(),
			Message:      "dev",
			ResourceType: "dev",
			EventType:    "test",
			Resource: map[string]interface{}{
				"test":     "test",
				"ips":      []string{"ip1", "ip2"},
				"k8sNodes": []models.K8sNode{{}, {}},
			},
			Extra: nil,
		})
		Expect(err).ShouldNot(HaveOccurred())
		bts, err := io.ReadAll(res.Body)
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Printf("%s\n", string(bts))
	})
})
