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
	ttypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

const (
	fMaintainWhiteList   ttypes.FlowName = "maintain white list"
	fMaintainCharset     ttypes.FlowName = "maintain charset"
	fMaintainUnitNum     ttypes.FlowName = "maintain unit num"
	fMaintainLocality    ttypes.FlowName = "maintain locality"
	fMaintainPrimaryZone ttypes.FlowName = "maintain primary zone"
	fMaintainUnitConfig  ttypes.FlowName = "maintain unit config"

	fCreateTenant             ttypes.FlowName = "create tenant"
	fAddPool                  ttypes.FlowName = "add pool"
	fDeletePool               ttypes.FlowName = "delete pool"
	fDeleteTenant             ttypes.FlowName = "delete tenant"
	fRestoreTenant            ttypes.FlowName = "Restore tenant"
	fCancelRestoreFlow        ttypes.FlowName = "cancel restore"
	fCreateEmptyStandbyTenant ttypes.FlowName = "create empty standby tenant"
)

const (
	tCheckTenant                     ttypes.TaskName = "create tenant check"
	tCheckPoolAndUnitConfig          ttypes.TaskName = "create pool and unit config check"
	tCreateTenant                    ttypes.TaskName = "create tenant"
	tCreateResourcePoolAndUnitConfig ttypes.TaskName = "create resource pool and unit config"

	tMaintainWhiteList   ttypes.TaskName = "maintain white list"
	tMaintainCharset     ttypes.TaskName = "maintain charset"
	tMaintainUnitNum     ttypes.TaskName = "maintain unit num"
	tMaintainLocality    ttypes.TaskName = "maintain locality"
	tMaintainPrimaryZone ttypes.TaskName = "maintain primary zone"
	tAddResourcePool     ttypes.TaskName = "add resource pool"
	tDeleteResourcePool  ttypes.TaskName = "delete resource pool"
	tMaintainUnitConfig  ttypes.TaskName = "maintain unit config"
	tDeleteTenant        ttypes.TaskName = "delete tenant"

	tCreateRestoreJobCR            ttypes.TaskName = "create restore job CR"
	tWatchRestoreJobToFinish       ttypes.TaskName = "watch restore job to finish"
	tCancelRestoreJob              ttypes.TaskName = "cancel restore job"
	tCreateUsersByCredentials      ttypes.TaskName = "create users by credentials"
	tCheckPrimaryTenantLSIntegrity ttypes.TaskName = "check primary tenant ls integrity"
	tCreateEmptyStandbyTenant      ttypes.TaskName = "create empty standby tenant"
	tUpgradeTenantIfNeeded         ttypes.TaskName = "upgrade tenant if needed"
)
