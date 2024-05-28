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

const (
	DefaultConnMaxOpenCount = 20
	DefaultConnMaxIdleCount = 1
	DefaultConnMaxLifetime  = 0
	DefaultConnMaxIdleTime  = 0

	DefaultPingTimeoutSeconds = 3
)

const (
	DefaultLRUCacheSize = 1000
)

var (
	lruCacheSize = DefaultLRUCacheSize
)

func SetLRUCacheSize(size int) {
	lruCacheSize = size
}
