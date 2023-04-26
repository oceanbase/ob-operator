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
	SetParameter            = "ALTER SYSTEM SET %s = ?"
	SetParameterWithScope   = "ALTER SYSTEM SET %s = ? %s = ?"
	QueryParameter          = "SELECT ZONE, SVR_IP, SVR_PORT, NAME, VALUE, SCOPE, EDIT_LEVEL FROM __ALL_VIRTUAL_SYS_PARAMETER_STAT WHERE NAME = ?"
	QueryParameterWithScope = "SELECT ZONE, SVR_IP, SVR_PORT, NAME, VALUE, SCOPE, EDIT_LEVEL FROM __ALL_VIRTUAL_SYS_PARAMETER_STAT WHERE NAME = ? AND %s = ?"
)
