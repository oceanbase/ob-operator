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

package kube

import (
	"k8s.io/klog/v2"

	"github.com/oceanbase/ob-operator/pkg/util"
)

func LogForUniversal(msg string, obj interface{}) {
	klog.Infoln(msg, util.CovertToJSON(obj))
}

func LogForAppActionStatus(kind, appName, action string, obj interface{}) {
	klog.Infoln(action, kind, appName, util.CovertToJSON(obj))
}
