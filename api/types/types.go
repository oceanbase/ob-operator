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

package types

type BackupJobType string
type BackupJobStatus string
type BackupPolicyStatusType string
type BackupDestType string
type LogArchiveDestState string
type ArchiveBinding string

type BackupDestination struct {
	Path            string         `json:"path"`
	Type            BackupDestType `json:"type,omitempty"`
	OSSAccessSecret string         `json:"ossAccessSecret,omitempty"`
}

type RestoreJobStatus string

type TenantRole string
type TenantOperationStatus string
type TenantOperationType string

type ClusterOperationType string
type ClusterOperationStatus string
