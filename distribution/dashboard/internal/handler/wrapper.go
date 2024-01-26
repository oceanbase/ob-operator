package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	"github.com/oceanbase/oceanbase-dashboard/pkg/errors"
)

type Handler[T any] func(c *gin.Context) (T, error)

func W[T any](h Handler[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h(c)
		statusCode := http.StatusOK
		var errMsg string
		if err != nil {
			if obe := err.(errors.ObError); obe != nil {
				statusCode = obe.Status()
			} else {
				statusCode = http.StatusInternalServerError
			}
			errMsg = err.Error()
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
