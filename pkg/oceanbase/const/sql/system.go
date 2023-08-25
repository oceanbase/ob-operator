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
	QueryTenantWithName    = "SELECT tenant_id, tenant_name, tenant_type, create_time, modify_time, primary_zone, locality, compatibility_mode, status, in_recyclebin, locked, tenant_role, sync_scn, replayable_scn, readable_scn, recovery_until_scn, log_mode, arbitration_service_status FROM DBA_OB_TENANTS where tenant_name = ?"
	QueryUnitsWithTenantId = "SELECT unit_id, tenant_id, status, resource_pool_id, unit_group_id, create_time, modify_time, zone, svr_ip, svr_port, unit_config_id, max_cpu, min_cpu, memory_size, log_disk_size, max_iops, min_iops, iops_weight FROM DBA_OB_UNITS where tenant_id = ?"
)
