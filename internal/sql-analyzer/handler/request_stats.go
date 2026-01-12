package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/business"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

// @Summary Get request statistics
// @Description Get request statistics
// @Tags sql-analyzer
// @Accept json
// @Produce json
// @Param tenant_name path string true "Tenant Name"
// @Param request body model.RequestStatisticsRequest true "Query parameters"
// @Success 200 {object} model.APIResponse{data=model.RequestStatisticsResponse} "A list of aggregated SQL statistics"
// @Failure 400 {object} model.APIResponse "Error: Invalid request"
// @Failure 500 {object} model.APIResponse "Error: Internal server error"
// @Router /api/v1/tenants/{tenant_name}/request-stats [post]
func GetRequestStatistics(c *gin.Context) (*model.RequestStatisticsResponse, error) {
	var req model.RequestStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// TODO: The data path should be configurable.
	auditStore, err := store.NewSqlAuditStore(c.Request.Context(), "/data/sql_audit")
	if err != nil {
		return nil, err
	}
	defer auditStore.Close()

	return business.GetRequestStatistics(auditStore, req)
}
