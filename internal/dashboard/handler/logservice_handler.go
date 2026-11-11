package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	biz "github.com/oceanbase/ob-operator/internal/dashboard/business/logservice"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"io"
	"net/http"
)

func bindLogService(c *gin.Context, p any) error {
	d := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024))
	d.DisallowUnknownFields()
	if err := d.Decode(p); err != nil {
		return oberr.NewBadRequest("Invalid LogService request: " + err.Error())
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return oberr.NewBadRequest("Request must contain one JSON object")
	}
	return nil
}

// @ID ListLogServices
// @Summary List LogServices
// @Tags LogService
// @Param namespace query string false "Namespace"
// @Success 200 object response.APIResponse{data=[]logservice.Item}
// @Router /api/v1/logservices [get]
// @Security ApiKeyAuth
func ListLogServices(c *gin.Context) ([]biz.Item, error) {
	return biz.Default().List(c.Request.Context(), c.Query("namespace"))
}

// @ID GetLogService
// @Summary Read LogService topology, nodes, PVCs, references and events
// @Tags LogService
// @Param namespace path string true "Namespace"
// @Param name path string true "Name"
// @Success 200 object response.APIResponse{data=logservice.Detail}
// @Router /api/v1/logservices/{namespace}/{name} [get]
// @Security ApiKeyAuth
func GetLogService(c *gin.Context) (*biz.Detail, error) {
	return biz.Default().Get(c.Request.Context(), c.Param("namespace"), c.Param("name"))
}

// @ID CreateLogService
// @Summary Create a LogService with existing namespace, storage classes and Secret
// @Tags LogService
// @Accept json
// @Param body body logservice.CreateRequest true "Creation configuration"
// @Success 200 object response.APIResponse{data=logservice.Item}
// @Router /api/v1/logservices [post]
// @Security ApiKeyAuth
func CreateLogService(c *gin.Context) (*biz.Item, error) {
	p := &biz.CreateRequest{}
	if err := bindLogService(c, p); err != nil {
		return nil, err
	}
	return biz.Default().Create(c.Request.Context(), p)
}

// @ID ScaleLogService
// @Summary Scale replicas in existing zones with optimistic concurrency
// @Tags LogService
// @Accept json
// @Param namespace path string true "Namespace"
// @Param name path string true "Name"
// @Param body body logservice.ScaleRequest true "Current resourceVersion and every existing zone replica count"
// @Success 200 object response.APIResponse{data=logservice.Item}
// @Router /api/v1/logservices/{namespace}/{name}/replicas [patch]
// @Security ApiKeyAuth
func ScaleLogService(c *gin.Context) (*biz.Item, error) {
	p := &biz.ScaleRequest{}
	if err := bindLogService(c, p); err != nil {
		return nil, err
	}
	return biz.Default().Scale(c.Request.Context(), c.Param("namespace"), c.Param("name"), p)
}

// @ID DeleteLogService
// @Summary Delete an unreferenced, unprotected LogService after exact-name confirmation
// @Tags LogService
// @Accept json
// @Param namespace path string true "Namespace"
// @Param name path string true "Name"
// @Param body body logservice.DeleteRequest true "Exact name and current resourceVersion"
// @Success 200 object response.APIResponse{data=bool}
// @Router /api/v1/logservices/{namespace}/{name} [delete]
// @Security ApiKeyAuth
func DeleteLogService(c *gin.Context) (bool, error) {
	p := &biz.DeleteRequest{}
	if err := bindLogService(c, p); err != nil {
		return false, err
	}
	return biz.Default().Delete(c.Request.Context(), c.Param("namespace"), c.Param("name"), p)
}
