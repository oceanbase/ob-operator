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

package _template

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/internal/telemetry"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/task"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type ObResourceManager[T client.Object] struct {
	opresource.ResourceManager

	Ctx      context.Context
	Resource T
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger

	con *operation.OceanbaseOperationManager
}

func (m *ObResourceManager[T]) IsNewResource() bool {
	return false
}

func (m *ObResourceManager[T]) GetStatus() string {
	return ""
}

func (m *ObResourceManager[T]) IsDeleting() bool {
	return false
}

func (m *ObResourceManager[T]) CheckAndUpdateFinalizers() error {
	return nil
}

func (m *ObResourceManager[T]) InitStatus() {}

func (m *ObResourceManager[T]) SetOperationContext(*tasktypes.OperationContext) {

}

func (m *ObResourceManager[T]) ClearTaskInfo() {}

func (m *ObResourceManager[T]) HandleFailure() {}

func (m *ObResourceManager[T]) FinishTask() {}

func (m *ObResourceManager[T]) UpdateStatus() error {
	return m.Client.Status().Update(m.Ctx, m.Resource)
}

func (m *ObResourceManager[T]) GetTaskFunc(string) (func() error, error) {
	return nil, nil
}

func (m *ObResourceManager[T]) GetTaskFlow() (*task.TaskFlow, error) {
	return nil, nil
}

func (m *ObResourceManager[T]) PrintErrEvent(err error) {
	m.Recorder.Event(m.Resource, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *ObResourceManager[T]) ArchiveResource() {
}
