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

package util

import (
	"runtime"
	"strings"

	"k8s.io/klog/v2"
)

func LogForOBClusterStatusConvert(funcName, clusterName, status, zoneName, zoneStatus string) {
	if funcName == "" {
		p, _, _, _ := runtime.Caller(2)
		tmp := strings.Split(runtime.FuncForPC(p).Name(), "/")
		funcName = tmp[len(tmp)-1]
	}
	klog.Infoln(funcName, "update OBCluster", clusterName, "to", status, "Zone", zoneName, "Status", zoneStatus)
}
