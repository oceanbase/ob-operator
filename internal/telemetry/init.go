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
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func init() {
	TelemetryDisabled = os.Getenv(DisableTelemetryEnvName) == "true"
	if host, ok := os.LookupEnv(TelemetryReportHostEnvName); ok && host != "" && strings.HasPrefix(host, "http") {
		if u, err := url.Parse(host); err == nil {
			clt := http.Client{
				Timeout: time.Second,
			}
			_, err := clt.Head(u.String())
			if err == nil {
				TelemetryReportScheme = u.Scheme
				TelemetryReportHost = u.Host
			}
		}
	}
}
