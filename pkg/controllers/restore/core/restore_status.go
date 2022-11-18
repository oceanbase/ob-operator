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
	"context"
	"reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"

	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *RestoreCtrl) UpdateRestoreStatus() error {
	restore := ctrl.Restore
	restoreExecuter := resource.NewRestoreResource(ctrl.Resource)
	restoreTmp, err := restoreExecuter.Get(context.TODO(), restore.Namespace, restore.Name)
	if err != nil {
		return err
	}
	restoreCurrent := restoreTmp.(cloudv1.Restore)
	restoreCurrentDeepCopy := restoreCurrent.DeepCopy()

	ctrl.Restore = *restoreCurrentDeepCopy
	restoreNew, err := ctrl.buildRestoreStatus(*restoreCurrentDeepCopy)
	if err != nil {
		return err
	}
	compareStatus := reflect.DeepEqual(restoreCurrent.Status, restoreNew.Status)
	if !compareStatus {
		err = restoreExecuter.UpdateStatus(context.TODO(), restoreNew)
		if err != nil {
			return err
		}
	}
	ctrl.Restore = restoreNew
	return nil
}

func (ctrl *RestoreCtrl) buildRestoreStatus(restore cloudv1.Restore) (cloudv1.Restore, error) {
	var restoreCurrentStatus cloudv1.RestoreStatus
	restoreSetList, err := ctrl.buildRestoreSetListFromDB()
	if err != nil {
		return restore, err
	}
	restoreSetStatus := ctrl.RestoreSetListToStatusList(restoreSetList)
	if err != nil {
		return restore, err
	}
	restoreCurrentStatus.RestoreSet = restoreSetStatus
	restore.Status = restoreCurrentStatus
	return restore, nil
}

func (ctrl *RestoreCtrl) buildRestoreSetListFromDB() ([]model.AllRestoreSet, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, err
	}
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSetCurrent := sqlOperator.GetAllRestoreCurrentSet()
	allRestoreSet := make([]model.AllRestoreSet, 0)
	allRestoreSet = append(allRestoreSet, restoreSetCurrent...)
	allRestoreSet = append(allRestoreSet, restoreSetHistory...)
	return allRestoreSet, nil
}

func (ctrl *RestoreCtrl) RestoreSetListToStatusList(restoreSetList []model.AllRestoreSet) []cloudv1.RestoreSetSpec {
	restoreSetStatusList := make([]cloudv1.RestoreSetSpec, 0)
	for _, restoreSet := range restoreSetList {
		restoreSetStatus := cloudv1.RestoreSetSpec{}
		restoreSetStatus.JodID = int(restoreSet.JobId)
		restoreSetStatus.ClusterID = int(restoreSet.BackupClusterId)
		restoreSetStatus.ClusterName = ctrl.Restore.Spec.SourceCluster.ClusterName
		restoreSetStatus.TenantName = restoreSet.TenantName
		restoreSetStatus.BackupTenantName = restoreSet.BackupTenantName
		restoreSetStatus.Status = restoreSet.Status
		restoreSetStatus.Timestamp = restoreSet.RestoreFinishTimestamp
		restoreSetStatusList = append(restoreSetStatusList, restoreSetStatus)
	}
	return restoreSetStatusList
}
