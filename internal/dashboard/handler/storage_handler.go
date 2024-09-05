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

	"github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID GetFile
// @Summary Get file
// @Description Get file by id
// @Tags Storage
// @Accept application/json
// @Produce application/zip
// @Success 200 {file} FileResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/storage/{id} [GET]
// @Security ApiKeyAuth
func GetFile(_ *gin.Context) (bool, error) {
	return true, errors.NewNotImplemented("")
}
