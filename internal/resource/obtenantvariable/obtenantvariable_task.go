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

package obtenantvariable

import (
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
	"github.com/pkg/errors"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*OBTenantVariableManager]()

func SetOBTenantVariable(m *OBTenantVariableManager) tasktypes.TaskError {
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		m.Logger.Error(err, "Get operation manager failed")
		return errors.Wrapf(err, "Get operation manager")
	}
	err = operationManager.SetGlobalVariable(m.Ctx, m.OBTenantVariable.Spec.Variable.Name, m.OBTenantVariable.Spec.Variable.Value)
	if err != nil {
		m.Logger.Error(err, "Set tenant variable failed")
		return errors.Wrapf(err, "Set tenant variable")
	}
	return nil
}
