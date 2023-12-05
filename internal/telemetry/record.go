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
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

func newRecordFromEvent(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...models.ExtraField) *models.TelemetryRecord {
	anno := annotations
	if len(extra) > 0 {
		if anno == nil {
			anno = make(map[string]string)
		}
		for _, field := range extra {
			anno[field.Key] = field.Value
		}
	}
	return &models.TelemetryRecord{
		Timestamp:    time.Now().Unix(),
		Message:      message,
		Reason:       reason,
		ResourceType: objectType,
		EventType:    eventType,
		Resource:     object,
		Extra:        anno,
	}
}

// Encode a TelemetryRecord into a io.ReadCloser
func encodeRecord(record *models.TelemetryRecord) (io.ReadCloser, error) {
	body := models.TelemetryUploadBody{
		Content:   *record,
		Time:      time.Unix(record.Timestamp, 0).Format(time.DateTime),
		Component: TelemetryComponent,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(bts)), nil
}
