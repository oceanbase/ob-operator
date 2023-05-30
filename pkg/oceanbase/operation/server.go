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
	"fmt"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
)

func (m *OceanbaseOperationManager) GetServer(s *model.ServerInfo) (*model.OBServer, error) {
	observer := &model.OBServer{}
	err := m.QueryRow(observer, sql.GetServer, s.Ip, s.Port)
	if err != nil {
		return nil, errors.Wrap(err, "Get observer failed")
	}
	return observer, nil
}

func (m *OceanbaseOperationManager) AddServer(serverInfo *model.ServerInfo) error {
	server := fmt.Sprintf("%s:%d", serverInfo.Ip, serverInfo.Port)
	err := m.ExecWithDefaultTimeout(sql.AddServer, server)
	if err != nil {
		m.Logger.Error(err, "Got exception when add server")
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteServer(serverInfo *model.ServerInfo) error {
	server := fmt.Sprintf("%s:%d", serverInfo.Ip, serverInfo.Port)
	err := m.ExecWithDefaultTimeout(sql.DeleteServer, server)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete server")
		return errors.Wrap(err, "Delete server")
	}
	return nil
}
