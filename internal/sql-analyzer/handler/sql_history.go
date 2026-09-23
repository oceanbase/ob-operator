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
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
	httperr "github.com/oceanbase/ob-operator/pkg/errors"
)

// @ID GetSqlHistoryInfo
// @Summary Get SQL history info
// @Description Get SQL history info
// @Tags SQL
// @Accept application/json
// @Produce application/json
// @Param body body model.SqlHistoryRequest true "sql history request"
// @Success 200 {object} model.SqlHistoryResponse
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /api/v1/tenants/{tenant_name}/sql-history [POST]
func GetSqlHistoryInfo(c *gin.Context) (*model.SqlHistoryResponse, error) {
	var req model.SqlHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, httperr.NewBadRequest("invalid SQL history request")
	}

	resp, err := business.GetSqlHistoryInfo(c.Request.Context(), store.GetSqlAuditStore(), req)
	if err != nil {
		if obErr, ok := err.(httperr.ObError); ok && obErr.IsType(httperr.ErrBadRequest) {
			return nil, err
		}
		// Preserve diagnostic detail in server logs, not in the public response.
		logHandlerError(c, err)
		return nil, httperr.NewInternal("failed to query SQL history")
	}
	return resp, nil
}
