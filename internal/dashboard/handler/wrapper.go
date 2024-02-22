package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/errors"
	logger "github.com/sirupsen/logrus"
)

type Handler[T any] func(c *gin.Context) (T, error)

func logHandlerError(c *gin.Context, err error) {
	logger.
		WithField("Method", c.Request.Method).
		WithField("Request URI", c.Request.RequestURI).
		WithField("Request ID", requestid.Get(c)).
		WithError(err).
		Error("handler error")
}

func Wrap[T any](h Handler[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h(c)
		statusCode := http.StatusOK
		var errMsg string
		if err != nil {
			if obe, ok := err.(errors.ObError); ok && obe != nil {
				statusCode = obe.Status()
			} else {
				statusCode = http.StatusInternalServerError
			}
			errMsg = err.Error()
			logHandlerError(c, err)
			// ensure that the response is nil
			res = *new(T)
		}
		c.JSON(statusCode, &response.APIResponse{
			Data:       res,
			Message:    errMsg,
			Successful: err == nil,
		})
	}
}
