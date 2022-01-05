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

package cable

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/util"
)

func CableStatusCheckExecuter(podIP string) error {
	url := fmt.Sprintf("%s%s:%d%s", observerconst.CableUrlProfix, podIP, observerconst.CablePort, observerconst.CableInfoUrl)
	code, _ := util.HTTPGET(url)
	if code != 200 {
		return errors.New("cable get ip failed")
	}
	return nil
}

func OBServerStartExecuter(podIP string, obServerStartArgs map[string]interface{}) {
	url := fmt.Sprintf("%s%s:%d%s", observerconst.CableUrlProfix, podIP, observerconst.CablePort, observerconst.CableStartUrl)
	code, _ := util.HTTPPOST(url, util.CovertToJSON(obServerStartArgs))
	if code != 200 {
		// TODO: need to check, why the OBServer start failed
		klog.Errorln("start observer", podIP, "failed", util.CovertToJSON(obServerStartArgs))
	} else {
		klog.Infoln("start observer", podIP, "succeed", util.CovertToJSON(obServerStartArgs))
	}
}

func OBServerStatusCheckExecuter(clusterName, podIP string) error {
	url := fmt.Sprintf("%s%s:%d%s", observerconst.CableUrlProfix, podIP, observerconst.CablePort, observerconst.CableStatusUrl)
	code, _ := util.HTTPGET(url)
	if code != 200 {
		klog.Errorln("OBCluster", clusterName, "observer", podIP, "starting not ready")
		return errors.New("wait for OBCluster starting")
	}
	return nil
}

func CableReadinessUpdateExecuter(podIP string) error {
	url := fmt.Sprintf("%s%s:%d%s", observerconst.CableUrlProfix, podIP, observerconst.CablePort, observerconst.CableReadinessUpdateUrl)
	code, _ := util.HTTPPOST(url, "")
	if code != 200 {
		return errors.New("update cable readiness failed")
	}
	return nil
}
