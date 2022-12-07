/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observerconst

// StatefulApp
const (
	ImgOb         = "observer"
	ImgObagent    = "obagent"
	ImgPullPolicy = "IfNotPresent"

	DatafileStorageName = "data-file"
	DatafileStoragePath = "/home/admin/data_file"
	DatalogStorageName  = "data-log"
	DatalogStoragePath  = "/home/admin/data_log"
	LogStorageName      = "log"
	LogStoragePath      = "/home/admin/log"

	BackupName = "backup"
	BackupPath = "/ob-backup"

	CableReadinessPeriod = 2

	// monagent
	MonagentConfigPeriod = 2
	ConfFileStorageName  = "obagent-conf-file"
	ConfFileStoragePath  = "/home/admin/obagent/conf"
)
