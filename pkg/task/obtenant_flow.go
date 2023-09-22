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
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	tenantstatus "github.com/oceanbase/ob-operator/pkg/const/status/tenantstatus"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/strategy"
)

func CreateTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name: flowname.CreateTenant,
			Tasks: []string{taskname.CheckTenant, taskname.CheckPoolAndUnitConfig,
				taskname.CreateResourcePoolAndUnitConfig, taskname.CreateTenant},
			TargetStatus: tenantstatus.Running,
			OnFailure: strategy.FailureRule{
				NextTryStatus: tenantstatus.CreatingTenant,
			},
		},
	}
}

func MaintainWhiteList() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainWhiteList,
			Tasks:        []string{taskname.MaintainWhiteList},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainCharset() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainCharset,
			Tasks:        []string{taskname.MaintainCharset},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainUnitNum() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainUnitNum,
			Tasks:        []string{taskname.MaintainUnitNum},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainLocality() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainLocality,
			Tasks:        []string{taskname.MaintainLocality},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainPrimaryZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainPrimaryZone,
			Tasks:        []string{taskname.MaintainPrimaryZone},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func AddPool() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.AddPool,
			Tasks:        []string{taskname.CheckPoolAndUnitConfig, taskname.AddResourcePool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func DeletePool() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeletePool,
			Tasks:        []string{taskname.DeleteResourcePool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainUnitConfig() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.MaintainUnitConfig,
			Tasks:        []string{taskname.MaintainUnitConfig},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func DeleteTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeleteTenant,
			Tasks:        []string{taskname.DeleteTenant},
			TargetStatus: tenantstatus.FinalizerFinished,
			OnFailure: strategy.FailureRule{
				NextTryStatus: tenantstatus.DeletingTenant,
			},
		},
	}
}

func RestoreTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name: flowname.RestoreTenant,
			Tasks: []string{
				taskname.CheckTenant,
				taskname.CheckPoolAndUnitConfig,
				taskname.CreateResourcePoolAndUnitConfig,
				taskname.CreateRestoreJobCR,
				taskname.WatchRestoreJobToFinish,
			},
			TargetStatus: tenantstatus.Running,
			OnFailure: strategy.FailureRule{
				NextTryStatus: tenantstatus.Restoring,
			},
		},
	}
}

func CancelRestoreJob() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name: flowname.CancelRestoreFlow,
			Tasks: []string{
				taskname.CancelRestoreJob,
			},
			TargetStatus: tenantstatus.RestoreCanceled,
		},
	}
}
