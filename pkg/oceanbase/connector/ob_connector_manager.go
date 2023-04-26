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
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"sync"
)

var oceanbaseConnectorManager *OceanbaseConnectorManager
var oceanbaseConnectorManagerCreateOnce sync.Once

type OceanbaseConnectorManager struct {
	// TODO: maintain cache size and data experiation, maybe change to a cache library
	Cache sync.Map
}

func GetOceanbaseConnectorManager() *OceanbaseConnectorManager {
	oceanbaseConnectorManagerCreateOnce.Do(func() {
		oceanbaseConnectorManager = &OceanbaseConnectorManager{}
	})
	return oceanbaseConnectorManager
}

func (ocm *OceanbaseConnectorManager) GetOceanbaseConnector(p *OceanbaseConnectProperties) (*OceanbaseConnector, error) {
	key := p.HashValue()
	connectorStored, loaded := ocm.Cache.Load(key)
	if loaded && connectorStored.(*OceanbaseConnector).IsAlive() {
		return connectorStored.(*OceanbaseConnector), nil
	} else {
		klog.Warningf("no connector or connector is not alive in cache with connect property: %s:%d %s", p.Address, p.Port, p.User)
		connector := NewOceanbaseConnector(p)
		err := connector.Init()
		if err != nil {
			klog.Errorf("init connector failed with connect property: %s:%d %s, %v", p.Address, p.Port, p.User, err)
			return nil, errors.Wrap(err, "create oceanbase connector")
		}
		ocm.Cache.Store(key, connector)
		return connector, nil
	}
}
