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
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/controllers/restore/model"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *RestoreCtrl) UpdateRestoreStatusFromDB() error {
	restoreStatus, err := ctrl.getRestoreStatusFromDB()
	if err != nil {
		return errors.Wrap(err, "get restore status from OB")
	}
	return ctrl.UpdateRestoreStatus(restoreStatus)
}

func (ctrl *RestoreCtrl) UpdateRestoreStatus(restoreStatus *cloudv1.RestoreStatus) error {
	restoreExecuter := resource.NewRestoreResource(ctrl.Resource)
	restoreTmp, err := restoreExecuter.Get(context.TODO(), ctrl.Restore.Namespace, ctrl.Restore.Name)
	if err != nil {
		return err
	}
	restoreCurrent := restoreTmp.(cloudv1.Restore)
	restoreCurrentDeepCopy := restoreCurrent.DeepCopy()
	restoreCurrentDeepCopy.Status = *restoreStatus
	ctrl.Restore = *restoreCurrentDeepCopy
	if !reflect.DeepEqual(restoreCurrent.Status, ctrl.Restore.Status) {
		err = restoreExecuter.UpdateStatus(context.TODO(), ctrl.Restore)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *RestoreCtrl) getRestoreStatusFromDB() (*cloudv1.RestoreStatus, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator to get restore status")
	}
	restoreSetCurrent := sqlOperator.GetAllRestoreCurrentSet()
	restoreSetHistory := sqlOperator.GetAllRestoreHistorySet()
	restoreSet := make([]model.RestoreStatus, 0)
	restoreSet = append(restoreSet, restoreSetCurrent...)
	restoreSet = append(restoreSet, restoreSetHistory...)
	for _, restoreRecord := range restoreSet {
		if restoreRecord.BackupClusterName == ctrl.Restore.Spec.Source.ClusterName &&
			restoreRecord.BackupTenantName == ctrl.Restore.Spec.Source.Tenant &&
			restoreRecord.RestoreTenantName == ctrl.Restore.Spec.Dest.Tenant {
			return &cloudv1.RestoreStatus{
				JobID:           restoreRecord.JobId,
				Status:          restoreRecord.Status,
				FinishTimestamp: restoreRecord.FinishTimestamp,
			}, nil
		}
	}
	return nil, errors.New("no record found for restore")
}
