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
	"errors"
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/telemetry/models"
)

func newRecordFromEvent(object any, reporter, objectType, eventType, reason, message string, annotations map[string]string, extra ...models.ExtraField) *models.TelemetryRecord {
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
		Reporter:     reporter,
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

func extractUID(resource any) (string, error) {
	switch resource.(type) {
	case
		*v1alpha1.OBCluster,
		*v1alpha1.OBZone,
		*v1alpha1.OBServer,
		*v1alpha1.OBParameter,
		*v1alpha1.OBTenant,
		*v1alpha1.OBTenantBackupPolicy,
		*v1alpha1.OBTenantBackup,
		*v1alpha1.OBTenantRestore,
		*v1alpha1.OBTenantOperation,
		*v1alpha1.OBResourceRescue:
		meta, ok := resource.(metav1.Object)
		if !ok {
			return "", errors.New("failed to extract UID from resource")
		}
		return string(meta.GetUID()), nil
	default:
		return "", errors.New("unsupported resource type")
	}
}
