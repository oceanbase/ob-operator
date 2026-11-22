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
	"crypto/sha256"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
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
	// Serialize replacements for an identity without blocking cache hits on its network I/O.
	replacements [256]sync.Mutex
}

func GetConnector(dataSource DataSource) (*Connector, error) {
	id := dataSource.ID()
	driverName, dsn, password := dataSource.DriverName(), dataSource.DataSourceName(), dataSource.GetPassword()
	// Compare the configuration actually used by the driver, even if the caller
	// subsequently changes the original DataSource in place.
	key := sha256.Sum256([]byte(driverName + "\x00" + dsn))
	cached, _ := p.Cache.Get(id)
	if cached != nil && cached.connectionKey == key && cached.IsAlive() {
		return cached, nil
	}

	lock := &p.replacements[sha256.Sum256([]byte(id))[0]]
	lock.Lock()
	defer lock.Unlock()
	if current, _ := p.Cache.Get(id); current != nil && current != cached && current.connectionKey == key && current.IsAlive() {
		return current, nil
	}
	connector := NewConnector(dataSource)
	if err := connector.open(driverName, dsn, password); err != nil {
		return nil, err
	}
	connector.connectionKey = key
	// Add does not evict the old value when replacing the same key. Removing
	// it after successful authentication also closes its database pool.
	p.Cache.Remove(id)
	p.Cache.Add(id, connector)
	return connector, nil
}
