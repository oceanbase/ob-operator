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

package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/pkg/cable/status"
	"github.com/oceanbase/ob-operator/pkg/cable/task/observer"
	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/oceanbase/ob-operator/pkg/util"
	"github.com/oceanbase/ob-operator/pkg/util/system"
)

func GetNicInfo(c *gin.Context) {
	data, err := system.GetNICInfo(constant.NIC)
	if err != nil {
		// TODO find a better way, panic just quit the process and the container should be recreated
		panic(err)
	}
	SendResponse(c, NewSuccessResponse(data))
}

func Paused(c *gin.Context) {
	// TODO send status map as a response
	status.Paused = true
	SendResponse(c, NewSuccessResponse(status.Paused))
}

func Rework(c *gin.Context) {
	// TODO send status map as a response
	status.Paused = false
	SendResponse(c, NewSuccessResponse(status.Paused))
}

func OBStart(c *gin.Context) {
	param := new(observer.StartObServerProcessArguments)
	if err := c.ShouldBind(&param); err != nil {
		panic(err)
	}
	log.Infof("start observer with param %s", util.CovertToJSON(param))

	if !status.ObserverStarted {
		if observer.ValidateStartParam(*param) {
			go observer.StartObserverProcess(*param)
			go observer.CheckObserverLoop()
			status.ObserverStarted = true
			SendResponse(c, NewSuccessResponse(status.ObserverStarted))
		} else {
			SendResponse(c, NewErrorResponse(errors.New("param is invalid")))
		}
	} else {
		SendResponse(c, NewErrorResponse(errors.New("observer already started")))
	}
}

func OBUpgradeRoute(c *gin.Context) {
	param := new(observer.OBUpgradeRouteParam)
	param.TargetVersion = c.Query(observer.TargetVersion)
	param.CurrentVersion = c.Query(observer.CurrentVersion)
	log.Infof("get upgrade route from V%s to V%s", param.CurrentVersion, param.TargetVersion)
	res, err := observer.GetOBUpgradeRoute(*param)
	if err != nil {
		SendResponse(c, NewErrorResponse(err))
	} else {
		SendResponse(c, NewSuccessResponse(res))
	}
}

func OBStop(c *gin.Context) {
	go observer.StopProcess()
	SendResponse(c, NewSuccessResponse("successful"))
}

func OBStatus(c *gin.Context) {
	if status.Liveness {
		SendResponse(c, NewSuccessResponse(status.Liveness))
	} else {
		SendResponse(c, NewErrorResponse(errors.New(fmt.Sprintf("liveness is %v", status.Liveness))))
	}
}

func OBRecoverConfig(c *gin.Context) {
	err := observer.RecoverConfig()
	if err != nil {
		SendResponse(c, NewErrorResponse(err))
	} else {
		SendResponse(c, NewSuccessResponse("successful"))
	}
}

func OBVersion(c *gin.Context) {
	res, err := observer.GetObVersion()
	if err != nil {
		SendResponse(c, NewErrorResponse(err))
	} else {
		SendResponse(c, NewSuccessResponse(res.Output))
	}
}

func OBReadiness(c *gin.Context) {
	if status.Readiness {
		SendResponse(c, NewSuccessResponse(status.Readiness))
	} else {
		SendResponse(c, NewErrorResponse(errors.New(fmt.Sprintf("readiness is %v", status.Readiness))))
	}
}

func OBReadinessUpdate(c *gin.Context) {
	status.Readiness = true
	SendResponse(c, NewSuccessResponse(status.Readiness))
}
