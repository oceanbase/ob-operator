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

package oceanbase

const (
	LabelRefOBTenant     = "ref-obtenant"
	LabelRefOBCluster    = "ref-obcluster"
	LabelRefOBZone       = "ref-obzone"
	LabelRefOBServer     = "ref-observer"
	LabelRefUID          = "ref-uid"
	LabelJobName         = "job-name"
	LabelRefBackupPolicy = "ref-backuppolicy"
	LabelRefOBClusterOp  = "ref-obclusterop"
	LabelRefOBTenantOp   = "ref-obtenantop"
)

const (
	LabelTenantName      = "oceanbase.oceanbase.com/tenant-name"
	LabelSecondaryTenant = "oceanbase.oceanbase.com/secondary-tenant"
	LabelBackupType      = "oceanbase.oceanbase.com/backup-type"
	LabelOBServerUID     = "oceanbase.oceanbase.com/observer-uid"

	LabelK8sCluster = "oceanbase.oceanbase.com/k8s-cluster"
)
