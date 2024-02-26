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

type TenantRole string
type BackupDestType string

type BackupDestination struct {
	Type            BackupDestType `json:"type,omitempty"`
	Path            string         `json:"path,omitempty"`
	OSSAccessSecret string         `json:"ossAccessSecret,omitempty"`
}

type NamespacedName struct {
	Namespace string `json:"namespace" uri:"namespace" binding:"required"`
	Name      string `json:"name" uri:"name" binding:"required"`
}
