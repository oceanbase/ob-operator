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

package task

import (
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func PrepareOBServerForBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.PrepareOBServerForBootstrap,
			Tasks:        []string{taskname.CreateOBPVC, taskname.CreateOBPod, taskname.WaitOBServerReady},
			TargetStatus: serverstatus.BootstrapReady,
		},
	}
}

func MaintainOBServerAfterBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.PrepareOBServerForBootstrap,
			Tasks:        []string{taskname.WaitOBClusterBootstrapped, taskname.AddServer, taskname.WaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func CreateOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.CreateOBServer,
			Tasks:        []string{taskname.CreateOBPVC, taskname.CreateOBPod, taskname.WaitOBServerReady, taskname.AddServer, taskname.WaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func DeleteOBServerFinalizer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.DeleteOBServerFinalizer,
			Tasks:        []string{taskname.DeleteOBServerInCluster, taskname.WaitOBServerDeletedInCluster},
			TargetStatus: serverstatus.FinalizerFinished,
		},
	}
}

func UpgradeOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.UpgradeOBServer,
			Tasks:        []string{taskname.UpgradeOBServerImage, taskname.WaitOBServerPodReady, taskname.WaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func RecoverOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.RecoverOBServer,
			Tasks:        []string{taskname.CreateOBPod, taskname.WaitOBServerReady, taskname.WaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func AddServerInOB() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.AddServerInOB,
			Tasks:        []string{taskname.AddServer, taskname.WaitOBServerActiveInCluster},
			TargetStatus: serverstatus.Running,
		},
	}
}

func AnnotateOBServerPod() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.AnnotateOBServerPod,
			Tasks:        []string{taskname.AnnotateOBServerPod},
			TargetStatus: serverstatus.Running,
		},
	}
}
