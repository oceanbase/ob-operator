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
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

func (m *OceanbaseOperationManager) ListUnitsWithServerIP(serverIP string) ([]*model.OBUnit, error) {
	units := make([]*model.OBUnit, 0)
	err := m.QueryList(&units, sql.ListUnitsWithServerIP, serverIP)
	if err != nil {
		m.Logger.Error(err, "Failed to list ob units")
		return nil, errors.Wrap(err, "List OB units")
	}
	return units, nil
}
