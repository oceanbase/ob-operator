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

	"github.com/oceanbase/ob-operator/internal/dashboard/model/obproxy"
)

// @ID ListOBProxies
// @Summary list obproxies
// @Description list obproxies
// @Tags OBProxy
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]obproxy.OBProxyOverview}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obproxies [GET]
// @Security ApiKeyAuth
func ListOBProxies(_ *gin.Context) ([]obproxy.OBProxyOverview, error) {
	return nil, nil
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
func CreateOBProxy(_ *gin.Context) (*obproxy.OBProxy, error) {
	return nil, nil
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
func GetOBProxy(_ *gin.Context) (*obproxy.OBProxy, error) {
	return nil, nil
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
func PatchOBProxy(_ *gin.Context) (*obproxy.OBProxy, error) {
	return nil, nil
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
func DeleteOBProxy(_ *gin.Context) (*obproxy.OBProxy, error) {
	return nil, nil
}
