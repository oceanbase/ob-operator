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

package models

type TelemetryRecord struct {
	IpHashes     []string `json:"ipHashes"`
	Timestamp    int64    `json:"timestamp"`
	Message      string   `json:"message"`
	Reason       string   `json:"reason"`
	ResourceType string   `json:"resourceType"`
	EventType    string   `json:"eventType"`
	Resource     any      `json:"resource,omitempty"`
	Extra        any      `json:"extra,omitempty"`
	Reporter     string   `json:"reporter"`
}

type ExtraField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TelemetryUploadBody struct {
	Content   TelemetryRecord `json:"content"`
	Time      string          `json:"time"`
	Component string          `json:"component"`
}
