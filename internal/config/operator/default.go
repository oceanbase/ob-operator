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

package operator

import (
	"github.com/spf13/viper"

	oc "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/database"
)

var defaultConfigMap = map[string]any{
	"namespace":                 "",
	"manager-namespace":         "oceanbase-system",
	"metric-addr":               ":8080",
	"leader-elect":              false,
	"health-probe-bind-address": ":8081",
	"log-verbosity":             0,
	"disable-webhooks":          false,

	"task.debug":    false,
	"task.poolSize": 10000,

	"telemetry.disabled": false,
	"telemetry.debug":    false,
	"telemetry.host":     "https://openwebapi.oceanbase.com",

	"database.connectionLRUCacheSize": database.DefaultLRUCacheSize,

	"resource.defaultDiskExpandPercent":  oc.DefaultDiskExpandPercent,
	"resource.defaultLogPercent":         oc.DefaultLogPercent,
	"resource.initialDataDiskUsePercent": oc.InitialDataDiskUsePercent,
	"resource.defaultDiskUsePercent":     oc.DefaultDiskUsePercent,
	"resource.defaultMemoryLimitPercent": oc.DefaultMemoryLimitPercent,
	"resource.defaultMemoryLimitSize":    oc.DefaultMemoryLimitSize,
	"resource.defaultDatafileMaxSize":    oc.DefaultDatafileMaxSize,
	"resource.defaultDatafileNextSize":   oc.DefaultDatafileNextSize,
	"resource.minMemorySize":             oc.MinMemorySizeS,
	"resource.minDataDiskSize":           oc.MinDataDiskSizeS,
	"resource.minRedoLogDiskSize":        oc.MinRedoLogDiskSizeS,
	"resource.minLogDiskSize":            oc.MinLogDiskSizeS,

	"time.tenantOpRetryTimes":               oc.TenantOpRetryTimes,
	"time.tenantOpRetryGapSeconds":          oc.TenantOpRetryGapSeconds,
	"time.taskMaxRetryTimes":                oc.TaskMaxRetryTimes,
	"time.taskRetryBackoffThreshold":        oc.TaskRetryBackoffThreshold,
	"time.probeCheckPeriodSeconds":          oc.ProbeCheckPeriodSeconds,
	"time.probeCheckDelaySeconds":           oc.ProbeCheckDelaySeconds,
	"time.getConnectionMaxRetries":          oc.GetConnectionMaxRetries,
	"time.checkConnectionInterval":          oc.CheckConnectionInterval,
	"time.checkJobInterval":                 oc.CheckJobInterval,
	"time.checkJobMaxRetries":               oc.CheckJobMaxRetries,
	"time.commonCheckInterval":              oc.CommonCheckInterval,
	"time.bootstrapTimeoutSeconds":          oc.BootstrapTimeoutSeconds,
	"time.localityChangeTimeoutSeconds":     oc.LocalityChangeTimeoutSeconds,
	"time.defaultStateWaitTimeout":          oc.DefaultStateWaitTimeout,
	"time.timeConsumingStateWaitTimeout":    oc.TimeConsumingStateWaitTimeout,
	"time.waitForJobTimeoutSeconds":         oc.WaitForJobTimeoutSeconds,
	"time.serverDeleteTimeoutSeconds":       oc.ServerDeleteTimeoutSeconds,
	"time.tolerateServerPodNotReadyMinutes": oc.TolerateServerPodNotReadyMinutes,
}

func setDefaultConfigs(vp *viper.Viper) {
	for k, v := range defaultConfigMap {
		vp.SetDefault(k, v)
	}
}
