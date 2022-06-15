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
	"time"

    log "github.com/sirupsen/logrus"
	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/util/system"
)

func StopProcess() {
	name := constant.ProcessObserver
	pm := &system.ProcessManager{}
	err := pm.TerminateProcessByName(name)
	if err != nil {
		log.WithError(err).Errorf("terminate observer process got error %v", err)
	}
	time.Sleep(2 * time.Second)
	err = pm.KillProcessByName(name)
	if err != nil {
		log.WithError(err).Errorf("kill observer process got error %v", err)
	}
	status.ObserverStarted = false
}
