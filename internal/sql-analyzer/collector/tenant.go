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

package collector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
)

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, connMgr *oceanbase.ConnectionManager, tenantName string) (uint64, error) {
	manager, err := connMgr.GetSysReadonlyConnection()
	if err != nil {
		return 0, fmt.Errorf("failed to get connection for tenant ID retrieval: %w", err)
	}
	var tenant model.Tenant
	err = manager.QueryRow(ctx, &tenant, sqlconst.GetTenantIDByName, tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}
