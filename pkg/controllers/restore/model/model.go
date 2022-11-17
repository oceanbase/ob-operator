/*
PURPOSECopyright (c) 2021 OceanBase
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

type AllRestoreSet struct {
	JobId                  int64
	BackupClusterId        int64
	BackupClusterName      string
	TenantName             string
	BackupTenantName       string
	Status                 string
	RestoreFinishTimestamp string
}

type RestoreConcurrency struct {
	ZoneName string
	SvrIP    string
	Value    int64
}
