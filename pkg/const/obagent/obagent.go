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

package monitor

const (
	HttpPort  = 8088
	PprofPort = 8089
)

const (
	HttpPortName  = "http"
	PprofPortName = "pprof"
)

const (
	ProbeCheckPeriodSeconds = 2
	ProbeCheckDelaySeconds  = 5
)

const (
	ContainerName      = "obagent"
	InstallPath        = "/home/admin/obagent"
	ConfigPath         = "/home/admin/obagent/conf"
	StatUrl            = "/metrics/stat"
	MonitorUser        = "monitor"
	ConfigVolumeSuffix = "monitor-conf"
)

const (
	EnvClusterName     = "CLUSTER_NAME"
	EnvClusterId       = "CLUSTER_ID"
	EnvZoneName        = "ZONE_NAME"
	EnvMonitorUser     = "MONITOR_USER"
	EnvMonitorPASSWORD = "MONITOR_PASSWORD"
	EnvOBMonitorStatus = "OB_MONITOR_STATUS"
)

const (
	ActiveStatus = "active"
)
