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

func genPrepareOBServerForBootstrapFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "prepare observer for bootstrap",
			Tasks:        []tasktypes.TaskName{tCreateOBServerSvc, tCreateOBServerPVC, tCreateOBServerPod, tWaitOBServerReady},
			TargetStatus: serverstatus.BootstrapReady,
		},
	}
}

func genMaintainOBServerAfterBootstrapFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "maintain observer after bootstrap",
			Tasks:        []tasktypes.TaskName{tWaitOBClusterBootstrapped, tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func genCreateOBServerFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "create observer",
			Tasks:        []tasktypes.TaskName{tCreateOBServerSvc, tCreateOBServerPVC, tCreateOBServerPod, tWaitOBServerReady, tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				NextTryStatus: "Failed",
			},
		},
	}
}

func genDeleteOBServerFinalizerFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "delete observer finalizer",
			Tasks:        []tasktypes.TaskName{tDeleteOBServerInCluster, tWaitOBServerDeletedInCluster},
			TargetStatus: serverstatus.FinalizerFinished,
		},
	}
}

func genUpgradeOBServerFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "upgrade observer",
			Tasks:        []tasktypes.TaskName{tUpgradeOBServerImage, tWaitOBServerPodReady, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func genRecoverOBServerFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "recover observer",
			Tasks:        []tasktypes.TaskName{tCreateOBServerPod, tWaitOBServerReady, tAddServer, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.RetryFromCurrent,
			},
		},
	}
}

func genAnnotateOBServerPodFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "annotate observer pod",
			Tasks:        []tasktypes.TaskName{tAnnotateOBServerPod},
			TargetStatus: serverstatus.Running,
		},
	}
}

func genScaleUpOBServerFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "scale up observer",
			Tasks:        []tasktypes.TaskName{tDeletePod, tWaitForPodDeleted, tCreateOBServerPod, tWaitOBServerReady, tWaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func genExpandPVCFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "expand pvc",
			Tasks:        []tasktypes.TaskName{tExpandPVC, tWaitForPvcResized},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.StartOver,
			},
		},
	}
}

func genModifyPodTemplateFlow(_ *OBServerManager) *tasktypes.TaskFlow {
	return &tasktypes.TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         "modify pod template",
			Tasks:        []tasktypes.TaskName{tDeletePod, tWaitForPodDeleted, tCreateOBServerPod, tWaitOBServerReady},
			TargetStatus: serverstatus.Running,
			OnFailure: tasktypes.FailureRule{
				Strategy: strategy.StartOver,
			},
		},
	}
}
