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

package ob

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/oceanbase/ob-operator/pkg/util/shell"
)

const (
	CPU_COUNT                       = 16
	MEMORY_LIMIT                    = 10
	MEMORY_LOW                      = "8"
	MEMORY_SIMPLE                   = 64
	NIC                             = "eth0"
	OBSERVER_MYSQL_PORT             = "2881"
	OBSERVER_RPC_PORT               = "2882"
	OBSERVER_START_COMMAND_TEMPLATE = "cd /home/admin/oceanbase; ulimit -s 10240; ulimit -c unlimited; LD_LIBRARY_PATH=/home/admin/oceanbase/lib:$LD_LIBRARY_PATH LD_PRELOAD='' /home/admin/oceanbase/bin/observer --appname ${OB_CLUSTER_NAME} --cluster_id ${OB_CLUSTER_ID} --zone ${ZONE_NAME} --devname ${DEV_NAME} -p 2881 -P 2882 -d /home/admin/oceanbase/store/ -l info -o 'rootservice_list=${RS_LIST},config_additional_dir=/home/admin/oceanbase/etc2,/home/admin/oceanbase/etc3,${OPTION}'"
)

type StartObServerProcessArguments struct {
	ClusterName string `json:"clusterName" binding:"required"`
	ClusterId   int    `json:"clusterId" binding:"required"`
	ZoneName    string `json:"zoneName" binding:"required"`
	RsList      string `json:"rsList" binding:"required"`
	CpuLimit    int    `json:"cpuLimit" binding:"required"`
	MemoryLimit int    `json:"memoryLimit" binding:"required"`
}

func StartOBServerProcess(param StartObServerProcessArguments) {
	cpu := getCPU(param.CpuLimit)
	memory := getMemory(param.MemoryLimit)
	systemMemory := getSystemMemory(param.MemoryLimit)
	datafileSize := getDatafileSize(memory)
	var cmd string
	var option string
	obClusterName := param.ClusterName
	obClusterId := param.ClusterId
	zoneName := param.ZoneName
	rsList := param.RsList
	deviceName := NIC
	memoryInt, err := strconv.Atoi(memory)
	if memoryInt <= MEMORY_SIMPLE {
		// 2C, 10G
		option = fmt.Sprintf("cpu_count=%s,memory_limit=%sG,system_memory=%sG,__min_full_resource_pool_memory=268435456,datafile_size=%sG,net_thread_count=%d,stack_size=512K,cache_wash_threshold=1G,schema_history_expire_time=1d,enable_separate_sys_clog=false,enable_merge_by_turn=false,enable_syslog_recycle=true,enable_syslog_wf=false,max_syslog_file_count=4", cpu, memory, systemMemory, datafileSize, param.CpuLimit)
	} else {
		// 16C, 64G
		option = fmt.Sprintf("cpu_count=%s,memory_limit=%sG,system_memory=%sG,__min_full_resource_pool_memory=1073741824,datafile_size=%sG,net_thread_count=%d", cpu, memory, systemMemory, datafileSize, param.CpuLimit)
	}
	cmd = replaceAll(OBSERVER_START_COMMAND_TEMPLATE, startObServerParamReplacer(obClusterName, obClusterId, zoneName, deviceName, rsList, option))
	_, err = shell.NewCommand(cmd).WithContext(context.TODO()).WithUser(shell.AdminUser).Execute()
	if err != nil {
		log.Println("cmd exec error", err)
	}
}

func getCPU(cpuLimit int) string {
	if CPU_COUNT+1 > cpuLimit {
		return strconv.Itoa(CPU_COUNT)
	}
	return strconv.Itoa(cpuLimit - 1)
}

func getMemory(memoryLimit int) string {
	if memoryLimit == MEMORY_LIMIT {
		log.Println("memoryLimit is low")
		return MEMORY_LOW
	}
	return strconv.Itoa(int(float32(memoryLimit) * 0.9))
}

func getSystemMemory(memory int) string {
	tmp := float32(memory) * 0.9
	var coefficient float32
	if tmp >= 150 {
		coefficient = 0.3
	} else if tmp >= 64 && tmp < 150 {
		coefficient = 0.4
	} else {
		coefficient = 0.5
	}
	return strconv.Itoa(int(tmp * coefficient))
}

func getDatafileSize(memoryLimit string) string {
	tmp, _ := strconv.Atoi(memoryLimit)
	return strconv.Itoa(tmp * 3)
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
