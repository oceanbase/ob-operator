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

package provider

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/pkg/cable/observer"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/ob"
	"github.com/oceanbase/ob-operator/pkg/util"
	"github.com/oceanbase/ob-operator/pkg/util/system"
)

func Info(c *gin.Context) {
	data, err := system.GetNICInfo(ob.NIC)
	if err != nil {
		panic(err)
	}
	Sender(c, 200, data)
}

func Paused(c *gin.Context) {
	observer.Paused = true
	log.Println("Paused is", observer.Paused)
	data := make(map[string]interface{})
	Sender(c, 200, data)
}

func Rework(c *gin.Context) {
	observer.Paused = false
	log.Println("Paused is", observer.Paused)
	data := make(map[string]interface{})
	Sender(c, 200, data)
}

func OBStart(c *gin.Context) {
	param := new(ob.StartObServerProcessArguments)
	if err := c.ShouldBind(&param); err != nil {
		panic(err)
	}
	log.Println(util.CovertToJSON(param))

	if !observer.OBStarted {
		go ob.StartOBServerProcess(*param)
		go observer.CheckOBServeStatus()
		observer.OBStarted = true
		data := make(map[string]interface{})
		Sender(c, 200, data)
	} else {
		data := make(map[string]interface{})
		Sender(c, 400, data)
	}
}

func OBStop(c *gin.Context) {
	go observer.StopProcess()
	data := make(map[string]interface{})
	Sender(c, 200, data)
}

func OBStatus(c *gin.Context) {
	data := make(map[string]interface{})
	if observer.Liveness {
		Sender(c, 200, data)
	} else {
		Sender(c, 400, data)
	}
}

func OBReadiness(c *gin.Context) {
	data := make(map[string]interface{})
	if observer.Readiness {
		Sender(c, 200, data)
	} else {
		Sender(c, 400, data)
	}
}

func OBReadinessUpdate(c *gin.Context) {
	data := make(map[string]interface{})
	observer.Readiness = true
	Sender(c, 200, data)
}
