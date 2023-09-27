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
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
)

type OceanbaseOperationManager struct {
	Connector *database.Connector
	Logger    *logr.Logger
}

func NewOceanbaseOperationManager(connector *database.Connector) *OceanbaseOperationManager {
	return &OceanbaseOperationManager{
		Connector: connector,
	}
}

func GetOceanbaseOperationManager(p *connector.OceanBaseDataSource) (*OceanbaseOperationManager, error) {
	connector, err := database.GetConnector(p)
	if err != nil {
		return nil, err
	}
	return NewOceanbaseOperationManager(connector), nil
}

func (m *OceanbaseOperationManager) ExecWithTimeout(timeout time.Duration, sql string, params ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	m.Logger.Info(fmt.Sprintf("Execute sql %s with param %v", sql, params))
	_, err := m.Connector.GetClient().ExecContext(ctx, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Execute sql failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Execute sql failed")
	}
	return err
}

func (m *OceanbaseOperationManager) ExecWithDefaultTimeout(sql string, params ...any) error {
	return m.ExecWithTimeout(config.DefaultSqlTimeout, sql, params...)
}

func (m *OceanbaseOperationManager) QueryRowWithTimeout(timeout time.Duration, ret any, sql string, params ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := m.Connector.GetClient().GetContext(ctx, ret, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query row, sql %s, param %v", sql, params)
	}
	return err
}

func (m *OceanbaseOperationManager) QueryRow(ret any, sql string, params ...any) error {
	return m.QueryRowWithTimeout(config.DefaultSqlTimeout, ret, sql, params...)
}

func (m *OceanbaseOperationManager) QueryListWithTimeout(timeout time.Duration, ret any, sql string, params ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := m.Connector.GetClient().SelectContext(ctx, ret, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query list failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Query list failed")
	}
	return err
}

func (m *OceanbaseOperationManager) QueryList(ret any, sql string, params ...any) error {
	return m.QueryListWithTimeout(config.DefaultSqlTimeout, ret, sql, params...)
}

func (m *OceanbaseOperationManager) QueryCount(count *int, sql string, params ...any) error {
	err := m.Connector.GetClient().Get(count, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query count failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Query count failed")
	}
	return err
}

func (m *OceanbaseOperationManager) Close() error {
	return m.Connector.Close()
}
