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

package observer

import (
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func PrepareOBServerForBootstrap() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPrepareOBServerForBootstrap,
			Tasks:        []tasktypes.TaskName{tCreateOBPVC, tCreateOBPod, tWaitOBServerReady},
			TargetStatus: serverstatus.BootstrapReady,
		},
	}
}

func MaintainOBServerAfterBootstrap() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fPrepareOBServerForBootstrap,
			Tasks:        []tasktypes.TaskName{tWaitOBClusterBootstrapped, tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func CreateOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fCreateOBServer,
			Tasks:        []tasktypes.TaskName{tCreateOBPVC, tCreateOBPod, tWaitOBServerReady, tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: "Failed",
			},
		},
	}
}

func DeleteOBServerFinalizer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fDeleteOBServerFinalizer,
			Tasks:        []tasktypes.TaskName{tDeleteOBServerInCluster, tWaitOBServerDeletedInCluster},
			TargetStatus: serverstatus.FinalizerFinished,
		},
	}
}

func UpgradeOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fUpgradeOBServer,
			Tasks:        []tasktypes.TaskName{tUpgradeOBServerImage, tWaitOBServerPodReady, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func RecoverOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fRecoverOBServer,
			Tasks:        []tasktypes.TaskName{tCreateOBPod, tWaitOBServerReady, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func AddServerInOB() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAddServerInOB,
			Tasks:        []tasktypes.TaskName{tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func AnnotateOBServerPod() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fAnnotateOBServerPod,
			Tasks:        []tasktypes.TaskName{tAnnotateOBServerPod},
			TargetStatus: serverstatus.Running,
		},
	}
}

func ScaleUpOBServer() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fScaleUpOBServer,
			Tasks:        []tasktypes.TaskName{tDeletePod, tWaitForPodDeleted, tCreateOBPod, tWaitOBServerReady, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func ResizePVC() *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         fExpandPVC,
			Tasks:        []tasktypes.TaskName{tExpandPVC, tWaitForPVCResized},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.StartOver,
			},
		},
	}
}
