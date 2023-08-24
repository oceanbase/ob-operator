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
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
)

func (m *OceanbaseOperationManager) QueryTenantWithName(tenantName string) ([]*model.OBTenant, error) {
	tenants := make([]*model.OBTenant, 0)
	err := m.QueryList(&tenants, sql.QueryTenantWithName, tenantName)
	if err != nil {
		m.Logger.Error(err, "Failed to query tenants")
		return nil, errors.Wrap(err, "Query tenants")
	}
	if len(tenants) == 0 {
		return nil, errors.Errorf("No tenants found")
	}
	return tenants, nil
}
