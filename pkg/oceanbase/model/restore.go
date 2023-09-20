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

package model

// RestoreProgress is the progress of restore job, matches view CDB_OB_RESTORE_PROGRESS
type RestoreProgress struct {
	TenantId           int64   `json:"tenant_id" db:"tenant_id"`
	JobId              int64   `json:"job_id" db:"job_id"`
	RestoreTenantName  string  `json:"restore_tenant_name" db:"restore_tenant_name"`
	RestoreTenantId    int64   `json:"restore_tenant_id" db:"restore_tenant_id"`
	BackupTenantName   string  `json:"backup_tenant_name" db:"backup_tenant_name"`
	BackupTenantId     int64   `json:"backup_tenant_id" db:"backup_tenant_id"`
	BackupClusterName  string  `json:"backup_cluster_name" db:"backup_cluster_name"`
	BackupDest         string  `json:"backup_dest" db:"backup_dest"`
	RestoreOption      string  `json:"restore_option" db:"restore_option"`
	RestoreScn         int64   `json:"restore_scn" db:"restore_scn"`
	RestoreScnDisplay  string  `json:"restore_scn_display" db:"restore_scn_display"`
	Status             string  `json:"status" db:"status"`
	StartTimestamp     string  `json:"start_timestamp" db:"start_timestamp"`
	BackupSetList      string  `json:"backup_set_list" db:"backup_set_list"`
	BackupPieceList    string  `json:"backup_piece_list" db:"backup_piece_list"`
	TotalBytes         *int64  `json:"total_bytes,omitempty" db:"total_bytes"`
	TotalBytesDisplay  *string `json:"total_bytes_display,omitempty" db:"total_bytes_display"`
	FinishBytes        *int64  `json:"finish_bytes,omitempty" db:"finish_bytes"`
	FinishBytesDisplay *string `json:"finish_bytes_display,omitempty" db:"finish_bytes_display"`
	Description        *string `json:"description,omitempty" db:"description"`
}

// RestoreHistory is the history of restore job, matches view CDB_OB_RESTORE_HISTORY
type RestoreHistory struct {
	RestoreProgress `json:",inline" db:",inline"`

	FinishTimestamp      string `json:"finish_timestamp" db:"finish_timestamp"`
	BackupClusterVersion string `json:"backup_cluster_version" db:"backup_cluster_version"`
	LsCount              int64  `json:"ls_count" db:"ls_count"`
	FinishLsCount        int64  `json:"finish_ls_count" db:"finish_ls_count"`
	TabletCount          int64  `json:"tablet_count" db:"tablet_count"`
	FinishTabletCount    int64  `json:"finish_tablet_count" db:"finish_tablet_count"`
}
