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

// Package oblogservicecluster implements the resource manager for OBLogServiceCluster.
package oblogservicecluster

import (
	lsstatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicecluster"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genBootstrapLogServiceFlow(_ *OBLogServiceClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "bootstrap log service cluster",
			Tasks:        []tasktypes.TaskName{tCreateZones, tWaitZonesBootstrapReady, tBootstrapLogService, tMarkNodesRegistered},
			TargetStatus: lsstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func genModifyZoneReplicaFlow(_ *OBLogServiceClusterManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "modify log service zone replica",
			Tasks:        []tasktypes.TaskName{tModifyZoneReplica, tWaitZonesRunning},
			TargetStatus: lsstatus.Running,
		},
	}
}
