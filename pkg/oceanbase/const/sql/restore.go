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
	restoreProgressFields = "tenant_id, job_id, restore_tenant_name, restore_tenant_id, restore_tenant_id, backup_tenant_name, backup_tenant_id, backup_cluster_name, backup_dest, restore_option, restore_scn, restore_scn_display, status, start_timestamp, backup_set_list, backup_piece_list, total_bytes, total_bytes_display, finish_bytes, finish_bytes_display, description"
	restoreHistoryFields  = restoreProgressFields + ", finish_timestamp, backup_cluster_version, ls_count, finish_ls_count, tablet_count, finish_tablet_count"
)

const (
	SetRestorePassword = "SET DECRYPTION IDENTIFIED BY ?"
	// tenant_name, uri, Time/SCN, restore_option
	StartRestoreWithLimit = "ALTER SYSTEM RESTORE %s FROM ? UNTIL %s=? WITH ?"
	// tenant_name, uri, restore_option
	StartRestoreUnlimited = "ALTER SYSTEM RESTORE %s FROM ? WITH ?"
	CancelRestore         = "ALTER SYSTEM CANCEL RESTORE ?"
	ReplayStandbyLog      = "ALTER SYSTEM RECOVER STANDBY TENANT ? UNTIL %s"
	ActivateStandby       = "ALTER SYSTEM ACTIVATE STANDBY TENANT ?"
	QueryRestoreProgress  = "SELECT " + restoreProgressFields + " FROM CDB_OB_RESTORE_PROGRESS"
	QueryRestoreHistory   = "SELECT " + restoreHistoryFields + " FROM CDB_OB_RESTORE_HISTORY"

	GetLatestRestoreProgress = QueryRestoreProgress + " WHERE restore_tenant_name = ? ORDER BY JOB_ID DESC LIMIT 1"
	GetLatestRestoreHistory  = QueryRestoreHistory + " WHERE restore_tenant_name = ? ORDER BY JOB_ID DESC LIMIT 1"
)
