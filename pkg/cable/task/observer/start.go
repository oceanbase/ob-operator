/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/oceanbase/ob-operator/pkg/config/constant"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/util/shell"
	log "github.com/sirupsen/logrus"
)

type StartObServerProcessArguments struct {
	ClusterName      string      `json:"clusterName" binding:"required"`
	ClusterId        int         `json:"clusterId" binding:"required"`
	ZoneName         string      `json:"zoneName" binding:"required"`
	RsList           string      `json:"rsList" binding:"required"`
	CpuLimit         int         `json:"cpuLimit" binding:"required"`
	MemoryLimit      int         `json:"memoryLimit" binding:"required"`
	CustomParameters []Parameter `json:"customParameters"`
	Version          string      `json:"Version" `
}

type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ValidateStartParam(param StartObServerProcessArguments) bool {
	if len(param.RsList) == 0 {
		log.Error("RsList is empty")
		return false
	}
	return true
}

func StartObserverProcess(param StartObServerProcessArguments) {
	cpu := getCPU(param.CpuLimit)
	memory := getMemoryLimit(param.MemoryLimit)
	systemMemory := getSystemMemory(param.MemoryLimit)
	datafileSize := getDatafileSize(memory)
	var cmd string
	var option string
	obClusterName := param.ClusterName
	obClusterId := param.ClusterId
	zoneName := param.ZoneName
	rsList := param.RsList
	version := param.Version
	deviceName := constant.NIC
	customOption := ""
	if param.CustomParameters != nil && len(param.CustomParameters) > 0 {
		for _, p := range param.CustomParameters {
			customOption = fmt.Sprintf("%s,%s=%s", customOption, p.Name, p.Value)
		}
	}
	if memory <= constant.MEMORY_SIMPLE {
		// 2C, 10G
		if version != "" && version[0:1] == observerconst.OBClusterV3 {
			option = fmt.Sprintf("cpu_count=%d,memory_limit=%dG,system_memory=%dG,__min_full_resource_pool_memory=1073741824,datafile_size=%dG,net_thread_count=%d,stack_size=512K,cache_wash_threshold=1G,schema_history_expire_time=1d,enable_separate_sys_clog=false,enable_merge_by_turn=false,enable_syslog_recycle=true,enable_syslog_wf=false,max_syslog_file_count=4%s", cpu, memory, systemMemory, datafileSize, param.CpuLimit, customOption)
		} else {
			option = fmt.Sprintf("cpu_count=%d,memory_limit=%dG,system_memory=%dG,__min_full_resource_pool_memory=1073741824,datafile_size=%dG,net_thread_count=%d,stack_size=512K,cache_wash_threshold=1G,schema_history_expire_time=1d,enable_syslog_recycle=true,enable_syslog_wf=false,max_syslog_file_count=4%s", cpu, memory, systemMemory, datafileSize, param.CpuLimit, customOption)
		}
	} else {
		// 16C, 64G
		option = fmt.Sprintf("cpu_count=%d,memory_limit=%dG,system_memory=%dG,__min_full_resource_pool_memory=1073741824,datafile_size=%dG,net_thread_count=%d%s", cpu, memory, systemMemory, datafileSize, param.CpuLimit, customOption)
	}
	cmd = replaceAll(constant.OBSERVER_START_COMMAND_TEMPLATE, startObServerParamReplacer(obClusterName, obClusterId, zoneName, deviceName, rsList, option))
	_, err := shell.NewCommand(cmd).WithContext(context.TODO()).WithUser(shell.AdminUser).Execute()
	if err != nil {
		log.WithError(err).Errorf("start observer command exec error %v", err)
	}
}

func getCPU(cpuLimit int) int {
	if constant.CPU_COUNT+1 > cpuLimit {
		return constant.CPU_COUNT
	}
	return cpuLimit - 1
}

func getMemoryLimit(memoryLimit int) int {
	if memoryLimit <= constant.MEMORY_LIMIT {
		log.Errorf("memoryLimit is low")
		return constant.MEMORY_LOW
	}
	return int(float32(memoryLimit) * 0.9)
}

func getSystemMemory(memory int) int {
	tmp := float32(memory) * 0.9
	var coefficient float32
	if tmp >= 150 {
		coefficient = 0.3
	} else if tmp >= 64 && tmp < 150 {
		coefficient = 0.4
	} else {
		coefficient = 0.5
	}
	return int(tmp * coefficient)
}

func getDatafileSize(memoryLimit int) int {
	return memoryLimit * 3
}

func replaceAll(template string, replacers ...*strings.Replacer) string {
	s := template
	for _, replacer := range replacers {
		s = replacer.Replace(s)
	}
	return s
}

func startObServerParamReplacer(obClusterName string, obClusterId int, zoneName, deviceName, rsList, option string) *strings.Replacer {
	return strings.NewReplacer("${OB_CLUSTER_NAME}", obClusterName, "${OB_CLUSTER_ID}", strconv.Itoa(obClusterId), "${ZONE_NAME}", zoneName, "${DEV_NAME}", deviceName, "${RS_LIST}", rsList, "${OPTION}", option)
}
