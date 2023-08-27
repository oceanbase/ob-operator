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
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	if m.BackupPolicy.Status.TenantInfo != nil &&
		m.BackupPolicy.Status.TenantInfo.LogMode == "NOARCHIVELOG" {
		err = con.SetLogArchiveDestForTenant(m.getArchiveDestPath())
		if err != nil {
			return err
		}
	}
	if m.BackupPolicy.Spec.LogArchive.Concurrency != 0 {
		err = con.SetLogArchiveConcurrency(m.BackupPolicy.Spec.LogArchive.Concurrency)
		if err != nil {
			return err
		}
	}
	err = con.SetDataBackupDestForTenant(m.getBackupDestPath())
	if err != nil {
		return err
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) GetTenantInfo() error {
	// Admission Control
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenants, err := con.QueryTenantWithName(m.BackupPolicy.Spec.TenantName)
	if err != nil {
		return err
	}
	if len(tenants) == 0 {
		// never happen by design
		return errors.Errorf("tenant %s not found", m.BackupPolicy.Spec.TenantName)
	}
	m.BackupPolicy.Status.TenantInfo = tenants[0]
	m.Logger.Info("get tenant info", "info", m.BackupPolicy.Status.TenantInfo)
	return nil
}

func (m *ObTenantBackupPolicyManager) StartBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	if m.BackupPolicy.Status.TenantInfo != nil &&
		m.BackupPolicy.Status.TenantInfo.LogMode == "NOARCHIVELOG" {
		err = con.EnableArchiveLogForTenant()
		if err != nil {
			return err
		}
	}
	// create backup job of full type
	// TODO: create backup job to manage the tasks
	err = con.CreateBackupFull()
	if err != nil {
		return err
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	cleanPolicy, err := con.QueryBackupCleanPolicy()
	if err != nil {
		return err
	}
	policyName := "default"
	if len(cleanPolicy) == 0 {
		// the name of the policy can only be 'default', and the recovery window can only be 1d-7d
		err = con.AddCleanBackupPolicy(policyName, cleanConfig.RecoverWindow)
		if err != nil {
			return err
		}
	} else {
		for _, policy := range cleanPolicy {
			if policy.RecoverWindow != cleanConfig.RecoverWindow {
				err = con.RemoveCleanBackupPolicy(policy.PolicyName)
				if err != nil {
					return err
				}
				err = con.AddCleanBackupPolicy(policyName, cleanConfig.RecoverWindow)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) StopBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	err = con.DisableArchiveLogForTenant()
	if err != nil {
		return err
	}
	err = con.StopBackupJobOfTenant()
	if err != nil {
		return err
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	err = con.RemoveCleanBackupPolicy(cleanConfig.Name)
	if err != nil {
		return err
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) SpawnCronBackupJob() error {
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
	con, err := GetTenantOperationClient(m.Client, m.Logger, obcluster, m.BackupPolicy.Spec.TenantName)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	m.con = con
	return con, nil
}

func (m *ObTenantBackupPolicyManager) getArchiveDestPath() string {
	targetDest := m.BackupPolicy.Spec.LogArchive.Destination
	if targetDest.Type == v1alpha1.BackupDestTypeNFS {
		var dest string
		if targetDest.Path == "" {
			dest = "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, "log_archive")
		} else {
			dest = "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, targetDest.Path)
		}
		if m.BackupPolicy.Spec.LogArchive.SwitchPieceInterval != "" {
			dest += fmt.Sprintf(" PIECE_SWITCH_INTERVAL=%s", m.BackupPolicy.Spec.LogArchive.SwitchPieceInterval)
		}
		return "location=" + dest
	} else {
		return targetDest.Path
	}
}

func (m *ObTenantBackupPolicyManager) getBackupDestPath() string {
	targetDest := m.BackupPolicy.Spec.DataBackup.Destination
	if targetDest.Type == v1alpha1.BackupDestTypeNFS {
		if targetDest.Path == "" {
			return "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, "data_backup")
		} else {
			return "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, targetDest.Path)
		}
	} else {
		return targetDest.Path
	}
}
