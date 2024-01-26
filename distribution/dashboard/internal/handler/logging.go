package handler

import (
	"github.com/gin-contrib/requestid"
	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func logHandlerError(c *gin.Context, err error) {
	logger.
		WithField("Method", c.Request.Method).
		WithField("Request URI", c.Request.RequestURI).
		WithField("Request ID", requestid.Get(c)).
		WithError(err).
		Error("handler error")
}
