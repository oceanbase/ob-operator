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
	tCheckTenantTask                 ttypes.TaskName = "create tenant check"
	tCheckPoolAndConfigTask          ttypes.TaskName = "create pool and unit config check"
	tCreateTenantTaskWithClear       ttypes.TaskName = "create tenant"
	tCreateResourcePoolAndConfigTask ttypes.TaskName = "create resource pool and unit config"
	tCheckAndApplyWhiteList          ttypes.TaskName = "maintain white list"
	tCheckAndApplyCharset            ttypes.TaskName = "maintain charset"
	tCheckAndApplyUnitNum            ttypes.TaskName = "maintain unit num"
	tCheckAndApplyLocality           ttypes.TaskName = "maintain locality"
	tCheckAndApplyPrimaryZone        ttypes.TaskName = "maintain primary zone"
	tAddPoolTask                     ttypes.TaskName = "add resource pool"
	tDeletePoolTask                  ttypes.TaskName = "delete resource pool"
	tMaintainUnitConfigTask          ttypes.TaskName = "maintain unit config"
	tDeleteTenantTask                ttypes.TaskName = "delete tenant"
	tCreateTenantRestoreJobCR        ttypes.TaskName = "create restore job CR"
	tWatchRestoreJobToFinish         ttypes.TaskName = "watch restore job to finish"
	tCancelTenantRestoreJob          ttypes.TaskName = "cancel restore job"
	tCreateUserWithCredentialSecrets ttypes.TaskName = "create users by credentials"
	tCheckPrimaryTenantLSIntegrity   ttypes.TaskName = "check primary tenant ls integrity"
	tCreateEmptyStandbyTenant        ttypes.TaskName = "create empty standby tenant"
	tUpgradeTenantIfNeeded           ttypes.TaskName = "upgrade tenant if needed"
)
