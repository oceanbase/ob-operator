package operation

import (
	"context"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	obsql "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

// PreparedClassicBootstrap holds the connection configured for one bootstrap
// statement. The caller closes it whether or not it obtains the dispatch claim.
type PreparedClassicBootstrap struct {
	conn   *sqlx.Conn
	sql    string
	logger logr.Logger
}

func (m *OceanbaseOperationManager) PrepareClassicBootstrap(ctx context.Context, servers []model.BootstrapServerInfo) (*PreparedClassicBootstrap, error) {
	statement := fmt.Sprintf(obsql.Bootstrap, bootstrapServerListSQL(servers))
	logger := m.Logger.WithValues("datasource", m.Connector.DataSource().String())
	return prepareClassicBootstrap(ctx, m.Connector.GetClient().DB, logger, statement)
}

func prepareClassicBootstrap(ctx context.Context, db *sqlx.DB, logger logr.Logger, statement string) (*PreparedClassicBootstrap, error) {
	preflight, cancel := context.WithTimeout(ctx, managerConfig.DefaultSqlTimeout)
	defer cancel()
	conn, err := db.Connx(preflight)
	if err != nil {
		return nil, err
	}
	prepared := &PreparedClassicBootstrap{conn: conn, sql: statement, logger: logger}
	if _, err = conn.ExecContext(preflight, "set ob_query_timeout=?", int64(config.BootstrapTimeout/time.Microsecond)); err != nil {
		prepared.Close()
		return nil, fmt.Errorf("prepare bootstrap session timeout: %w", err)
	}
	return prepared, nil
}

func (s *PreparedClassicBootstrap) Exec(ctx context.Context) error {
	dispatch, cancel := context.WithTimeout(ctx, config.BootstrapTimeout)
	defer cancel()
	s.logger.Info("Execute bootstrap sql", "sql", s.sql)
	_, err := s.conn.ExecContext(dispatch, s.sql)
	return err
}

func (s *PreparedClassicBootstrap) Close() {
	// Discard the session so its unusually long timeout never reaches another user.
	_ = s.conn.Raw(func(any) error { return driver.ErrBadConn })
	_ = s.conn.Close()
}
