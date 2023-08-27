package task

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	tenantstatus "github.com/oceanbase/ob-operator/pkg/const/status/tenantstatus"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/fail"
)

func CreateTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.CreateTenant,
			Tasks:        []string{taskname.CheckTenant, taskname.CheckPoolAndUnitConfig,taskname.CreateTenant},
			TargetStatus: tenantstatus.Running,
			FailureRule: fail.FailureRule {
				NextTryStatus: tenantstatus.Creating,
			},
		},
	}
}

func MaintainWhiteList() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainWhiteList,
			Tasks:        []string{taskname.MaintainWhiteList},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainCharset() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainCharset,
			Tasks:        []string{taskname.MaintainCharset},
			TargetStatus: tenantstatus.Running,
		},
	}
}
func MaintainUnitNum() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainUnitNum,
			Tasks:        []string{taskname.MaintainUnitNum},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainLocality() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainLocality,
			Tasks:        []string{taskname.MaintainLocality},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainPrimaryZone() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainPrimaryZone,
			Tasks:        []string{taskname.MaintainPrimaryZone},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func AddPool() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.AddPool,
			Tasks:        []string{taskname.CheckPoolAndUnitConfig, taskname.AddPool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func DeletePool() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.DeletePool,
			Tasks:        []string{taskname.DeletePool},
			TargetStatus: tenantstatus.Running,
		},
	}
}

func MaintainUnitConfig() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
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
			FailureRule: fail.FailureRule{
				NextTryStatus: tenantstatus.Deleting,
			},
		},
	}
}
