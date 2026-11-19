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
	GetGlobalVariable = "select name, coalesce(value, \"\") as value from DBA_OB_SYS_VARIABLES where scope like '%global%' and name = ?"
	// Older releases expose the raw stored value (e.g. "1" for autocommit)
	// separately from the list of variables that are valid at global scope.
	GetLegacyGlobalVariable = `SELECT sv.name AS name, COALESCE(sv.value, '') AS value
FROM oceanbase.__all_sys_variable AS sv
JOIN information_schema.GLOBAL_VARIABLES AS gv ON gv.VARIABLE_NAME = sv.name
WHERE sv.name = ?`
	SetGlobalVariable = "set global %s = ?"
)
