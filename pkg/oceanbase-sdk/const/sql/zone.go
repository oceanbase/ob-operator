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
	GetZone               = "select zone as name, lower(status) as status, idc, region, type from DBA_OB_ZONES where zone = ?"
	ListZones             = "select zone as name, lower(status) as status, idc, region, type from DBA_OB_ZONES"
	ListZoneStorageState  = "select upper(state) as state from DBA_OB_ZONE_STORAGE where zone = ?"
	ListZoneStorageDest   = "select path, endpoint, upper(used_for) as used_for, upper(state) as state from DBA_OB_ZONE_STORAGE where zone = ?"
	CountZoneServers      = "select count(*) from oceanbase.DBA_OB_SERVERS where zone = ?"
	DropSharedStorageDest = "alter system drop shared_storage_dest path = ? for all zone = ?"
	AddZone               = "alter system add zone ?"
	AddSharedStorageDest  = "alter system add shared_storage_dest path = ? access_info = ? attribute = ? for all zone = ?"
	DeleteZone            = "alter system delete zone ?"
	StartZone             = "alter system start zone ?"
	StopZone              = "alter system stop zone ?"
)
