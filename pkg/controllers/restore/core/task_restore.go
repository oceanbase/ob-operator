/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"fmt"
	"strconv"
	"strings"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	restoreconst "github.com/oceanbase/ob-operator/pkg/controllers/restore/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	tenantCore "github.com/oceanbase/ob-operator/pkg/controllers/tenant/core"
	"github.com/pkg/errors"
)

func (ctrl *RestoreCtrl) PrepareForRestore() ([]string, error) {

	resourcePools := make([]string, 0)
	tenantCtrl := tenantCore.TenantCtrl{
		Resource: ctrl.Resource,
		Tenant:   cloudv1.Tenant{},
	}
	for _, zone := range ctrl.Restore.Spec.Dest.Topology {
		err := tenantCtrl.CheckAndCreateUnitAndPool(ctrl.Restore.Spec.Dest.Tenant, zone)
		if err != nil {
			return resourcePools, errors.Wrap(err, "failed to prepare restore")
		}
		poolName := tenantCtrl.GeneratePoolName(ctrl.Restore.Spec.Dest.Tenant, zone.ZoneName)
		resourcePools = append(resourcePools, poolName)
	}
	return resourcePools, nil
}

func (ctrl *RestoreCtrl) DoRestore(pools []string) error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to do restore ")
	}
	spec := ctrl.Restore.Spec
	restoreOption := ctrl.GetRestoreOption()
	path := spec.Source.Path.Root
	if path == "" {
		path = fmt.Sprintf("%s,%s", spec.Source.Path.Data, spec.Source.Path.Log)
	}
	savePoint := spec.SavePoint.Value
	if spec.SavePoint.Type != "" {
		savePoint = fmt.Sprintf("%s=%s", spec.SavePoint.Type, spec.SavePoint.Value)
	}
	return sqlOperator.DoRestore(spec.Dest.Tenant, spec.Source.Tenant, path, savePoint, spec.Source.ClusterName, strconv.FormatInt(spec.Source.ClusterID, 10), strings.Join(pools, ","), restoreOption)
}

func (ctrl *RestoreCtrl) GetRestoreOption() string {
	tenantCtrl := tenantCore.TenantCtrl{
		Resource: ctrl.Resource,
		Tenant:   cloudv1.Tenant{},
	}
	localityOption := fmt.Sprintf("locality=%s", tenantCtrl.GenerateLocality(ctrl.Restore.Spec.Dest.Topology))
	primaryZoneOption := fmt.Sprintf("primary_zone=%s", tenantCtrl.GenerateSpecPrimaryZone(ctrl.Restore.Spec.Dest.Topology))

	restoreOption := fmt.Sprintf("%s&%s", localityOption, primaryZoneOption)

	if ctrl.Restore.Spec.Dest.KmsEncryptInfo != "" {
		kmsEncryptInfoOption := fmt.Sprintf("kms_encrypt=%s", ctrl.Restore.Spec.Dest.KmsEncryptInfo)
		restoreOption = fmt.Sprintf("%s&%s", restoreOption, kmsEncryptInfoOption)
	}
	return restoreOption
}

func (ctrl *RestoreCtrl) GetRestoreSetCurrentFromDB() ([]model.RestoreStatus, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator when trying to Get RestoreSetCurrent From DB")
	}
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSetCurrent := sqlOperator.GetAllRestoreHistorySet()
	allRestoreSet := make([]model.RestoreStatus, 0)
	allRestoreSet = append(allRestoreSet, restoreSetCurrent...)
	allRestoreSet = append(allRestoreSet, restoreSetHistory...)
	return allRestoreSet, nil
}

func (ctrl *RestoreCtrl) GetRestoreSetHistoryFromDB() ([]model.RestoreStatus, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator when trying to Get RestoreSetHistory From DB")
	}
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSetCurrent := sqlOperator.GetAllRestoreHistorySet()
	allRestoreSet := make([]model.RestoreStatus, 0)
	allRestoreSet = append(allRestoreSet, restoreSetCurrent...)
	allRestoreSet = append(allRestoreSet, restoreSetHistory...)
	return allRestoreSet, nil
}

func (ctrl *RestoreCtrl) getParameter(name string) string {
	params := ctrl.Restore.Spec.Parameters
	for _, parameter := range params {
		if parameter.Name == name {
			return parameter.Value
		}
	}
	return ""
}

func (ctrl *RestoreCtrl) isConcurrencyZero() (error, bool) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether restore_concurrency is 0 "), false
	}
	valueList := sqlOperator.GetRestoreConcurrency()
	for _, value := range valueList {
		if value.Value == restoreconst.RestoreConcurrencyZero {
			return nil, true
		}
	}
	return nil, false
}

func (ctrl *RestoreCtrl) SetParameter(param cloudv1.Parameter) error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set parameter "+param.Name)
	}
	return sqlOperator.SetParameter(param.Name, param.Value)
}
