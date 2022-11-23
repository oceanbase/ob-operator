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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/pkg/config/constant"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Router     *gin.Engine
	HttpServer *http.Server
}

var CableServer Server

func (S *Server) Init() {
	router := NewCustomRouter()
	systemGroup := router.Group("/api/system")
	{
		// status for nic & ip
		systemGroup.GET("/info", GetNicInfo)
		// paused
		systemGroup.POST("/paused", Paused)
		systemGroup.POST("/rework", Rework)
	}
	obGroup := router.Group("/api/ob")
	{
		// start
		obGroup.POST("/start", OBStart)
		// stop
		obGroup.POST("/stop", OBStop)
		// status
		obGroup.GET("/status", OBStatus)
		// readiness
		obGroup.GET("/readiness", OBReadiness)
		// readiness update
		obGroup.POST("/readinessUpdate", OBReadinessUpdate)
		// get version
		obGroup.GET("/veriosn", OBVersion)
	}
	S.Router = router
	S.HttpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", constant.CablePort),
		Handler: router,
	}
}

func (S *Server) Run() {
	err := S.HttpServer.ListenAndServe()
	if err != nil {
		log.WithError(err).Errorf("run server got exception: %v", err)
	}
}

func (S *Server) Stop(ctx context.Context) {
	S.HttpServer.Shutdown(ctx)
}

func NewCustomRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(GinLogger())
	router.Use(GinPanic())
	return router
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeStart := time.Now()
		c.Next()
		timeEnd := time.Now()
		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		latency := timeEnd.Sub(timeStart)
		comment := c.Errors
		log.Infof("request: from %s, method %s, path %s, response: status code %d, latency %s, comment %v", clientIP, method, path, statusCode, latency.String(), comment)
	}
}

func GinPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				resp := NewErrorResponse(errors.New(fmt.Sprintf("recover error %v", err)))
				SendResponse(c, resp)
			}
		}()
		c.Next()
	}
}

func SendResponse(c *gin.Context, resp *ApiResponse) {
	responseJSON, _ := json.Marshal(resp)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.String(resp.Code, string(responseJSON))
}
