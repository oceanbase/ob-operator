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
)

type TelemetryRecord struct {
	IpHashes     []string `json:"ipHashes"`
	Timestamp    int64    `json:"timestamp"`
	Message      string   `json:"message"`
	ResourceType string   `json:"resourceType"`
	EventType    string   `json:"eventType"`
	Resource     any      `json:"resource,omitempty"`
	Extra        any      `json:"extra,omitempty"`

	K8sNodes []K8sNode `json:"k8sNodes,omitempty"`
}

type ExtraField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type K8sNode struct {
	KernelVersion           string `json:"kernelVersion,omitempty"`
	OsImage                 string `json:"osImage,omitempty"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion,omitempty"`
	KubeletVersion          string `json:"kubeletVersion,omitempty"`
	KubeProxyVersion        string `json:"kubeProxyVersion,omitempty"`
	OperatingSystem         string `json:"operatingSystem,omitempty"`
	Architecture            string `json:"architecture,omitempty"`
}

type TelemetryUploadBody struct {
	Content   TelemetryRecord `json:"content"`
	Time      string          `json:"time"`
	Component string          `json:"component"`
}

func newTelemetryRecord(object any, objectType, eventType, reason, message string, annotations map[string]string, extra ...ExtraField) *TelemetryRecord {
	anno := annotations
	if len(extra) > 0 {
		if anno == nil {
			anno = make(map[string]string)
		}
		for _, field := range extra {
			anno[field.Key] = field.Value
		}
	}
	return &TelemetryRecord{
		Timestamp:    time.Now().Unix(),
		Message:      message,
		ResourceType: objectType,
		EventType:    eventType,
		Resource:     object,
		Extra:        anno,
	}
}

// Encode a TelemetryRecord into a io.ReadCloser
func encodeTelemetryRecord(record *TelemetryRecord) (io.ReadCloser, error) {
	body := TelemetryUploadBody{
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
