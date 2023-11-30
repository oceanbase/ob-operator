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
	zonestatus "github.com/oceanbase/ob-operator/internal/const/status/obzone"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

func PrepareOBZoneForBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.PrepareOBZoneForBootstrap,
			Tasks:        []string{taskname.CreateOBServer, taskname.WaitOBServerBootstrapReady},
			TargetStatus: zonestatus.BootstrapReady,
		},
	}
}

func MaintainOBZoneAfterBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.MaintainOBZoneAfterBootstrap,
			Tasks:        []string{taskname.WaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func CreateOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.CreateOBZone,
			Tasks:        []string{taskname.AddZone, taskname.StartOBZone, taskname.CreateOBServer, taskname.WaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func AddOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.AddOBServer,
			Tasks:        []string{taskname.CreateOBServer, taskname.WaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.DeleteOBServer,
			Tasks:        []string{taskname.DeleteOBServer, taskname.WaitReplicaMatch},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBZoneFinalizer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.DeleteOBZoneFinalizer,
			Tasks:        []string{taskname.StopOBZone, taskname.DeleteAllOBServer, taskname.WaitOBServerDeleted, taskname.DeleteOBZoneInCluster},
			TargetStatus: zonestatus.FinalizerFinished,
		},
	}
}

func UpgradeOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.UpgradeOBZone,
			Tasks:        []string{taskname.OBClusterHealthCheck, taskname.StopOBZone, taskname.UpgradeOBServer, taskname.WaitOBServerUpgraded, taskname.OBZoneHealthCheck, taskname.StartOBZone},
			TargetStatus: zonestatus.Running,
		},
	}
}

func ForceUpgradeOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &tasktypes.OperationContext{
			Name:         flowname.ForceUpgradeOBZone,
			Tasks:        []string{taskname.OBClusterHealthCheck, taskname.UpgradeOBServer, taskname.WaitOBServerUpgraded, taskname.OBZoneHealthCheck},
			TargetStatus: zonestatus.Running,
		},
	}
}
