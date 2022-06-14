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
	"log"
	"time"

	"github.com/oceanbase/ob-operator/pkg/cable/config/constant"
	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/util/system"
)

func StopProcess() {
	name := constant.ProcessObserver
	pm := &system.ProcessManager{}
	err := pm.TerminateProcessByName(constant.ProcessObserver)
	if err != nil {
		log.Println(err)
	}
	time.Sleep(2 * time.Second)
	err = pm.KillProcessByName(name)
	if err != nil {
		log.Println(err)
	}
	status.ObserverStarted = false
}
