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
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"time"
)

type OceanbaseOperationManager struct {
	Connector *connector.OceanbaseConnector
}

func NewOceanbaseOperationManager(connector *connector.OceanbaseConnector) *OceanbaseOperationManager {
	return &OceanbaseOperationManager{
		Connector: connector,
	}
}

func GetOceanbaseOperationManager(p *connector.OceanbaseConnectProperties) (*OceanbaseOperationManager, error) {
	connector, err := connector.GetOceanbaseConnectorManager().GetOceanbaseConnector(p)
	if err != nil {
		return nil, errors.Wrap(err, "Get OceanBase connector")
	}
	return NewOceanbaseOperationManager(connector), nil
}

func (m *OceanbaseOperationManager) ExecWithTimeout(timeout time.Duration, sql string, params ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := m.Connector.Client.ExecContext(ctx, sql, params...)
	if err != nil {
		klog.Errorf("Execute sql %s with param %v got error: %v", sql, params, err)
		return errors.Wrap(err, "Execute sql")
	}
	return nil
}

func (m *OceanbaseOperationManager) ExecWithDefaultTimeout(sql string, params ...interface{}) error {
	return m.ExecWithTimeout(config.DefaultSqlTimeout, sql, params...)
}

func (m *OceanbaseOperationManager) Query(sql string, params ...interface{}) error {
	return m.ExecWithTimeout(config.DefaultSqlTimeout, sql, params...)
}
