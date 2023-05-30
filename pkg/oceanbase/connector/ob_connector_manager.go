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
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

var oceanbaseConnectorManager *OceanbaseConnectorManager
var oceanbaseConnectorManagerCreateOnce sync.Once

type OceanbaseConnectorManager struct {
	// TODO: maintain cache size and data experiation, maybe change to a cache library
	Cache  sync.Map
	Logger *logr.Logger
}

func GetOceanbaseConnectorManager() *OceanbaseConnectorManager {
	oceanbaseConnectorManagerCreateOnce.Do(func() {
		logger := logr.FromContextOrDiscard(context.TODO())
		oceanbaseConnectorManager = &OceanbaseConnectorManager{
			Logger: &logger,
		}
	})
	return oceanbaseConnectorManager
}

func (m *OceanbaseConnectorManager) GetOceanbaseConnector(p *OceanbaseConnectProperties) (*OceanbaseConnector, error) {
	key := p.HashValue()
	connectorStored, loaded := m.Cache.Load(key)
	if loaded && connectorStored.(*OceanbaseConnector).IsAlive() {
		return connectorStored.(*OceanbaseConnector), nil
	}
	m.Logger.Info("no connector or connector is not alive in cache with connect property", "address", p.Address, "port", p.Port, "user", p.User)
	connector := NewOceanbaseConnector(p)
	err := connector.Init()
	if err != nil {
		m.Logger.Error(err, "init connector failed with connect property", "address", p.Address, "port", p.Port, "user", p.User)
		return nil, errors.Wrap(err, "create oceanbase connector")
	}
	m.Cache.Store(key, connector)
	return connector, nil
}
