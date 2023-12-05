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

package obtenant

import (
	"github.com/oceanbase/ob-operator/pkg/task"
)

func init() {
	// obtenant
	task.GetRegistry().Register(fCreateTenant, CreateTenant)
	task.GetRegistry().Register(fMaintainWhiteList, MaintainWhiteList)
	task.GetRegistry().Register(fMaintainCharset, MaintainCharset)
	task.GetRegistry().Register(fMaintainUnitNum, MaintainUnitNum)
	task.GetRegistry().Register(fMaintainPrimaryZone, MaintainPrimaryZone)
	task.GetRegistry().Register(fMaintainLocality, MaintainLocality)
	task.GetRegistry().Register(fAddPool, AddPool)
	task.GetRegistry().Register(fDeletePool, DeletePool)
	task.GetRegistry().Register(fMaintainUnitConfig, MaintainUnitConfig)
	task.GetRegistry().Register(fDeleteTenant, DeleteTenant)

	task.GetRegistry().Register(fRestoreTenant, RestoreTenant)
	task.GetRegistry().Register(fCancelRestoreFlow, CancelRestoreJob)
	task.GetRegistry().Register(fCreateEmptyStandbyTenant, CreateEmptyStandbyTenant)
}
