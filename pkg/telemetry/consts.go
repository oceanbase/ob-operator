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
	DefaultThrottlerBufferSize = 100
	DefaultWorkerCount         = 10
)

const (
	TelemetryReportDevHost  = "http://openwebapi.dev.alipay.net"
	TelemetryReportTestHost = "http://openwebapi.test.alipay.net"
	TelemetryReportProdHost = "https://openwebapi.ocenbase.com"
	TelemetryReportHost     = TelemetryReportProdHost
	TelemetryReportPath     = "/api/web/oceanbase/report"
)

const (
	TelemetryComponent = "ob-operator"
)

const (
	// devel use
	TelemetryRequestSignature = "dbe97393a695335d67de91dd4049ba"
)

const (
	ResourceBehaviorCreate = "Create"
	ResourceBehaviorUpdate = "Update"
	ResourceBehaviorDelete = "Delete"

	ResourceBehaviorError  = "Error"
	ResourceBehaviorNormal = "Normal"
)

const (
	DisableTelemetryEnvName = "DISABLE_TELEMETRY"
)
