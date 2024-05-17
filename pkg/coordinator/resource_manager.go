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

package coordinator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type ResourceManager interface {
	ResourceUpdater
	TaskHandler
}

type TaskHandler interface {
	SetOperationContext(*tasktypes.OperationContext)
	ClearTaskInfo()
	HandleFailure()
	FinishTask()
	GetTaskFunc(tasktypes.TaskName) (tasktypes.TaskFunc, error)
	GetTaskFlow() (*tasktypes.TaskFlow, error)
	PrintErrEvent(error)
}

type ResourceUpdater interface {
	CheckAndUpdateFinalizers() error
	InitStatus()
	UpdateStatus() error
	GetStatus() string
	ArchiveResource()
	GetMeta() metav1.Object
}
