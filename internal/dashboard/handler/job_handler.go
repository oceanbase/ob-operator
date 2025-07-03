/*
Copyright (c) 2025 OceanBase
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

	jobbiz "github.com/oceanbase/ob-operator/internal/dashboard/business/job"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
)

// @ID GetJob
// @Summary Get job
// @Description Get job by id
// @Tags Job
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=job.Job}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "namespace of the job"
// @Param name path string true "name of the job"
// @Router /api/v1/jobs/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetJob(c *gin.Context) (*job.Job, error) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	return jobbiz.GetJob(c.Request.Context(), namespace, name)
}

// @ID DeleteJob
// @Summary Delete a job
// @Description Delete a job by id
// @Tags Job
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=bool}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "namespace of the job"
// @Param name path string true "name of the job"
// @Router /api/v1/jobs/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteJob(c *gin.Context) (bool, error) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	return true, jobbiz.DeleteJob(c.Request.Context(), namespace, name)
}
