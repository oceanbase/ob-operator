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
	"fmt"
	"strings"

	"github.com/pkg/errors"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/model"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func (ctrl *TenantBackupCtrl) GetTenantSecret(tenant cloudv1.TenantSpec) (model.TenantSecret, error) {
	var tenantSecret model.TenantSecret
	obcluster := ctrl.TenantBackup.Spec.SourceCluster
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	userSecret, err := secretExecutor.Get(context.TODO(), obcluster.ClusterNamespace, tenant.UserSecret)
	if err != nil {
		klog.Errorf("get tenant '%s' user secret error '%s'", tenant.Name, err)
		return tenantSecret, err
	}
	backupSecret, err := secretExecutor.Get(context.TODO(), obcluster.ClusterNamespace, tenant.BackupSecret)
	if err != nil {
		klog.Errorf("get tenant '%s' backup secret error '%s'", tenant.Name, err)
		return tenantSecret, err
	}
	tenantSecret.User = strings.Replace(string(userSecret.(corev1.Secret).Data[tenantBackupconst.User]), "\n", "", -1)
	tenantSecret.UserSecret = strings.Replace(string(userSecret.(corev1.Secret).Data[tenantBackupconst.UserSecret]), "\n", "", -1)
	tenantSecret.IncrementalSecret = strings.Replace(string(backupSecret.(corev1.Secret).Data[tenantBackupconst.IncrementalSecret]), "\n", "", -1)
	tenantSecret.DatabaseSecret = strings.Replace(string(backupSecret.(corev1.Secret).Data[tenantBackupconst.DatabaseSecret]), "\n", "", -1)
	return tenantSecret, nil
}

func (ctrl *TenantBackupCtrl) GetTenantSqlOperator(tenant cloudv1.TenantSpec) (*sql.SqlOperator, error) {
	tenantSecret, err := ctrl.GetTenantSecret(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' secret error '%s'", tenant.Name, err)
		return nil, err
	}
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.TenantBackup.Namespace, ctrl.TenantBackup.Spec.SourceCluster.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	p := &sql.DBConnectProperties{
		IP:       clusterIP,
		Port:     observerconst.MysqlPort,
		User:     fmt.Sprint(tenantSecret.User, "@", tenant.Name),
		Password: tenantSecret.UserSecret,
		Database: "oceanbase",
		Timeout:  10,
	}
	so := sql.NewSqlOperator(p)
	if so.TestOK() {
		return so, nil
	}
	return nil, errors.New("failed to get tenant sql operator")
}

func (ctrl *TenantBackupCtrl) CheckAndSetLogArchiveDest(tenant cloudv1.TenantSpec) error {
	logArchiveDest, err := ctrl.GetLogArchiveDest(tenant)
	if err != nil {
		klog.Errorf("get tenant '%s' LogArchiveDest error '%s'", tenant.Name, err)
		return err
	}
	if ctrl.NeedSetArchiveDest(tenant, logArchiveDest) {
		return ctrl.SetArchiveDest(tenant)
	}
	return nil
}

func (ctrl *TenantBackupCtrl) SetArchiveDest(tenant cloudv1.TenantSpec) error {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' get sql operator error when set LogArchiveDest", tenant.Name)
		return errors.Wrap(err, "get sql operator error when set LogArchiveDest")
	}
	value := fmt.Sprint("LOCATION=", tenant.LogArchiveDest)
	if tenant.Binding != "" {
		value = fmt.Sprint(value, " BINDING=", tenant.Binding)
	}
	if tenant.PieceSwitchInterval != "" {
		value = fmt.Sprint(value, " PIECE_SWITCH_INTERVAL=", tenant.PieceSwitchInterval)
	}
	return sqlOperator.SetParameter(tenantBackupconst.LogAechiveDest, value)
}

func (ctrl *TenantBackupCtrl) GetLogArchiveDest(tenant cloudv1.TenantSpec) ([]model.TenantArchiveDest, error) {
	sqlOperator, err := ctrl.GetTenantSqlOperator(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "get sql operator error when get LogArchiveDest")
	}
	return sqlOperator.GetArchieveLogDest(), nil
}

func (ctrl *TenantBackupCtrl) NeedSetArchiveDest(tenant cloudv1.TenantSpec, logArchiveDestList []model.TenantArchiveDest) bool {
	if len(logArchiveDestList) == 0 {
		return true
	}
	for _, logArchiveDest := range logArchiveDestList {
		if (logArchiveDest.Name == tenantBackupconst.Path && logArchiveDest.Value != tenant.LogArchiveDest) ||
			(logArchiveDest.Name == tenantBackupconst.Binding && logArchiveDest.Value != tenant.Binding) ||
			(logArchiveDest.Name == tenantBackupconst.PieceSwitchInterval && logArchiveDest.Value != tenant.PieceSwitchInterval) {
			return true
		}
	}
	return false
}
