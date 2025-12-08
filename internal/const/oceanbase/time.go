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

package oceanbase

const (
	TenantOpRetryTimes      = 9
	TenantOpRetryGapSeconds = 9
)

const (
	TaskMaxRetryTimes         = 99
	TaskRetryBackoffThreshold = 16
)

const (
	ProbeCheckPeriodSeconds = 2
	ProbeCheckDelaySeconds  = 5
	GetConnectionMaxRetries = 100
	CheckConnectionInterval = 3
	CheckJobInterval        = 3
	CheckJobMaxRetries      = 100
	CommonCheckInterval     = 5
)

const (
	BootstrapTimeoutSeconds       = 2100
	LocalityChangeTimeoutSeconds  = 86400 // 1 day
	DefaultStateWaitTimeout       = 1800
	TimeConsumingStateWaitTimeout = 3600
	WaitForJobTimeoutSeconds      = 7200
	ServerDeleteTimeoutSeconds    = 604800 // 7 days
)

const (
	TolerateServerPodNotReadyMinutes = 5
)
