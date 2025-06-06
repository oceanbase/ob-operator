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
	"os"

	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = Describe("Config", func() {
	GinkgoHelper()

	Context("default", func() {
		It("should return default config", func() {
			output := Config{}
			Expect(mapstructure.Decode(defaultConfigMap, &output)).To(Succeed())

			got := newConfig()
			Expect(got.Database.ConnectionLRUCacheSize).To(BeEquivalentTo(defaultConfigMap["database.connectionLRUCacheSize"]))
			Expect(got.Resource.DefaultDiskExpandPercent).To(BeEquivalentTo(defaultConfigMap["resource.defaultDiskExpandPercent"]))
			Expect(got.Resource.DefaultLogPercent).To(BeEquivalentTo(defaultConfigMap["resource.defaultLogPercent"]))
			Expect(got.Resource.InitialDataDiskUsePercent).To(BeEquivalentTo(defaultConfigMap["resource.initialDataDiskUsePercent"]))
			Expect(got.Resource.DefaultDiskUsePercent).To(BeEquivalentTo(defaultConfigMap["resource.defaultDiskUsePercent"]))
			Expect(got.Resource.DefaultMemoryLimitPercent).To(BeEquivalentTo(defaultConfigMap["resource.defaultMemoryLimitPercent"]))
			Expect(got.Resource.DefaultMemoryLimitSize).To(BeEquivalentTo(defaultConfigMap["resource.defaultMemoryLimitSize"]))
			Expect(got.Resource.DefaultDatafileMaxSize).To(BeEquivalentTo(defaultConfigMap["resource.defaultDatafileMaxSize"]))
			Expect(got.Resource.DefaultDatafileNextSize).To(BeEquivalentTo(defaultConfigMap["resource.defaultDatafileNextSize"]))
			Expect(got.Resource.MinMemorySize).To(BeEquivalentTo(defaultConfigMap["resource.minMemorySize"]))
			Expect(got.Resource.MinDataDiskSize).To(BeEquivalentTo(defaultConfigMap["resource.minDataDiskSize"]))
			Expect(got.Resource.MinRedoLogDiskSize).To(BeEquivalentTo(defaultConfigMap["resource.minRedoLogDiskSize"]))
			Expect(got.Resource.MinLogDiskSize).To(BeEquivalentTo(defaultConfigMap["resource.minLogDiskSize"]))
			Expect(got.Time.TenantOpRetryTimes).To(BeEquivalentTo(defaultConfigMap["time.tenantOpRetryTimes"]))
			Expect(got.Time.TenantOpRetryGapSeconds).To(BeEquivalentTo(defaultConfigMap["time.tenantOpRetryGapSeconds"]))
			Expect(got.Time.TaskMaxRetryTimes).To(BeEquivalentTo(defaultConfigMap["time.taskMaxRetryTimes"]))
			Expect(got.Time.TaskRetryBackoffThreshold).To(BeEquivalentTo(defaultConfigMap["time.taskRetryBackoffThreshold"]))
			Expect(got.Time.ProbeCheckPeriodSeconds).To(BeEquivalentTo(defaultConfigMap["time.probeCheckPeriodSeconds"]))
			Expect(got.Time.ProbeCheckDelaySeconds).To(BeEquivalentTo(defaultConfigMap["time.probeCheckDelaySeconds"]))
			Expect(got.Time.GetConnectionMaxRetries).To(BeEquivalentTo(defaultConfigMap["time.getConnectionMaxRetries"]))
			Expect(got.Time.CheckConnectionInterval).To(BeEquivalentTo(defaultConfigMap["time.checkConnectionInterval"]))
			Expect(got.Time.CheckJobInterval).To(BeEquivalentTo(defaultConfigMap["time.checkJobInterval"]))
			Expect(got.Time.CheckJobMaxRetries).To(BeEquivalentTo(defaultConfigMap["time.checkJobMaxRetries"]))
			Expect(got.Time.CommonCheckInterval).To(BeEquivalentTo(defaultConfigMap["time.commonCheckInterval"]))
			Expect(got.Time.BootstrapTimeoutSeconds).To(BeEquivalentTo(defaultConfigMap["time.bootstrapTimeoutSeconds"]))
			Expect(got.Time.LocalityChangeTimeoutSeconds).To(BeEquivalentTo(defaultConfigMap["time.localityChangeTimeoutSeconds"]))
			Expect(got.Time.DefaultStateWaitTimeout).To(BeEquivalentTo(defaultConfigMap["time.defaultStateWaitTimeout"]))
			Expect(got.Time.TimeConsumingStateWaitTimeout).To(BeEquivalentTo(defaultConfigMap["time.timeConsumingStateWaitTimeout"]))
			Expect(got.Time.WaitForJobTimeoutSeconds).To(BeEquivalentTo(defaultConfigMap["time.waitForJobTimeoutSeconds"]))
			Expect(got.Time.ServerDeleteTimeoutSeconds).To(BeEquivalentTo(defaultConfigMap["time.serverDeleteTimeoutSeconds"]))
			Expect(got.Telemetry.Disabled).To(BeEquivalentTo(defaultConfigMap["telemetry.disabled"]))
			Expect(got.Telemetry.Reporter).To(BeEquivalentTo(defaultConfigMap["telemetry.reporter"]))
			Expect(got.Telemetry.Debug).To(BeEquivalentTo(defaultConfigMap["telemetry.debug"]))
			Expect(got.Telemetry.Host).To(BeEquivalentTo(defaultConfigMap["telemetry.host"]))
			Expect(got.Telemetry.ThrottlerBufferSize).To(BeEquivalentTo(defaultConfigMap["telemetry.throttlerBufferSize"]))
			Expect(got.Telemetry.ThrottlerWorkerCount).To(BeEquivalentTo(defaultConfigMap["telemetry.throttlerWorkerCount"]))
			Expect(got.Telemetry.FilterSize).To(BeEquivalentTo(defaultConfigMap["telemetry.filterSize"]))
			Expect(got.Telemetry.FilterExpireTimeout).To(BeEquivalentTo(defaultConfigMap["telemetry.filterExpireTimeout"]))
			Expect(got.Task.Debug).To(BeEquivalentTo(defaultConfigMap["task.debug"]))
			Expect(got.Task.PoolSize).To(BeEquivalentTo(defaultConfigMap["task.poolSize"]))
			Expect(got.Manager.DisableWebhooks).To(BeEquivalentTo(defaultConfigMap["disable-webhooks"]))
			Expect(got.Manager.LogVerbosity).To(BeEquivalentTo(defaultConfigMap["log-verbosity"]))
		})
	})

	Context("envVars", func() {
		BeforeEach(func() {
			os.Setenv("OB_OPERATOR_TASK_POOLSIZE", "9876")
			os.Setenv("OB_OPERATOR_TIME_TASKMAXRETRYTIMES", "1234")
			os.Setenv("OB_OPERATOR_TELEMETRY_DISABLED", "true")
			os.Setenv("OB_OPERATOR_DATABASE_CONNECTIONLRUCACHESIZE", "999")
		})
		AfterEach(func() {
			os.Unsetenv("OB_OPERATOR_TASK_POOLSIZE")
			os.Unsetenv("OB_OPERATOR_TIME_TASKMAXRETRYTIMES")
		})
		It("should return config with envVars", func() {
			Expect(os.Getenv("OB_OPERATOR_TASK_POOLSIZE")).To(Equal("9876"))
			got := newConfig()
			Expect(got.Task.PoolSize).To(BeEquivalentTo(9876))
			Expect(got.Time.TaskMaxRetryTimes).To(Equal(1234))
			Expect(got.Telemetry.Disabled).To(BeTrue())
			Expect(got.Database.ConnectionLRUCacheSize).To(Equal(999))
		})
	})

	Context("flags", func() {
		It("should return config with flags", func() {
			var namespace string
			var managerNamespace string
			var metricsAddr string
			var enableLeaderElection bool
			var probeAddr string
			var logVerbosity int
			pflag.StringVar(&namespace, "namespace", "", "The namespace to run oceanbase, default value is empty means all.")
			pflag.StringVar(&managerNamespace, "manager-namespace", "oceanbase-system", "The namespace to run manager tools.")
			pflag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
			pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
			pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
				"Enable leader election for controller manager. "+
					"Enabling this will ensure there is only one active controller manager.")
			pflag.IntVar(&logVerbosity, "log-verbosity", 0, "Log verbosity level, 0 is info, 1 is debug, 2 is trace")
			Expect(pflag.CommandLine.Parse([]string{
				"--log-verbosity", "1",
			})).To(Succeed())
			GinkgoLogr.Info("logVerbosity", "logVerbosity", logVerbosity)

			got := newConfig()
			Expect(got.Manager.LogVerbosity).To(Equal(1))
		})
	})
})
