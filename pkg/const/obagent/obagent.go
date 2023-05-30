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
	EnvZoneName        = "Zone_NAME"
	EnvMonitorUser     = "MONITOR_USER"
	EnvMonitorPASSWORD = "MONITOR_PASSWORD"
	EnvOBMonitorStatus = "OB_MONITOR_STATUS"
)

const (
	ActiveStatus = "active"
)
