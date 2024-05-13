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
	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	biz "github.com/oceanbase/ob-operator/internal/dashboard/business/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID ListOBProxies
// @Summary list obproxies
// @Description list obproxies
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Param ns query string false "ns"
// @Success 200 object response.APIResponse{data=[]obproxy.OBProxyOverview}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies [GET]
// @Security ApiKeyAuth
func ListOBProxies(c *gin.Context) ([]obproxy.OBProxyOverview, error) {
	return biz.ListOBProxies(c, c.Query("ns"), metav1.ListOptions{})
}

// @ID CreateOBPROXY
// @Summary Create OBProxy
// @Description Create OBProxy with the specified parameters
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Param body body obproxy.CreateOBProxyParam true "Request body for creating obproxy"
// @Success 200 object response.APIResponse{data=obproxy.OBProxy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies [PUT]
// @Security ApiKeyAuth
func CreateOBProxy(c *gin.Context) (*obproxy.OBProxy, error) {
	param := &obproxy.CreateOBProxyParam{}
	err := c.BindJSON(param)
	if err != nil {
		return nil, httpErr.NewBadRequest("Failed to bind json, err msg: " + err.Error())
	}
	return biz.CreateOBProxy(c, param)
}

// @ID GetOBProxy
// @Summary Get OBProxy
// @Description Get OBProxy by namespace and name
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=obproxy.OBProxy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetOBProxy(c *gin.Context) (*obproxy.OBProxy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return biz.GetOBProxy(c, nn.Namespace, nn.Name)
}

// @ID PatchOBProxy
// @Summary Patch OBProxy
// @Description Patch OBProxy with the specified parameters
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Param body body obproxy.PatchOBProxyParam true "Request body for patching obproxy"
// @Success 200 object response.APIResponse{data=obproxy.OBProxy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies/{namespace}/{name} [PATCH]
// @Security ApiKeyAuth
func PatchOBProxy(c *gin.Context) (*obproxy.OBProxy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	param := &obproxy.PatchOBProxyParam{}
	err = c.BindJSON(param)
	if err != nil {
		return nil, httpErr.NewBadRequest("Failed to bind json, err msg: " + err.Error())
	}
	return biz.PatchOBProxy(c, nn.Namespace, nn.Name, param)
}

// @ID DeleteOBProxy
// @Summary Delete OBProxy
// @Description Delete OBProxy by namespace and name
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=obproxy.OBProxy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteOBProxy(c *gin.Context) (*obproxy.OBProxy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	return biz.DeleteOBProxy(c, nn.Namespace, nn.Name)
}
