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
	"context"

    log "github.com/sirupsen/logrus"
	"github.com/oceanbase/ob-operator/pkg/util/shell"
	"github.com/oceanbase/ob-operator/pkg/util/system"
	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/config/constant"
)


func CheckObserverStatus() {
	time.Sleep(constant.GracefulTime)
	// checker
	tick := time.Tick(constant.TickTime)
	for {
		select {
		case <-tick:
			checkerOBServer()
		}
	}
}

func checkerOBServer() {
	name := constant.ProcessObserver
	pm := &system.ProcessManager{}
	isRunning := pm.ProcessIsRunningByName(name)
	if isRunning {
		// update liveness
		status.Liveness = true
	} else {
		if status.Paused {
			// update liveness
			log.Info("observer process not running, but in paused status")
			status.Liveness = true
		} else {
			log.Error("observer process not running, try restart...")
            // TODO: config sleep time with parameters
	        time.Sleep(constant.GracefulTime)
            _, err := shell.NewCommand(constant.OBSERVER_START_COMMAND_WITHOUT_PARAM).WithContext(context.TODO()).WithUser(shell.AdminUser).Execute()
            if err != nil {
                log.Println("cmd exec error", err)
            }
			// system.Exit()
		}
	}
}
