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
	// register mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Connector represents a connection pool.
type Connector struct {
	ds     DataSource
	client *Client
}

// DataSource represents a data source.
type DataSource interface {
	ID() string
	DriverName() string
	DataSourceName() string
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
	db, err := sqlx.Open(c.ds.DriverName(), c.ds.DataSourceName())
	if err != nil {
		return errors.Wrapf(err, "Open datasource %s", c.ds.String())
	}
	err = db.Ping()
	if err != nil {
		return errors.Wrapf(err, "Ping datasource %s", c.ds.String())
	}
	c.client = &Client{db}
	c.client.configure()
	return nil
}

func (c *Connector) IsAlive() bool {
	if c.client == nil {
		return false
	}
	return c.client.Ping() == nil
}

func (c *Connector) GetClient() *Client {
	return c.client
}

func (c *Connector) DataSource() string {
	return c.ds.String()
}

func (c *Connector) Close() error {
	if c.client.DB != nil {
		return c.client.DB.Close()
	}
	return nil
}
