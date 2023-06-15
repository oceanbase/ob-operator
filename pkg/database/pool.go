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
	"sync"

	"github.com/pkg/errors"
)

var p = &pool{}

type pool struct {
	// TODO: maintain cache size and data expiration, maybe change to a cache library.
	Cache sync.Map
}

func GetConnector(dataSource DataSource) (*Connector, error) {
	c, ok := p.Cache.Load(dataSource.ID())
	if ok && c.(*Connector).IsAlive() {
		return c.(*Connector), nil
	}
	connector := NewConnector(dataSource)
	err := connector.Init()
	if err != nil {
		err = errors.Wrap(err, "init connector failed with datasource: "+dataSource.String())
		return nil, err
	}
	p.Cache.Store(dataSource.ID(), connector)
	return connector, nil
}
