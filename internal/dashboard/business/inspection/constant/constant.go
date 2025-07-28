/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package constant

const (
	InspectionPackageBasic       = "k8s_basic"
	InspectionPackagePerformance = "k8s_performance"
	ConfigVolumeName             = "config"
	ConfigMountPath              = "/etc/config"
	ClusterRoleName              = "oceanbase-dashboard-cluster-role"
	ServiceAccountNameFmt        = "ob-ins-%s"
	ClusterRoleBindingNameFmt    = "ob-ins-%s-%s"
)
