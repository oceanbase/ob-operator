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

	"github.com/oceanbase/ob-operator/pkg/util/system"
)

var Liveness bool
var Readiness bool

const GracefulTime = 10 * time.Second
const TickTime = 5 * time.Second

func CheckOBServeStatus() {
	time.Sleep(GracefulTime)
	// checker
	tick := time.Tick(TickTime)
	for {
		select {
		case <-tick:
			checkerOBServer()
		}
	}
}

func checkerOBServer() {
	name := ProcessObserver
	pm := &system.ProcessManager{}
	status := pm.ProcessIsRunningByName(name)
	if status {
		// update liveness
		Liveness = true
	} else {
		if Paused {
			// update liveness
			Liveness = true
		} else {
			log.Println("not Paused")
			system.Exit()
		}
	}
}
