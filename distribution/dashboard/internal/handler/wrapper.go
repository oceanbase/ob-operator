package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/errors"
)

type Handler func(c *gin.Context) (interface{}, errors.ObError)

type Wrapper func(h Handler) gin.HandlerFunc

func ErrorWrapper(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h(c)
		statusCode := http.StatusOK
		var errMsg string
		if err != nil {
			statusCode = err.Status()
			errMsg = err.Error()
			// ensure that the response is nil
			res = nil
		}
		c.JSON(statusCode, &response.APIResponse{
			Data:       res,
			Message:    errMsg,
			Successful: err == nil,
		})
	}
}
