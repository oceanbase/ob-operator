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

package sql

const (
	GetZone    = "select name, value, lower(info) as info from __all_zone where zone = ? and name in ('idc', 'recovery_status', 'region', 'status', 'storage_type', 'zone_type')"
	AddZone    = "alter system add zone ?"
	DeleteZone = "alter system delete zone ?"
	StartZone  = "alter system start zone ?"
	StopZone   = "alter system stop zone ?"
)
