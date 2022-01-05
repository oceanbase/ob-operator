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
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Router     *gin.Engine
	HttpServer *http.Server
}

var Tiny Server

func (S *Server) Init() {
	router := RouteCustom()
	systemoperator := router.Group("/api/system")
	{
		// status for nic & ip
		systemoperator.GET("/info", Info)
		// paused
		systemoperator.POST("/paused", Paused)
		systemoperator.POST("/rework", Rework)
	}
	oboperator := router.Group("/api/ob")
	{
		// start
		oboperator.POST("/start", OBStart)
		// stop
		oboperator.POST("/stop", OBStop)
		// status
		oboperator.GET("/status", OBStatus)
		// readiness
		oboperator.GET("/readiness", OBReadiness)
		// readiness update
		oboperator.POST("/readinessUpdate", OBReadinessUpdate)
	}
	S.Router = router
	S.HttpServer = &http.Server{
		Addr:    ":19001",
		Handler: router,
	}
}

func (S *Server) Run() {
	err := S.HttpServer.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func (S *Server) Stop(ctx context.Context) {
	S.HttpServer.Shutdown(ctx)
}

func RouteCustom() *gin.Engine {
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
		log.Println(clientIP, statusCode, method, path, latency, comment)
	}
}

func GinPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Error Info] %s", err)
				// return
				data := make(map[string]interface{})
				Sender(c, 400, data)
			}
		}()
		c.Next()
	}
}

func Sender(c *gin.Context, code int, data interface{}) {
	responseJSON, _ := json.Marshal(data)
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.String(code, string(responseJSON))
}
