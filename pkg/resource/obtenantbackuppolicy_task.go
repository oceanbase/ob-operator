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

package resource

import (
	"fmt"
	"path"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const backupVolumePath = oceanbaseconst.BackupPath

func (m *ObTenantBackupPolicyManager) ConfigureServerForBackup() error {
	m.Logger.Info("Start configuring server for backup")
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantName := m.BackupPolicy.Spec.TenantName
	err = con.SetLogArchiveDestForTenant(tenantName, m.getArchiveDestPath())
	if err != nil {
		return err
	}
	err = con.SetDataBackupDestForTenant(tenantName, m.getBackupDestPath())
	if err != nil {
		return err
	}
	if m.BackupPolicy.Spec.LogArchive.Concurrency != 0 {
		err = con.SetLogArchiveConcurrency(tenantName, m.BackupPolicy.Spec.LogArchive.Concurrency)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) GetTenantInfo() error {
	m.Logger.Info("Start getting tenant info")
	// Admission Control
	con, err := m.getOperationManager()
	if err != nil {
		m.Logger.Error(err, "Failed to get operator manager")
		return err
	}
	tenants, err := con.QueryTenantWithName(m.BackupPolicy.Spec.TenantName)
	if err != nil {
		m.Logger.Error(err, "Failed to query tenant with name")
		return err
	}
	if len(tenants) == 0 {
		// never happen by design
		return errors.Errorf("tenant %s not found", m.BackupPolicy.Spec.TenantName)
	}
	m.Logger.Info("get tenant info:", "tenants", tenants[0])
	m.BackupPolicy.Status.TenantInfo = tenants[0]
	return nil
}

func (m *ObTenantBackupPolicyManager) StartBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantName := m.BackupPolicy.Spec.TenantName
	err = con.EnableArchiveLogForTenant(tenantName)
	if err != nil {
		return err
	}
	err = con.CreateBackupFull(tenantName)
	// if m.BackupPolicy.Spec.DataBackup.Type == v1alpha1.BackupFull {
	// } else {
	// 	err = con.CreateBackupIncr(tenantName)
	// }
	if err != nil {
		return err
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	err = con.AddCleanBackupPolicy(cleanConfig.Name, cleanConfig.RecoverWindow, m.BackupPolicy.Spec.TenantName)
	if err != nil {
		return err
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) StopBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantName := m.BackupPolicy.Spec.TenantName
	err = con.DisableArchiveLogForTenant(tenantName)
	if err != nil {
		return err
	}
	err = con.StopBackupJobOfTenant(tenantName)
	if err != nil {
		return err
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	err = con.RemoveCleanBackupPolicy(cleanConfig.Name, m.BackupPolicy.Spec.TenantName)
	if err != nil {
		return err
	}
	return nil
}

// get operation manager to exec sql
func (m *ObTenantBackupPolicyManager) getOperationManager() (*operation.OceanbaseOperationManager, error) {
	if m.con != nil {
		return m.con, nil
	}
	clusterName, _ := m.BackupPolicy.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := GetOceanbaseOperationManagerFromOBCluster(m.Client, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	m.con = con
	return con, nil
}

func (m *ObTenantBackupPolicyManager) getArchiveDestPath() string {
	dest := path.Join("file://", backupVolumePath, m.BackupPolicy.Spec.TenantName, "log_archive")
	if m.BackupPolicy.Spec.LogArchive.SwitchPieceInterval != "" {
		dest += fmt.Sprintf(" PIECE_SWITCH_INTERVAL=%s", m.BackupPolicy.Spec.LogArchive.SwitchPieceInterval)
	}
	return dest
}

func (m *ObTenantBackupPolicyManager) getBackupDestPath() string {
	return path.Join("file://", backupVolumePath, m.BackupPolicy.Spec.TenantName, "data_backup")
}
