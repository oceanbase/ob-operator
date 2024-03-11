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

package param

type CreateNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type QueryEventParam struct {
	ObjectType string `json:"objectType" query:"objectType" binding:"omitempty"`
	Type       string `json:"type" query:"type" binding:"omitempty"`
	Name       string `json:"name" query:"name" binding:"omitempty"`
	Namespace  string `json:"namespace" query:"namespace" binding:"omitempty"`
}
