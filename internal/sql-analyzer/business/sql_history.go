package business

import (
	"context"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetSqlHistoryInfo(ctx context.Context, auditStore *store.SqlAuditStore, req model.SqlHistoryRequest) (*model.SqlHistoryResponse, error) {
	return auditStore.QuerySqlHistoryInfo(req)
}
