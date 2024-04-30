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

import "github.com/spf13/viper"

type Config struct {
	v *viper.Viper

	Manager   Manager   `mapstructure:",squash" yaml:"manager"`
	Database  Database  `mapstructure:"database" yaml:"database"`
	Task      Task      `mapstructure:"task" yaml:"task"`
	Telemetry Telemetry `mapstructure:"telemetry" yaml:"telemetry"`
	Time      Time      `mapstructure:"time" yaml:"time"`
	Resource  Resource  `mapstructure:"resource" yaml:"resource"`
}

type Manager struct {
	Namespace        string `mapstructure:"namespace" yaml:"namespace"`
	ManagerNamespace string `mapstructure:"manager-namespace" yaml:"managerNamespace"`
	MetricsAddr      string `mapstructure:"metrics-bind-address" yaml:"metricsAddr"`
	LeaderElect      bool   `mapstructure:"leader-elect" yaml:"enableElect"`
	ProbeAddr        string `mapstructure:"health-probe-bind-address" yaml:"probeAddr"`
	LogVerbosity     int    `mapstructure:"log-verbosity" yaml:"logVerbosity"`
	DisableWebhooks  bool   `mapstructure:"disable-webhooks" yaml:"disableWebhooks"`
}

type Task struct {
	Debug    bool `mapstructure:"debug" yaml:"debug"`
	PoolSize int  `mapstructure:"poolSize" yaml:"poolSize"`
}

type Telemetry struct {
	Disabled bool   `mapstructure:"disabled" yaml:"disabled"`
	Debug    bool   `mapstructure:"debug" yaml:"debug"`
	Host     string `mapstructure:"host" yaml:"host"`
}

type Database struct {
	ConnectionLRUCacheSize int `mapstructure:"connectionLRUCacheSize" yaml:"connectionLRUCacheSize"`
}

type Resource struct {
	DefaultDiskExpandPercent  int `mapstructure:"defaultDiskExpandPercent" yaml:"defaultDiskExpandPercent"`
	DefaultLogPercent         int `mapstructure:"defaultLogPercent" yaml:"defaultLogPercent"`
	InitialDataDiskUsePercent int `mapstructure:"initialDataDiskUsePercent" yaml:"initialDataDiskUsePercent"`
	DefaultDiskUsePercent     int `mapstructure:"defaultDiskUsePercent" yaml:"defaultDiskUsePercent"`
	DefaultMemoryLimitPercent int `mapstructure:"defaultMemoryLimitPercent" yaml:"defaultMemoryLimitPercent"`

	DefaultMemoryLimitSize  string `mapstructure:"defaultMemoryLimitSize" yaml:"defaultMemoryLimitSize"`
	DefaultDatafileMaxSize  string `mapstructure:"defaultDatafileMaxSize" yaml:"defaultDatafileMaxSize"`
	DefaultDatafileNextSize string `mapstructure:"defaultDatafileNextSize" yaml:"defaultDatafileNextSize"`

	MinMemorySize      string `mapstructure:"minMemorySize" yaml:"minMemorySizeQ"`
	MinDataDiskSize    string `mapstructure:"minDataDiskSize" yaml:"minDataDiskSizeQ"`
	MinRedoLogDiskSize string `mapstructure:"minRedoLogDiskSize" yaml:"minRedoLogDiskSizeQ"`
	MinLogDiskSize     string `mapstructure:"minLogDiskSize" yaml:"minLogDiskSizeQ"`
}

type Time struct {
	TenantOpRetryTimes      int `mapstructure:"tenantOpRetryTimes" yaml:"tenantOpRetryTimes"`
	TenantOpRetryGapSeconds int `mapstructure:"tenantOpRetryGapSeconds" yaml:"tenantOpRetryGapSeconds"`

	TaskMaxRetryTimes         int `mapstructure:"taskMaxRetryTimes" yaml:"taskMaxRetryTimes"`
	TaskRetryBackoffThreshold int `mapstructure:"taskRetryBackoffThreshold" yaml:"taskRetryBackoffThreshold"`

	ProbeCheckPeriodSeconds int `mapstructure:"probeCheckPeriodSeconds" yaml:"probeCheckPeriodSeconds"`
	ProbeCheckDelaySeconds  int `mapstructure:"probeCheckDelaySeconds" yaml:"probeCheckDelaySeconds"`
	GetConnectionMaxRetries int `mapstructure:"getConnectionMaxRetries" yaml:"getConnectionMaxRetries"`
	CheckConnectionInterval int `mapstructure:"checkConnectionInterval" yaml:"checkConnectionInterval"`
	CheckJobInterval        int `mapstructure:"checkJobInterval" yaml:"checkJobInterval"`
	CheckJobMaxRetries      int `mapstructure:"checkJobMaxRetries" yaml:"checkJobMaxRetries"`
	CommonCheckInterval     int `mapstructure:"commonCheckInterval" yaml:"commonCheckInterval"`

	BootstrapTimeoutSeconds       int `mapstructure:"bootstrapTimeoutSeconds" yaml:"bootstrapTimeoutSeconds"`
	LocalityChangeTimeoutSeconds  int `mapstructure:"localityChangeTimeoutSeconds" yaml:"localityChangeTimeoutSeconds"`
	DefaultStateWaitTimeout       int `mapstructure:"defaultStateWaitTimeout" yaml:"defaultStateWaitTimeout"`
	TimeConsumingStateWaitTimeout int `mapstructure:"timeConsumingStateWaitTimeout" yaml:"timeConsumingStateWaitTimeout"`
	WaitForJobTimeoutSeconds      int `mapstructure:"waitForJobTimeoutSeconds" yaml:"waitForJobTimeoutSeconds"`
	ServerDeleteTimeoutSeconds    int `mapstructure:"serverDeleteTimeoutSeconds" yaml:"serverDeleteTimeoutSeconds"`

	TolerateServerPodNotReadyMinutes int `mapstructure:"tolerateServerPodNotReadyMinutes" yaml:"tolerateServerPodNotReadyMinutes"`
}
