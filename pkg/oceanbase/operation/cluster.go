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
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"strings"
)

// TODO
func (m *OceanbaseOperationManager) Bootstrap(bootstrapServerList []model.BootstrapServerInfo) error {
	serverInfoList := make([]string, 0, len(bootstrapServerList))
	for _, bootstrapServer := range bootstrapServerList {
		serverInfoList = append(serverInfoList, fmt.Sprintf(sql.BootstrapServer, bootstrapServer.Region, bootstrapServer.Zone, bootstrapServer.Server.Ip, bootstrapServer.Server.Port))
	}
	bootstrapInfo := strings.Join(serverInfoList, ", ")
	bootstrapSql := fmt.Sprintf(sql.Bootstrap, bootstrapInfo)
	err := m.ExecWithTimeout(config.BootstrapTimeout, bootstrapSql)
	if err != nil {
		klog.Errorf("Got exception when bootstrap: %v", err)
		return errors.Wrap(err, "Bootstrap")
	}
	return nil
}
