/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package constants

import "github.com/oceanbase/ob-operator/api/types"

const (
	ClusterOpTypeAddZones           types.ClusterOperationType = "AddZones"
	ClusterOpTypeDeleteZones        types.ClusterOperationType = "DeleteZones"
	ClusterOpTypeAdjustReplicas     types.ClusterOperationType = "AdjustReplicas"
	ClusterOpTypeUpgrade            types.ClusterOperationType = "Upgrade"
	ClusterOpTypeRestartOBServers   types.ClusterOperationType = "RestartOBServers"
	ClusterOpTypeExpandStorageSize  types.ClusterOperationType = "ExpandStorageSize"
	ClusterOpTypeModifyStorageClass types.ClusterOperationType = "ModifyStorageClass"
	ClusterOpTypeSetParameters      types.ClusterOperationType = "SetParameters"
)

const (
	ClusterOpStatusPending   types.ClusterOperationStatus = "Pending"
	ClusterOpStatusRunning   types.ClusterOperationStatus = "Running"
	ClusterOpStatusSucceeded types.ClusterOperationStatus = "Succeeded"
	ClusterOpStatusFailed    types.ClusterOperationStatus = "Failed"
)
