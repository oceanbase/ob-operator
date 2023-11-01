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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Telemetry", func() {
	Context("Test Telemetry", Label("metrics"), func() {
		It("Test LocalIP", func() {
			ips, err := LocalIP()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(ips).ShouldNot(BeNil())
		})

		It("Test K8sNodes", func() {
			nodes, err := K8sNodes()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(nodes).ShouldNot(BeNil())
		})

		It("Test TelemetryEnvMetrics", func() {
			metrics, err := GetHostMetrics()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metrics).ShouldNot(BeNil())
			fmt.Printf("%+v\n", metrics)
		})

		It("Cancel context multiple times", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			cancel()
			select {
			case <-ctx.Done():
				Expect(ctx.Err()).ShouldNot(BeNil())
			}
		})
	})
})
