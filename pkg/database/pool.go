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
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
)

var p = &pool{}

func onCacheEvicted(_ string, value *Connector) {
	_ = value.Close()
}

func init() {
	cacheSize := lruCacheSize
	var err error
	p.Cache, err = lru.NewWithEvict[string, *Connector](cacheSize, onCacheEvicted)
	if err != nil {
		panic(err)
	}
}

type pool struct {
	Cache *lru.Cache[string, *Connector]
}

func GetConnector(dataSource DataSource) (*Connector, error) {
	c, ok := p.Cache.Get(dataSource.ID())
	if ok && c.IsAlive() {
		return c, nil
	}
	connector := NewConnector(dataSource)
	err := connector.Init()
	if err != nil {
		return nil, errors.Wrapf(err, "Init connector failed with datasource: %s", dataSource.String())
	}
	p.Cache.Add(dataSource.ID(), connector)
	return connector, nil
}
