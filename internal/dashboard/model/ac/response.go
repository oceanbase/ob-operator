/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package ac

import "time"

type Account struct {
	Username    string     `json:"username" binding:"required"`
	Nickname    string     `json:"nickname"`
	Description string     `json:"description"`
	Roles       []Role     `json:"roles" binding:"required"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	NeedReset   bool       `json:"needReset"`
}

type Role struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Policies    []Policy `json:"policies" binding:"required"`
}
