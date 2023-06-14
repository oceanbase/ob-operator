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
)

// Connector represents a connection pool.
type Connector struct {
	ds     DataSource
	client *sqlx.DB
}

// DataSource represents a data source.
type DataSource interface {
	ID() string
	DriverName() string
	DataSourceName() string
	String() string
}

func NewConnector(dataSource DataSource) *Connector {
	return &Connector{
		ds: dataSource,
	}
}

func (c *Connector) Init() error {
	db, err := sqlx.Connect(c.ds.DriverName(), c.ds.DataSourceName())
	if err != nil {
		return err
	}
	c.client = db
	return nil
}

func (c *Connector) IsAlive() bool {
	if c.client == nil {
		return false
	}
	err := c.client.Ping()
	if err != nil {
		return false
	}
	return true
}

func (c *Connector) GetClient() *sqlx.DB {
	return c.client
}

func (c *Connector) Close() error {
	if c.client.DB != nil {
		return c.client.DB.Close()
	}
	return nil
}
