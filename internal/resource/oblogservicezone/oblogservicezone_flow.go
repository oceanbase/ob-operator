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

// Package oblogservicezone implements the resource manager for OBLogServiceZone.
package oblogservicezone

import (
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicezone"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genPrepareZoneForBootstrapFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "prepare log service zone for bootstrap",
			Tasks:        []tasktypes.TaskName{tCreateNodes, tWaitNodesReady},
			TargetStatus: zonestatus.BootstrapReady,
		},
	}
}

func genMaintainZoneAfterBootstrapFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain log service zone after bootstrap",
			Tasks:        []tasktypes.TaskName{tWaitNodesRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genCreateNodesFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "create log service nodes",
			Tasks:        []tasktypes.TaskName{tCreateNodes, tWaitNodesRunning, tRegisterNodeToCluster},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genAddNodeFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "add log service node",
			Tasks:        []tasktypes.TaskName{tCreateNodes, tWaitNodesRunning, tRegisterNodeToCluster},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genDeleteNodeFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete log service node",
			Tasks:        []tasktypes.TaskName{tUnregisterNodeFromCluster, tDeleteExcessNodes},
			TargetStatus: zonestatus.Running,
		},
	}
}

func genDeleteZoneFlow(_ *OBLogServiceZoneManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete log service zone",
			Tasks:        []tasktypes.TaskName{tUnregisterAllNodesFromCluster, tDeleteAllNodes},
			TargetStatus: zonestatus.FinalizerFinished,
		},
	}
}
