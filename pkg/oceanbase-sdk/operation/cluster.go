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
	"strings"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) Bootstrap(ctx context.Context, bootstrapServerList []model.BootstrapServerInfo) error {
	serverInfoList := make([]string, 0, len(bootstrapServerList))
	for _, bootstrapServer := range bootstrapServerList {
		serverInfoList = append(serverInfoList, fmt.Sprintf(sql.BootstrapServer, bootstrapServer.Zone, bootstrapServer.Server.Ip, bootstrapServer.Server.Port))
	}
	bootstrapInfo := strings.Join(serverInfoList, ", ")
	bootstrapSql := fmt.Sprintf(sql.Bootstrap, bootstrapInfo)
	m.Logger.Info("Execute bootstrap sql", "sql", bootstrapSql, "datasource", m.Connector.DataSource().String())
	err := m.ExecWithTimeout(ctx, config.BootstrapTimeout, bootstrapSql)
	if err != nil {
		m.Logger.Error(err, "Got exception when bootstrap")
		return errors.Wrap(err, "Bootstrap")
	}
	return nil
}

// GetVersion gets oceanbase version from every observer with build number
func (m *OceanbaseOperationManager) GetVersion(ctx context.Context) (*model.OBVersion, error) {
	observers, err := m.ListServers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to List observers of cluster")
	}

	var version *model.OBVersion
	for _, observer := range observers {
		v, err := model.ParseOBVersion(observer.BuildVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse version %s of observer %s:%d", observer.BuildVersion, observer.Ip, observer.Port)
		}
		if version != nil && version.Compare(v) != 0 {
			return nil, errors.Errorf("Version %s of observer %s:%d is not consistent with other observer", observer.BuildVersion, observer.Ip, observer.Port)
		}
		version = v
	}
	return version, nil
}

func (m *OceanbaseOperationManager) ListClusterParameters(ctx context.Context) ([]*model.Parameter, error) {
	params := make([]*model.Parameter, 0)
	err := m.QueryList(ctx, &params, sql.ListParametersGroupByName)
	if err != nil {
		return nil, errors.Wrap(err, "List cluster parameters")
	}
	return params, nil
}
