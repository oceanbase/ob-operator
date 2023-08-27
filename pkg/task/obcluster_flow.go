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
	clusterstatus "github.com/oceanbase/ob-operator/pkg/const/status/obcluster"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/fail"
)

func BootstrapOBCluster() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.BootstrapOBCluster,
			Tasks:        []string{taskname.CreateOBZone, taskname.WaitOBZoneBootstrapReady, taskname.Bootstrap},
			TargetStatus: clusterstatus.Bootstrapped,
		},
	}
}

func MaintainOBClusterAfterBootstrap() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainOBClusterAfterBootstrap,
			Tasks:        []string{taskname.WaitOBZoneRunning, taskname.CreateUsers, taskname.MaintainOBParameter},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func AddOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.AddOBZone,
			Tasks:        []string{taskname.CreateOBZone, taskname.WaitOBZoneRunning},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func DeleteOBZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeleteOBZone,
			Tasks:        []string{taskname.DeleteOBZone, taskname.WaitOBZoneDeleted},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func ModifyOBZoneReplica() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.ModifyOBZoneReplica,
			Tasks:        []string{taskname.ModifyOBZoneReplica, taskname.WaitOBZoneTopologyMatch, taskname.WaitOBZoneRunning},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func MaintainOBParameter() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainOBParameter,
			Tasks:        []string{taskname.MaintainOBParameter},
			TargetStatus: clusterstatus.Running,
		},
	}
}

func UpgradeOBCluster() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.UpgradeOBCluster,
			Tasks:        []string{taskname.ValidateUpgradeInfo, taskname.BackupEssentialParameters, taskname.UpgradeCheck, taskname.BeginUpgrade, taskname.RollingUpgradeByZone, taskname.FinishUpgrade, taskname.RestoreEssentialParameters},
			TargetStatus: clusterstatus.Running,
			FailureRule: fail.FailureRule{
				Strategy:      fail.PauseReconcile,
			},
		},
	}
}
