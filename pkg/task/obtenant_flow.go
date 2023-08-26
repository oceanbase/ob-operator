package task

import (
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	obtenantstatus "github.com/oceanbase/ob-operator/pkg/const/status/obtenant"
	flowname "github.com/oceanbase/ob-operator/pkg/task/const/flow/name"
	taskname "github.com/oceanbase/ob-operator/pkg/task/const/task/name"
	"github.com/oceanbase/ob-operator/pkg/task/fail"
)

func CreateTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.CreateTenant,
			Tasks:        []string{taskname.CreateTenant},
			TargetStatus: obtenantstatus.Running,
			FailureRule: &fail.FailureRule {
				FailureStatus: obtenantstatus.Pending,
			},
		},
	}
}

func MaintainTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext {
			Name:         flowname.MaintainTenant,
			Tasks:        []string{taskname.MaintainTenant},
			TargetStatus: obtenantstatus.Running,
		},
	}
}

func DeleteTenant() *TaskFlow {
	return &TaskFlow{
		OperationContext: &v1alpha1.OperationContext{
			Name:         flowname.DeleteTenant,
			Tasks:        []string{taskname.DeleteTenant},
			TargetStatus: obtenantstatus.FinalizerFinished,
		},
	}
}
