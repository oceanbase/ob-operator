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

const (
	TelemetryReportProdHost = "openwebapi.oceanbase.com" // https

	TelemetryReportPath = "/api/web/oceanbase/report"
)

const (
	SchemeHttp  = "http"
	SchemeHttps = "https"
)

const ContentTypeJson = "application/json"

const (
	TelemetryComponent          = "ob-operator"
	TelemetryComponentDashboard = "oceanbase-dashboard"
)

const (
	ResourceBehaviorCreate = "Create"
	ResourceBehaviorUpdate = "Update"
	ResourceBehaviorDelete = "Delete"

	ResourceBehaviorError  = "Error"
	ResourceBehaviorNormal = "Normal"
)

const (
	DisableTelemetryEnvName    = "DISABLE_TELEMETRY"
	TelemetryDebugEnvName      = "TELEMETRY_DEBUG"
	TelemetryReportHostEnvName = "TELEMETRY_REPORT_HOST"
)

const (
	ObjectTypeUnknown  = "Unknown"
	ObjectTypeOperator = "Operator"
)
