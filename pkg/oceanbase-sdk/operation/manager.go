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
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
)

type ManagerConfig struct {
	DefaultSqlTimeout    time.Duration
	TenantSqlTimeout     time.Duration
	TenantRestoreTimeout time.Duration
	PollingJobSleepTime  time.Duration
}

var (
	managerConfig = &ManagerConfig{
		DefaultSqlTimeout:    config.DefaultSqlTimeout,
		TenantSqlTimeout:     config.TenantSqlTimeout,
		TenantRestoreTimeout: config.TenantRestoreTimeOut,
		PollingJobSleepTime:  config.PollingJobSleepTime,
	}
	once sync.Once
)

func SetManagerConfig(cfg *ManagerConfig) {
	once.Do(func() {
		managerConfig = cfg
	})
}

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

func (m *OceanbaseOperationManager) ExecWithTimeout(ctx context.Context, timeout time.Duration, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	m.Logger.Info(fmt.Sprintf("set timeout to %d seconds", int64(timeout/time.Second)))
	_, err := m.Connector.GetClient().ExecContext(c, "set ob_query_timeout=?", int64(timeout/time.Microsecond))
	if err != nil {
		return errors.Wrap(err, "Failed to set timeout variable")

	}
	m.Logger.Info(fmt.Sprintf("Execute sql %s with param %v", sql, params))
	_, err = m.Connector.GetClient().ExecContext(c, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Execute sql failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Execute sql failed")
	}
	return err
}

func (m *OceanbaseOperationManager) ExecWithDefaultTimeout(ctx context.Context, sql string, params ...any) error {
	m.Logger.Info("Check default sql timeout", "timeout", managerConfig.DefaultSqlTimeout)
	return m.ExecWithTimeout(ctx, managerConfig.DefaultSqlTimeout, sql, params...)
}

func (m *OceanbaseOperationManager) QueryRowWithTimeout(ctx context.Context, timeout time.Duration, ret any, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := m.Connector.GetClient().GetContext(c, ret, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query row, sql %s, param %v", sql, params)
	}
	return err
}

func (m *OceanbaseOperationManager) QueryRow(ctx context.Context, ret any, sql string, params ...any) error {
	return m.QueryRowWithTimeout(ctx, managerConfig.DefaultSqlTimeout, ret, sql, params...)
}

func (m *OceanbaseOperationManager) QueryListWithTimeout(ctx context.Context, timeout time.Duration, ret any, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := m.Connector.GetClient().SelectContext(c, ret, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query list failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Query list failed")
	}
	return err
}

func (m *OceanbaseOperationManager) QueryList(ctx context.Context, ret any, sql string, paramstx ...any) error {
	return m.QueryListWithTimeout(ctx, managerConfig.DefaultSqlTimeout, ret, sql, paramstx...)
}

func (m *OceanbaseOperationManager) QueryCount(ctx context.Context, count *int, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, managerConfig.DefaultSqlTimeout)
	defer cancel()
	err := m.Connector.GetClient().GetContext(c, count, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query count failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Query count failed")
	}
	return err
}

func (m *OceanbaseOperationManager) Close() error {
	return m.Connector.Close()
}
