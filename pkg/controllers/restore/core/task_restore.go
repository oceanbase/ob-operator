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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	restoreconst "github.com/oceanbase/ob-operator/pkg/controllers/restore/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

func (ctrl *RestoreCtrl) CreateResourcePool() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether restore_concurrency is 0 ")
	}
	resourcePoolSpec := ctrl.Restore.Spec.ResourcePool
	return sqlOperator.CreateResourcePool(resourcePoolSpec.Name, resourcePoolSpec.UnitName, resourcePoolSpec.UnitNum, resourcePoolSpec.ZoneList)

}

func (ctrl *RestoreCtrl) CreateResourceUnit() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether restore_concurrency is 0 ")
	}
	resourceUnitSpec := ctrl.Restore.Spec.ResourceUnit

	return sqlOperator.CreateResourceUnit(resourceUnitSpec.Name, strconv.Itoa(resourceUnitSpec.MaxCPU), resourceUnitSpec.MaxMemory, strconv.Itoa(resourceUnitSpec.MaxIops), resourceUnitSpec.MaxDiskSize, strconv.Itoa(resourceUnitSpec.MaxSessionNum), strconv.Itoa(resourceUnitSpec.MinCPU), resourceUnitSpec.MinMemory, strconv.Itoa(resourceUnitSpec.MinIops))
}

func (ctrl *RestoreCtrl) DoResotre() error {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when checking whether restore_concurrency is 0 ")
	}
	spec := ctrl.Restore.Spec

	return sqlOperator.DoResotre(spec.DestTenant, spec.SourceTenant, spec.Path, spec.Timestamp, spec.SourceCluster.ClusterName, strconv.Itoa(spec.SourceCluster.ClusterID), spec.PoolList)
}

func (ctrl *RestoreCtrl) GetRestoreSetFromDB() ([]model.AllRestoreSet, error) {
	klog.Infoln("Check whether backup is doing")
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator when checking whether backup is doing")
	}
	return sqlOperator.GetAllRestoreSet(), nil
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
	klog.Infoln("begin set parameter ", param.Name)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when trying to set backup dest option")
	}
	return sqlOperator.SetParameter(param.Name, param.Value)
}
