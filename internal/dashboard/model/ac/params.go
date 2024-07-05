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

type CreateAccountParam struct {
	Username    string `json:"username" binding:"required"`
	Nickname    string `json:"nickname" binding:"required"`
	Description string `json:"description"`
	Password    string `json:"password" binding:"required"`
	RoleName    string `json:"roleName" binding:"required"`
}

type PatchAccountParam struct {
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
	RoleName    string `json:"roleName"`
	Password    string `json:"password"`
}

type CreateRoleParam struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []Policy `json:"permissions" binding:"required"`
}

type PatchRoleParam struct {
	Description string   `json:"description"`
	Permissions []Policy `json:"permissions"`
}
