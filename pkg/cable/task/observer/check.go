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
	"time"

	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/oceanbase/ob-operator/pkg/util/shell"
	"github.com/oceanbase/ob-operator/pkg/util/system"
	log "github.com/sirupsen/logrus"
)

func CheckObserverLoop() {
	time.Sleep(constant.GracefulTime)
	// checker
	tick := time.Tick(constant.TickTime)
	for {
		select {
		case <-tick:
			checkerObserver()
		}
	}
}

func checkerObserver() {
	name := constant.ProcessObserver
	pm := &system.ProcessManager{}
	isRunning := pm.ProcessIsRunningByName(name)
	if isRunning {
		// update liveness
		status.Liveness = true
		status.ObserverProcessStarted = true
		status.ObserverProcessStartTimes = 0
	} else {
		if status.Paused {
			// update liveness
			log.Info("observer process not running, but in paused status")
			status.Liveness = true
		} else {
			if !status.ObserverProcessStarted {
				if status.ObserverProcessStartTimes >= constant.StartObserverRetryTimes {
					log.Errorf("start observer has retried %d times, and still failed, quit", constant.StartObserverRetryTimes)
					system.Exit()
				}
				log.Error("observer process not running, try restart...")
				// TODO: config sleep time and retry times with parameters
				time.Sleep(constant.GracefulTime)
				_, err := shell.NewCommand(constant.OBSERVER_START_COMMAND_WITHOUT_PARAM).WithContext(context.TODO()).WithUser(shell.AdminUser).Execute()
				if err != nil {
					log.WithError(err).Errorf("restart observer command exec error", err)
				}
				status.ObserverProcessStartTimes += 1
			}
			status.ObserverProcessStarted = false
		}
	}
}
