package business

import (
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func GetRequestStatistics(store *store.SqlAuditStore, req model.RequestStatisticsRequest) (*model.RequestStatisticsResponse, error) {
	return store.QueryRequestStatistics(req)
}
