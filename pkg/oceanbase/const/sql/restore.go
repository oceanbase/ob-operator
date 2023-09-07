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
	SetRestorePassword = "SET DECRYPTION IDENTIFIED BY ?"
	// tenant_name, uri, Time/SCN, restore_option
	StartRestoreWithLimit = "ALTER SYSTEM RESTORE ? FROM ? UNTIL %s=? WITH ?"
	// tenant_name, uri, restore_option
	StartRestoreUnlimited = "ALTER SYSTEM RESTORE ? FROM ? WITH ?"
	CancelRestore         = "ALTER SYSTEM CANCEL RESTORE ?"
	ReplayStandbyLog      = "ALTER SYSTEM RECOVER STANDBY TENANT ? UNTIL %s"
	ActivateStandby       = "ALTER SYSTEM ACTIVATE STANDBY TENANT ?"
	QueryRestoreProgress  = "SELECT * FROM CDB_OB_RESTORE_PROGRESS"
	QueryRestoreHistory   = "SELECT * FROM CDB_OB_RESTORE_HISTORY"
)
