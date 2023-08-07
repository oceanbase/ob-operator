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
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	zonestatus "github.com/oceanbase/ob-operator/pkg/const/status/obzone"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
)

func PrepareOBZoneForBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.PrepareOBZoneForBootstrap,
			Tasks:        []string{taskname.CreateOBServer, taskname.WaitOBServerBootstrapReady},
			TargetStatus: zonestatus.BootstrapReady,
		},
	}
}

func MaintainOBZoneAfterBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainOBZoneAfterBootstrap,
			Tasks:        []string{taskname.WaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func CreateOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.CreateOBZone,
			Tasks:        []string{taskname.AddZone, taskname.CreateOBServer, taskname.WaitOBServerRunning, taskname.StartOBZone},
			TargetStatus: zonestatus.Running,
		},
	}
}

func AddOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.AddOBServer,
			Tasks:        []string{taskname.CreateOBServer, taskname.WaitOBServerRunning},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBServer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeleteOBServer,
			Tasks:        []string{taskname.DeleteOBServer, taskname.WaitReplicaMatch},
			TargetStatus: zonestatus.Running,
		},
	}
}

func DeleteOBZoneFinalizer() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeleteOBZoneFinalizer,
			Tasks:        []string{taskname.StopOBZone, taskname.DeleteAllOBServer, taskname.WaitOBServerDeleted, taskname.DeleteOBZoneInCluster},
			TargetStatus: zonestatus.FinalizerFinished,
		},
	}
}

func UpgradeOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.UpgradeOBZone,
			Tasks:        []string{taskname.OBClusterHealthCheck, taskname.StopOBZone, taskname.UpgradeOBServer, taskname.WaitOBServerUpgraded, taskname.OBZoneHealthCheck, taskname.StartOBZone},
			TargetStatus: zonestatus.Running,
		},
	}
}
