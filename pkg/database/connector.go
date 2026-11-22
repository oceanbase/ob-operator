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

package database

import (
	"context"
	"crypto/sha256"
	"strings"
	"time"

	// register mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Connector represents a connection pool.
type Connector struct {
	ds            DataSource
	client        *Client
	connectionKey [sha256.Size]byte
}

// DataSource represents a data source.
type DataSource interface {
	ID() string
	DriverName() string
	DataSourceName() string
	GetAddress() string
	GetPort() int64
	GetUser() string
	GetPassword() string
	GetDatabase() string
	String() string
}

// Client represents a wrapper around sqlx.DB.
type Client struct {
	*sqlx.DB
}

func (c *Client) configure() {
	c.SetMaxOpenConns(DefaultConnMaxOpenCount)
	c.SetMaxIdleConns(DefaultConnMaxIdleCount)
	c.SetConnMaxLifetime(DefaultConnMaxLifetime)
	c.SetConnMaxIdleTime(DefaultConnMaxIdleTime)
}

func NewConnector(dataSource DataSource) *Connector {
	return &Connector{
		ds: dataSource,
	}
}

func (c *Connector) Init() error {
	return c.open(c.ds.DriverName(), c.ds.DataSourceName(), c.ds.GetPassword())
}

func (c *Connector) open(driverName, dsn, password string) error {
	db, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return connectionError("open", c.ds.String(), password, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultPingTimeoutSeconds)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		_ = db.Close()
		return connectionError("ping", c.ds.String(), password, err)
	}
	c.client = &Client{db}
	c.client.configure()
	return nil
}

func (c *Connector) IsAlive() bool {
	if c.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultPingTimeoutSeconds)
	defer cancel()
	return c.client.PingContext(ctx) == nil
}

func (c *Connector) GetClient() *Client {
	return c.client
}

func (c *Connector) DataSource() DataSource {
	return c.ds
}

func (c *Connector) Close() error {
	if c.client != nil && c.client.DB != nil {
		return c.client.DB.Close()
	}
	return nil
}

// Drivers may echo rejected credentials in their diagnostics.
type redactedDriverError struct{ cause error }

func (e redactedDriverError) Error() string { return "database driver response contains credentials" }
func (e redactedDriverError) Unwrap() error { return e.cause }

func connectionError(operation, description, password string, err error) error {
	if password != "" {
		description = strings.ReplaceAll(description, password, "[redacted]")
		if strings.Contains(err.Error(), password) {
			err = redactedDriverError{cause: err}
		}
	}
	return errors.Wrapf(err, "%s datasource %s", operation, description)
}
