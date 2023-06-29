/*
Copyright (c) 2021 OceanBase
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

import "strings"

func ReplaceAll(template string, replacers ...*strings.Replacer) string {
	s := template
	for _, replacer := range replacers {
		s = replacer.Replace(s)
	}
	return s
}

func SetParameterSQLReplacer(name, value string) *strings.Replacer {
	return strings.NewReplacer("${NAME}", name, "${VALUE}", value)
}

func CreateResourceUnitSQLReplacer(unit_name, max_cpu, max_memory, max_iops, max_disk_size, max_session_num, min_cpu, min_memory, min_iops string) *strings.Replacer {
	return strings.NewReplacer("${unit_name}", unit_name, "${max_cpu}", max_cpu, "${max_memory}", max_memory, "${max_iops}", max_iops, "${max_disk_size}", max_disk_size, "${max_session_num}", max_session_num, "${min_cpu}", min_cpu, "${min_memory}", min_memory, "${min_iops}", min_iops)
}

func ActivateTenantSqlReplacer(tenant string) *strings.Replacer {
	return strings.NewReplacer("${tenant}", tenant)
}

func CreateResourcePoolSQLReplacer(pool_name, unit_name, unit_num, zone_list string) *strings.Replacer {
	return strings.NewReplacer("${pool_name}", pool_name, "${unit_name}", unit_name, "${unit_num}", unit_num, "${zone_list}", zone_list)
}

func DoRestoreSQLReplacer(dest_tenant, source_tenant, dest_path, time, backup_cluster_name, backup_cluster_id, pool_list, restore_option string) *strings.Replacer {
	return strings.NewReplacer("${dest_tenant}", dest_tenant, "${source_tenant}", source_tenant, "${dest_path}", dest_path, "${time}", time, "${backup_cluster_name}", backup_cluster_name, "${backup_cluster_id}", backup_cluster_id, "${pool_list}", pool_list, "${restore_option}", restore_option)
}

func DoRestoreSQLReplacer4(dest_tenant, dest_path, save_point, backup_cluster_name, backup_cluster_id, pool_list, restore_option string) *strings.Replacer {
	return strings.NewReplacer("${dest_tenant}", dest_tenant, "${dest_path}", dest_path, "${save_point}", save_point, "${backup_cluster_name}", backup_cluster_name, "${backup_cluster_id}", backup_cluster_id, "${pool_list}", pool_list, "${restore_option}", restore_option)
}
