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

package variables

var ReadonlyVariables = []string{
	"plugin_dir",
	"version_comment",
	"ob_tcp_invited_nodes",
	"ob_compatibility_mode",
	"version",
	"system_time_zone",
	"license",
	"character_set_system",
	"lower_case_table_names",
	"datadir",
	"nls_characterset",
	"query_cache_type",
}

var UnsupportedVariables = []string{
	"nls_language",
	"nls_territory",
	"nls_sort",
	"nls_comp",
	"nls_characterset",
	"nls_nchar_characterset",
	"nls_date_language",
	"nls_nchar_conv_excp",
	"nls_calendar",
	"nls_numeric_characters",
}
