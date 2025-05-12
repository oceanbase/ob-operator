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

	"github.com/oceanbase/ob-operator/internal/dashboard/model/job"
	"github.com/oceanbase/ob-operator/pkg/errors"
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
// @Router /api/v1/jobs/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetJob(_ *gin.Context) (*job.Job, error) {
	return nil, errors.NewNotImplemented("")
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
// @Router /api/v1/jobs/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteJob(_ *gin.Context) (bool, error) {
	return true, errors.NewNotImplemented("")
}
