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
