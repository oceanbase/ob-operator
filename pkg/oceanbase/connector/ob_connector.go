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

package connector

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// oceanbase connector, support mysql mode only
type OceanbaseConnector struct {
	ConnectProperties *OceanbaseConnectProperties
	Client            *sqlx.DB
	PoolConfig        *ConnectionPoolConfig
}

func NewOceanbaseConnector(p *OceanbaseConnectProperties) *OceanbaseConnector {
	return &OceanbaseConnector{
		ConnectProperties: p,
		Client:            nil,
	}
}

func (oc *OceanbaseConnector) Init() error {
	var err error
	var db *sqlx.DB
	dsn := oc.ConnectProperties.GetDSN()
	db, err = sqlx.Connect(DRIVER_MYSQL, dsn)
	if err != nil {
		return errors.Wrapf(err, "Init db connection %s", dsn)
	}
	oc.Client = db
	return nil
}

func (oc *OceanbaseConnector) IsAlive() bool {
	err := oc.Client.Ping()
	if err != nil {
		return false
	}
	return true
}

func (oc *OceanbaseConnector) GetClient() *sqlx.DB {
	return oc.Client
}

func (oc *OceanbaseConnector) Close() error {
	if oc.Client.DB != nil {
		return oc.Client.DB.Close()
	}
	return nil
}
