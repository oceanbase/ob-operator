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
	"strconv"
	"strings"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	restoreconst "github.com/oceanbase/ob-operator/pkg/controllers/restore/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	"github.com/pkg/errors"
)

func (ctrl *RestoreCtrl) CreateResourcePool() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to create resource pool ")
	}
	resourcePoolSpec := ctrl.Restore.Spec.ResourcePool
	var resourcePoolName = restoreconst.ResourcePoolName
	if ctrl.Restore.Spec.ResourcePool.Name != "" {
		resourcePoolName = ctrl.Restore.Spec.ResourcePool.Name
	}
	var resourceUnitName = restoreconst.ResourceUnitName
	if ctrl.Restore.Spec.ResourceUnit.Name != "" {
		resourceUnitName = ctrl.Restore.Spec.ResourceUnit.Name
	}
	var zoneList string
	for _, zone := range resourcePoolSpec.ZoneList {
		zoneList += "'" + zone + "',"
	}
	zoneList = strings.TrimRight(zoneList, ",")
	return sqlOperator.CreateResourcePool(resourcePoolName, resourceUnitName, strconv.Itoa(resourcePoolSpec.UnitNum), zoneList)

}

func (ctrl *RestoreCtrl) CreateResourceUnit() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to create resource unit ")
	}
	resourceUnitSpec := ctrl.Restore.Spec.ResourceUnit
	var resourceUnitName = restoreconst.ResourceUnitName
	if ctrl.Restore.Spec.ResourceUnit.Name != "" {
		resourceUnitName = ctrl.Restore.Spec.ResourceUnit.Name
	}
	return sqlOperator.CreateResourceUnit(resourceUnitName, strconv.Itoa(resourceUnitSpec.MaxCPU), resourceUnitSpec.MaxMemory, strconv.Itoa(resourceUnitSpec.MaxIops), resourceUnitSpec.MaxDiskSize, strconv.Itoa(resourceUnitSpec.MaxSessionNum), strconv.Itoa(resourceUnitSpec.MinCPU), resourceUnitSpec.MinMemory, strconv.Itoa(resourceUnitSpec.MinIops))
}

func (ctrl *RestoreCtrl) DoResotre() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to do restore ")
	}
	spec := ctrl.Restore.Spec
	restoreOption := ctrl.GetRestoreOption()
	var resourcePoolName = restoreconst.ResourcePoolName
	if ctrl.Restore.Spec.ResourcePool.Name != "" {
		resourcePoolName = ctrl.Restore.Spec.ResourcePool.Name
	}
	return sqlOperator.DoResotre(spec.DestTenant, spec.SourceTenant, spec.Path, spec.Timestamp, spec.SourceCluster.ClusterName, strconv.Itoa(spec.SourceCluster.ClusterID), resourcePoolName, restoreOption)
}

func (ctrl *RestoreCtrl) GetRestoreOption() string {
	var locality = cloudv1.Parameter{Name: restoreconst.LocalityName, Value: ""}
	var primaryZone = cloudv1.Parameter{Name: restoreconst.PrimaryZoneName, Value: ""}
	var kmsEncrypt = cloudv1.Parameter{Name: restoreconst.KmsEncryptName, Value: ""}
	paramList := [3]cloudv1.Parameter{locality, primaryZone, kmsEncrypt}
	allParams := ctrl.Restore.Spec.Parameters
	var isSet bool
	for _, p := range paramList {
		for _, param := range allParams {
			if p.Name == param.Name {
				isSet = true
			}
		}
	}
	if !isSet {
		return ""
	}
	var restoreOption = "&"
	for _, p := range paramList {
		if ctrl.getParameter(p.Name) != "" {
			p.Value = ctrl.getParameter(p.Name)
		}
		restoreOption += p.Name + "=" + p.Value + "&"
	}
	restoreOption = strings.TrimRight(restoreOption, "&")
	return restoreOption
}

func (ctrl *RestoreCtrl) GetRestoreSetCurrentFromDB() ([]model.AllRestoreSet, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator when trying to Get RestoreSetCurrent From DB")
	}
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSetCurrent := sqlOperator.GetAllRestoreHistorySet()
	allRestoreSet := make([]model.AllRestoreSet, 0)
	allRestoreSet = append(allRestoreSet, restoreSetCurrent...)
	allRestoreSet = append(allRestoreSet, restoreSetHistory...)
	return allRestoreSet, nil
}

func (ctrl *RestoreCtrl) GetRestoreSetHistoryFromDB() ([]model.AllRestoreSet, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator when trying to Get RestoreSetHistory From DB")
	}
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSetCurrent := sqlOperator.GetAllRestoreHistorySet()
	allRestoreSet := make([]model.AllRestoreSet, 0)
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
