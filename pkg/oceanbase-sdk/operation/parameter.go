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
	"fmt"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
)

func (m *OceanbaseOperationManager) GetParameter(ctx context.Context, name string, scope *param.Scope) ([]model.Parameter, error) {
	parameters := make([]model.Parameter, 0)
	var err error
	if scope == nil {
		err = m.QueryList(ctx, &parameters, sql.QueryParameter, name)
	} else {
		queryParameterSql := fmt.Sprintf(sql.QueryParameterWithScope, scope.Name)
		err = m.QueryList(ctx, &parameters, queryParameterSql, name, scope.Value)
	}
	return parameters, err
}

func (m *OceanbaseOperationManager) SetParameter(ctx context.Context, name string, value any, scope *param.Scope) error {
	if scope == nil {
		setParameterSql := fmt.Sprintf(sql.SetParameter, name)
		return m.ExecWithDefaultTimeout(ctx, setParameterSql, value)
	}
	setParameterSql := fmt.Sprintf(sql.SetParameterWithScope, name, scope.Name)
	m.Logger.Info("SetParameterSql statement", "statement", setParameterSql)
	return m.ExecWithDefaultTimeout(ctx, setParameterSql, value, scope.Value)
}

func (m *OceanbaseOperationManager) SelectCompatibleOfTenants(ctx context.Context) ([]*model.Parameter, error) {
	parameters := make([]*model.Parameter, 0)
	var err error
	err = m.QueryList(ctx, &parameters, sql.SelectCompatibleOfTenants)
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
