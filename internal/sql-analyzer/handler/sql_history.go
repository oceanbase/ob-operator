package handler

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
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
		return nil, err
	}

	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "/data"
	}

	auditStore, err := store.NewSqlAuditStore(c.Request.Context(), filepath.Join(dataPath, "sql_audit"))
	if err != nil {
		return nil, err
	}
	defer auditStore.Close()

	return business.GetSqlHistoryInfo(c.Request.Context(), auditStore, req)
}
