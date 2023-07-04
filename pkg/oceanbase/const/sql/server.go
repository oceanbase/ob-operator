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
	ListServer   = "select id, zone, svr_ip, svr_port, inner_port, with_rootserver, with_partition, lower(status) as status, start_service_time from __all_server"
	GetServer    = "select id, zone, svr_ip, svr_port, inner_port, with_rootserver, with_partition, lower(status) as status, start_service_time from __all_server where svr_ip = ? and svr_port = ?"
	AddServer    = "alter system add server ?"
	DeleteServer = "alter system delete server ?"
)
