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

package initialization

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/pkg/cable/server"
	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/util"
)

func InitApp() {
	// init logger
	InitLogger()

	util.FuncList = append(util.FuncList, StopApp)

	log.Info("init directory for oceanbase")
	// TODO use paths in dockerfile and remove this process
	InitDir()

	log.Info("init status variables")
	// TODO set variable values, move from observer package to a meaningful one
	status.Readiness = false
	status.ObserverStarted = false

	log.Info("init http server")
	server.CableServer.Init()
	go server.CableServer.Run()

}

func StopApp() {
	log.Info("stop cable server")
	server.CableServer.Stop(context.TODO())
}
