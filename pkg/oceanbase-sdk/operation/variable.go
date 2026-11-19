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

package operation

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) SetGlobalVariable(ctx context.Context, name string, value any) error {
	setGlobalVariableSql := fmt.Sprintf(sql.SetGlobalVariable, name)
	return m.ExecWithDefaultTimeout(ctx, setGlobalVariableSql, value)
}

func (m *OceanbaseOperationManager) GetGlobalVariable(ctx context.Context, name string) (*model.Variable, error) {
	variable := &model.Variable{}
	err := m.QueryRow(ctx, variable, sql.GetGlobalVariable, name)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1146 { // ER_NO_SUCH_TABLE
		err = m.QueryRow(ctx, variable, sql.GetLegacyGlobalVariable, name)
	}
	return variable, err
}

// GetTenantConnectWhiteList reads the tenant's persisted value through the cluster's sys session.
func (m *OceanbaseOperationManager) GetTenantConnectWhiteList(ctx context.Context, tenantID int64) (string, error) {
	return readTenantConnectWhiteList(ctx, m.QueryRow, tenantID)
}

func readTenantConnectWhiteList(ctx context.Context, queryRow func(context.Context, any, string, ...any) error, tenantID int64) (string, error) {
	var value string
	err := queryRow(ctx, &value, sql.GetTenantConnectWhiteList, tenantID)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1146 {
		err = queryRow(ctx, &value, sql.GetLegacyTenantConnectWhiteList, tenantID)
	}
	return value, err
}
