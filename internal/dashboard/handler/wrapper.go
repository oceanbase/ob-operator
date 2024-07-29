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

package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	logger "github.com/sirupsen/logrus"

	acbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/ac"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/pkg/errors"
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

func Wrap[T any](h Handler[T], funcs ...acbiz.EnforceFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(funcs) > 0 {
			for _, f := range funcs {
				ok, err := f(c)
				if err != nil {
					logHandlerError(c, err)
					c.JSON(http.StatusInternalServerError, &response.APIResponse{
						Data:       nil,
						Message:    err.Error(),
						Successful: false,
					})
					return
				}
				if !ok {
					c.JSON(http.StatusForbidden, &response.APIResponse{
						Data:       nil,
						Message:    "forbidden",
						Successful: false,
					})
					return
				}
			}
		}
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
