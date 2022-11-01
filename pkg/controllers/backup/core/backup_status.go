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
	// "context"
	// "reflect"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	//"github.com/oceanbase/ob-operator/pkg/controllers/backup/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/backup/model"
	//"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *BackupCtrl) UpdateBackupSetStatus() error {
	// backup := ctrl.Backup
	// backupExecuter := resource.NewBackupResource(ctrl.Resource)
	// backupCurrent, err := backupExecuter.Get(context.TODO(), backup.Namespace, backup.Spec.SourceCluster[0].ClusterName)
	// if err != nil {
	// 	return err
	// }
	// sqlOperator, err := ctrl.GetSqlOperator()
	// if err != nil {
	// 	return err
	// }
	// backupSetList := make([]model.AllBackupSet, 0, 0)
	// backupSetList = sqlOperator.GetAllBackupSet()
	// //
	// //backupSetStatus := converter.BackupSetListToStatus(backupCurrent, backupSetList)
	// backupSetStatus := ctrl.BackupSetListToStatus(backupCurrent, backupSetList)

	// status := reflect.DeepEqual(backupSetStatus, backupCurrent.Status)

	return nil

}

// to do
func (ctrl *BackupCtrl) BackupSetListToStatus(backupCurrent interface{}, backupSetList []model.AllBackupSet) cloudv1.BackupSetStatus {
	//backupList := ctrl.buildBackupSetListFromDB()
	bakupSetStatus := cloudv1.BackupSetStatus{}
	return bakupSetStatus

}

// to do
func (ctrl *BackupCtrl) buildBackupSetStatus(backupCurrent interface{}, backupSetList []model.AllBackupSet) cloudv1.BackupSetStatus {
	//backupList := ctrl.buildBackupSetListFromDB()
	bakupSetStatus := cloudv1.BackupSetStatus{}
	return bakupSetStatus

}

func (ctrl *BackupCtrl) buildBackupSetListFromDB() []model.AllBackupSet {
	AllBackupSetList := make([]model.AllBackupSet, 0)
	sqlOperator, err := ctrl.GetSqlOperator()
	if err == nil {
		AllBackupSetList = sqlOperator.GetAllBackupSet()
	}
	return AllBackupSetList
}
