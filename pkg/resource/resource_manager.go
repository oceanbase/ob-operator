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

package resource

import (
	"github.com/oceanbase/ob-operator/pkg/task"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
)

type ResourceManager interface {
	// IsNewResource 判断是否为新创建资源
	IsNewResource() bool
	// IsDeleting 判断是否正在删除
	IsDeleting() bool
	// CheckAndUpdateFinalizers 检查并且更新 finalizers
	CheckAndUpdateFinalizers() error
	// InitStatus 初始化资源状态
	InitStatus()
	// SetOperationContext 设置任务流的上下文
	SetOperationContext(*v1alpha1.OperationContext)
	// ClearTaskInfo 清除任务信息
	ClearTaskInfo()
	// FinishTask 完成任务
	FinishTask()
	// UpdateStatus 更新资源状态
	UpdateStatus() error
	// GetTaskFunc 获取任务函数
	GetTaskFunc(string) (func() error, error)
	// GetTaskFlow 根据资源状态获取任务流
	GetTaskFlow() (*task.TaskFlow, error)
}
