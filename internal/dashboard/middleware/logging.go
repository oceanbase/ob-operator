/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package middleware

import (
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logging() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Starting time request
		startTime := time.Now()

		// Processing request
		ctx.Next()

		// End Time request
		endTime := time.Now()

		// execution time
		latencyTime := endTime.Sub(startTime)

		// Request method
		reqMethod := ctx.Request.Method

		// Request route
		reqUri := ctx.Request.RequestURI

		// Status code
		statusCode := ctx.Writer.Status()

		// Request ID
		requestID := requestid.Get(ctx)

		log.WithFields(log.Fields{
			"METHOD":      reqMethod,
			"REQUEST_URI": reqUri,
			"STATUS":      statusCode,
			"LATENCY":     latencyTime,
			"REQUEST_ID":  requestID,
		}).Info("[HTTP REQUEST]")

		ctx.Next()
	}
}
