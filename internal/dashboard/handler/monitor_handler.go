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

	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/internal/dashboard/business/monitor"
)

// @ID ListEndpoints
// @Summary list all endpoints
// @Description list all endpoints to query metrics
// @Tags Metric
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.MonitorEndpoint}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/metrics/endpoints [GET]
// @Security ApiKeyAuth
func ListEndpoints(c *gin.Context) {
	endpoints, err := monitor.ListEndpoints(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
	}
	c.JSON(http.StatusOK, endpoints)
}
