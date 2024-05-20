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

import "time"

const (
	DefaultThrottlerBufferSize  = 30
	DefaultThrottlerWorkerCount = 30

	DefaultFilterSize          = 1000
	DefaultFilterExpireTimeout = 60 * time.Minute
)

var TelemetryReportHost = TelemetryReportProdHost
var TelemetryReportScheme = SchemeHttps
var TelemetryDisabled = false

var (
	ThrottlerBufferSize  = DefaultThrottlerBufferSize
	ThrottlerWorkerCount = DefaultThrottlerWorkerCount

	FilterSize          = DefaultFilterSize
	FilterExpireTimeout = DefaultFilterExpireTimeout
)

func SetTelemetryReportHost(host string) {
	TelemetryReportHost = host
}

func SetTelemetryReportScheme(scheme string) {
	TelemetryReportScheme = scheme
}

func SetTelemetryDisabled(disabled bool) {
	TelemetryDisabled = disabled
}

func SetThrottlerBufferSize(size int) {
	ThrottlerBufferSize = size
}

func SetThrottlerWorkerCount(count int) {
	ThrottlerWorkerCount = count
}

func SetFilterSize(size int) {
	FilterSize = size
}

func SetFilterExpireTimeout(timeout time.Duration) {
	FilterExpireTimeout = timeout
}
