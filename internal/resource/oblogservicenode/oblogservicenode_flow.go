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

// Package oblogservicenode implements the resource manager for OBLogServiceNode.
package oblogservicenode

import (
	nodestatus "github.com/oceanbase/ob-operator/internal/const/status/oblogservicenode"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func genCreateNodeFlow(_ *OBLogServiceNodeManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "create log service node",
			Tasks:        []tasktypes.TaskName{tCreatePod, tWaitPodReady},
			TargetStatus: nodestatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: nodestatus.Failed,
			},
		},
	}
}

func genRecoverNodeFlow(_ *OBLogServiceNodeManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "recover log service node",
			Tasks:        []tasktypes.TaskName{tDeletePod, tCreatePod, tWaitPodReady},
			TargetStatus: nodestatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: nodestatus.Running,
			},
		},
	}
}

func genDeleteNodeFlow(_ *OBLogServiceNodeManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete log service node",
			Tasks:        []tasktypes.TaskName{tDeletePod},
			TargetStatus: nodestatus.FinalizerFinished,
		},
	}
}
