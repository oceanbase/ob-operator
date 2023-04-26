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
	"k8s.io/klog/v2"
)

func (m *OceanbaseOperationManager) AddServer(serverInfo *model.ServerInfo) error {
	server := fmt.Sprintf("%s:%d", serverInfo.Ip, serverInfo.Port)
	err := m.ExecWithDefaultTimeout(sql.AddServer, server)
	if err != nil {
		klog.Errorf("Got exception when add server: %v", err)
		return errors.Wrap(err, "Add server")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteServer(serverInfo *model.ServerInfo) error {
	server := fmt.Sprintf("%s:%d", serverInfo.Ip, serverInfo.Port)
	err := m.ExecWithDefaultTimeout(sql.DeleteServer, server)
	if err != nil {
		klog.Errorf("Got exception when delete server: %v", err)
		return errors.Wrap(err, "Delete server")
	}
	return nil
}
