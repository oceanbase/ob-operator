package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/requestid"
	logger "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
)

func SendSuccessfulResponse(c *gin.Context, data interface{}) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    "",
		Successful: true,
	}
	c.JSON(http.StatusOK, resp)
}

func SendBadRequestResponse(c *gin.Context, data interface{}, err error) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    fmt.Sprintf("bad request: %s", err.Error()),
		Successful: false,
	}
	c.JSON(http.StatusBadRequest, resp)
}

func SendUnauthorizedResponse(c *gin.Context, data interface{}, err error) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    fmt.Sprintf("unauthorized: %s", err.Error()),
		Successful: false,
	}
	c.JSON(http.StatusUnauthorized, resp)
}

func SendNotFoundResponse(c *gin.Context, data interface{}, err error) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    fmt.Sprintf("resource not found: %s", err.Error()),
		Successful: false,
	}
	c.JSON(http.StatusNotFound, resp)
}

func SendInternalServerErrorResponse(c *gin.Context, data interface{}, err error) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    fmt.Sprintf("internal server error: %s", err.Error()),
		Successful: false,
	}
	c.JSON(http.StatusInternalServerError, resp)
}

func SendNotImplementedResponse(c *gin.Context, data interface{}, err error) {
	resp := &response.APIResponse{
		Data:       data,
		Message:    fmt.Sprintf("not implemented: %s", err.Error()),
		Successful: false,
	}
	c.JSON(http.StatusNotImplemented, resp)
}

func logHandlerError(c *gin.Context, err error) {
	logger.
		WithField("Method", c.Request.Method).
		WithField("Request URI", c.Request.RequestURI).
		WithField("Request ID", requestid.Get(c)).
		WithError(err).
		Error("handler error")
}
