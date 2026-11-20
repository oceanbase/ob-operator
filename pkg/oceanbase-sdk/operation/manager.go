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
	stdsql "database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/database"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	obsql "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
)

const (
	redactedSQL             = "<redacted: SQL contains sensitive credentials>"
	sensitiveDetailsOmitted = "details omitted because the statement contains sensitive credentials"
)

var sharedStorageCredentialAssignment = regexp.MustCompile(`(?i)(^|[^[:alnum:]_])access_(id|key)[[:space:]]*=`)

func containsCredentials(value string) bool {
	return sharedStorageCredentialAssignment.MatchString(value)
}

func redactSQLForLog(statement string) (string, bool) {
	if statement == obsql.SetBackupPassword || statement == obsql.SetRestorePassword || containsCredentials(statement) {
		return redactedSQL, true
	}
	return statement, false
}

func redactParamsForLog(params []any) (string, bool) {
	return redactParamsForLogWithPolicy(params, false)
}

func redactParamsForLogWithPolicy(params []any, redactAll bool) (string, bool) {
	redacted := make([]string, len(params))
	containsSensitiveParam := redactAll
	for i, param := range params {
		if redactAll || containsCredentials(paramValueForCredentialDetection(param)) {
			redacted[i] = "***"
			containsSensitiveParam = true
		} else {
			redacted[i] = fmt.Sprint(param)
		}
	}
	return fmt.Sprintf("%v", redacted), containsSensitiveParam
}

func paramValueForCredentialDetection(param any) string {
	switch value := param.(type) {
	case []byte:
		return string(value)
	case stdsql.RawBytes:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}

func redactSensitiveExecutionError(err error) error {
	if errors.Is(err, context.Canceled) {
		return errors.Wrap(context.Canceled, "database execution canceled; "+sensitiveDetailsOmitted)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.Wrap(context.DeadlineExceeded, "database execution timed out; "+sensitiveDetailsOmitted)
	}

	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) {
		return &mysql.MySQLError{
			Number:   mysqlError.Number,
			SQLState: mysqlError.SQLState,
			Message:  sensitiveDetailsOmitted,
		}
	}
	return errors.New("database execution failed; " + sensitiveDetailsOmitted)
}

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

type operationSQLClient interface {
	ExecContext(context.Context, string, ...any) (stdsql.Result, error)
	GetContext(context.Context, any, string, ...any) error
	SelectContext(context.Context, any, string, ...any) error
}

type OceanbaseOperationManager struct {
	session   *sqlx.Conn
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

func (m *OceanbaseOperationManager) sqlClient() operationSQLClient {
	if m.session != nil {
		return m.session
	}
	return m.Connector.GetClient()
}

// WithBackupSession keeps session-local encryption or decryption settings on the
// same physical connection as the backup or restore that consumes them.
func (m *OceanbaseOperationManager) WithBackupSession(ctx context.Context, run func(*OceanbaseOperationManager) error) error {
	return m.WithDedicatedSession(ctx, run)
}

// WithDedicatedSession pins identity checks and subsequent statements to one
// physical SQL connection. The connection is discarded before returning.
func (m *OceanbaseOperationManager) WithDedicatedSession(ctx context.Context, run func(*OceanbaseOperationManager) error) error {
	acquire, cancel := context.WithTimeout(ctx, managerConfig.DefaultSqlTimeout)
	conn, err := m.Connector.GetClient().Connx(acquire)
	cancel()
	if err != nil {
		return err
	}
	defer func() {
		// Discard session-local credentials and timeout settings.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		_ = conn.Close()
	}()
	session := *m
	session.session = conn
	return run(&session)
}

func (m *OceanbaseOperationManager) setQueryTimeoutVariable(ctx context.Context, timeout time.Duration) error {
	m.Logger.V(1).Info(fmt.Sprintf("set timeout to %d seconds", int64(timeout/time.Second)))
	_, err := m.sqlClient().ExecContext(ctx, "set ob_query_timeout=?", int64(timeout/time.Microsecond))
	return err
}

func (m *OceanbaseOperationManager) ExecWithTimeout(ctx context.Context, timeout time.Duration, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := m.setQueryTimeoutVariable(c, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to set timeout variable")
	}
	return m.executeStatement(c, sql, params...)
}

func (m *OceanbaseOperationManager) executeStatement(ctx context.Context, sql string, params ...any) error {
	redactedSQL, sqlContainsCredentials := redactSQLForLog(sql)
	redactedParams, paramsContainCredentials := redactParamsForLogWithPolicy(params, sqlContainsCredentials)
	containsSensitiveCredentials := sqlContainsCredentials || paramsContainCredentials
	m.Logger.V(1).Info(fmt.Sprintf("Execute sql %s with param %s", redactedSQL, redactedParams))
	_, err := m.sqlClient().ExecContext(ctx, sql, params...)
	if err != nil {
		if containsSensitiveCredentials {
			// Drivers may echo an interpolated SQL statement in their error. Do not
			// retain the original error in the returned chain when credentials were
			// present, otherwise logs and Kubernetes events can expose the secret.
			err = redactSensitiveExecutionError(err)
		}
		err = errors.Wrapf(err, "Execute sql failed, sql %s, param %s", redactedSQL, redactedParams)
		m.Logger.Error(err, "Execute sql failed")
	}
	return err
}

func (m *OceanbaseOperationManager) ExecWithDefaultTimeout(ctx context.Context, sql string, params ...any) error {
	m.Logger.V(1).Info("Check default sql timeout", "timeout", managerConfig.DefaultSqlTimeout)
	return m.ExecWithTimeout(ctx, managerConfig.DefaultSqlTimeout, sql, params...)
}

func (m *OceanbaseOperationManager) QueryRowWithTimeout(ctx context.Context, timeout time.Duration, ret any, sql string, params ...any) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := m.setQueryTimeoutVariable(c, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to set timeout variable")
	}
	err = m.sqlClient().GetContext(c, ret, sql, params...)
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
	err := m.setQueryTimeoutVariable(c, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to set timeout variable")
	}
	err = m.sqlClient().SelectContext(c, ret, sql, params...)
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
	err := m.sqlClient().GetContext(c, count, sql, params...)
	if err != nil {
		err = errors.Wrapf(err, "Query count failed, sql %s, param %v", sql, params)
		m.Logger.Error(err, "Query count failed")
	}
	return err
}

func (m *OceanbaseOperationManager) Close() error {
	m.Logger.V(1).Info("Closing OceanbaseOperationManager")
	return m.Connector.Close()
}
