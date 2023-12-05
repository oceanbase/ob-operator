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
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

var _ = Describe("Telemetry record", func() {
	It("Marshal TelemetryBody", func() {
		record := &models.TelemetryRecord{
			IpHashes:     []string{},
			Timestamp:    time.Now().Unix(),
			Message:      "",
			ResourceType: "",
			EventType:    "",
			Resource:     nil,
			Extra:        nil,
		}
		body := models.TelemetryUploadBody{
			Content:   *record,
			Time:      time.Unix(record.Timestamp, 0).Format(time.DateTime),
			Component: TelemetryComponent,
		}
		_, err := json.Marshal(body)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Marshal TelemetryRecord", func() {
		var err error
		var body []byte
		record := &models.TelemetryRecord{
			IpHashes:     []string{},
			Timestamp:    time.Now().Unix(),
			Message:      "",
			ResourceType: "",
			EventType:    "",
			Resource:     nil,
			Extra:        nil,
		}
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.IpHashes = nil
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.IpHashes = []string{"test", "test2", "test3"}
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.Extra = "test"
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.Extra = []string{"test", "test2", "test3"}
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.Extra = []models.ExtraField{
			{
				Key:   "test",
				Value: "value",
			},
		}
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		m := make(map[string]string)
		m["test"] = "value"
		record.Extra = m
		_, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())

		record.Resource = v1alpha1.OBCluster{}
		body, err = json.Marshal(record)
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Printf("%s\n", string(body))
	})
})
