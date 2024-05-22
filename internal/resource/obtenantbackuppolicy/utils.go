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

package obtenantbackuppolicy

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/constants"
	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func (m *ObTenantBackupPolicyManager) syncLatestJobs() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	latestArchiveJob, err := con.GetLatestArchiveLogJob(m.Ctx)
	if err != nil {
		return err
	}
	latestCleanJob, err := con.GetLatestBackupCleanJob(m.Ctx)
	if err != nil {
		return err
	}
	m.BackupPolicy.Status.LatestArchiveLogJob = latestArchiveJob
	m.BackupPolicy.Status.LatestBackupCleanJob = latestCleanJob
	return nil
}

func (m *ObTenantBackupPolicyManager) getLatestBackupJob(jobType apitypes.BackupJobType) (*model.OBBackupJob, error) {
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	return con.GetLatestBackupJobOfType(m.Ctx, string(jobType))
}

func (m *ObTenantBackupPolicyManager) getLatestBackupJobOfTypeAndPath(jobType apitypes.BackupJobType, path string) (*model.OBBackupJob, error) {
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	return con.GetLatestBackupJobOfTypeAndPath(m.Ctx, string(jobType), path)
}

// get operation manager to exec sql
func (m *ObTenantBackupPolicyManager) getOperationManager() (*operation.OceanbaseOperationManager, error) {
	var con *operation.OceanbaseOperationManager
	var err error
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      m.BackupPolicy.Spec.ObClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	if m.BackupPolicy.Spec.TenantName != "" && m.BackupPolicy.Spec.TenantSecret != "" {
		con, err = resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, m.BackupPolicy.Spec.TenantName, m.BackupPolicy.Spec.TenantSecret)
		if err != nil {
			return nil, errors.Wrap(err, "get oceanbase operation manager")
		}
	} else if m.BackupPolicy.Spec.TenantCRName != "" {
		tenantCR := &v1alpha1.OBTenant{}
		err = m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.BackupPolicy.Namespace,
			Name:      m.BackupPolicy.Spec.TenantCRName,
		}, tenantCR)
		if err != nil {
			return nil, err
		}

		con, err = resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, tenantCR.Spec.TenantName, tenantCR.Status.Credentials.Root)
		if err != nil {
			return nil, errors.Wrap(err, "get oceanbase operation manager")
		}
	}
	return con, nil
}

func (m *ObTenantBackupPolicyManager) getArchiveDestPath() string {
	targetDest := m.BackupPolicy.Spec.LogArchive.Destination
	if targetDest.Type == constants.BackupDestTypeNFS || resourceutils.IsZero(targetDest.Type) {
		return "file://" + path.Join(oceanbaseconst.BackupPath, targetDest.Path)
	} else if targetDest.Type == constants.BackupDestTypeOSS && targetDest.OSSAccessSecret != "" {
		secret := &v1.Secret{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.BackupPolicy.GetNamespace(),
			Name:      targetDest.OSSAccessSecret,
		}, secret)
		if err != nil {
			m.PrintErrEvent(err)
			return ""
		}
		return strings.Join([]string{targetDest.Path, "access_id=" + string(secret.Data["accessId"]), "access_key=" + string(secret.Data["accessKey"])}, "&")
	}
	return targetDest.Path
}

func (m *ObTenantBackupPolicyManager) getArchiveDestSettingValue() string {
	path := m.getArchiveDestPath()
	archiveSpec := m.BackupPolicy.Spec.LogArchive
	if archiveSpec.SwitchPieceInterval != "" {
		path += fmt.Sprintf(" PIECE_SWITCH_INTERVAL=%s", archiveSpec.SwitchPieceInterval)
	}
	if archiveSpec.Binding != "" {
		path += fmt.Sprintf(" BINDING=%s", archiveSpec.Binding)
	}
	return "LOCATION=" + path
}

func (m *ObTenantBackupPolicyManager) getBackupDestPath() string {
	targetDest := m.BackupPolicy.Spec.DataBackup.Destination
	if targetDest.Type == constants.BackupDestTypeNFS || resourceutils.IsZero(targetDest.Type) {
		return "file://" + path.Join(oceanbaseconst.BackupPath, targetDest.Path)
	} else if targetDest.Type == constants.BackupDestTypeOSS && targetDest.OSSAccessSecret != "" {
		secret := &v1.Secret{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.BackupPolicy.GetNamespace(),
			Name:      targetDest.OSSAccessSecret,
		}, secret)
		if err != nil {
			m.PrintErrEvent(err)
			return ""
		}
		return strings.Join([]string{targetDest.Path, "access_id=" + string(secret.Data["accessId"]), "access_key=" + string(secret.Data["accessKey"])}, "&")
	}
	return targetDest.Path
}

func (m *ObTenantBackupPolicyManager) createBackupJob(jobType apitypes.BackupJobType) error {
	var path string
	switch jobType {
	case constants.BackupJobTypeClean:
		fallthrough
	case constants.BackupJobTypeIncr:
		fallthrough
	case constants.BackupJobTypeFull:
		path = m.getBackupDestPath()

	case constants.BackupJobTypeArchive:
		path = m.getArchiveDestPath()
	}
	var tenantRecordName string
	var tenantSecret string
	if m.BackupPolicy.Spec.TenantName != "" {
		tenantRecordName = m.BackupPolicy.Spec.TenantName
		tenantSecret = m.BackupPolicy.Spec.TenantSecret
	} else {
		tenant, err := m.getOBTenantCR()
		if err != nil {
			return err
		}
		tenantRecordName = tenant.Spec.TenantName
		tenantSecret = tenant.Status.Credentials.Root
	}

	backupJob := &v1alpha1.OBTenantBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.BackupPolicy.Name + "-" + strings.ToLower(string(jobType)) + "-" + time.Now().Format("20060102150405"),
			Namespace: m.BackupPolicy.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         m.BackupPolicy.APIVersion,
				Kind:               m.BackupPolicy.Kind,
				Name:               m.BackupPolicy.Name,
				UID:                m.BackupPolicy.GetUID(),
				BlockOwnerDeletion: resourceutils.GetRef(true),
			}},
			Labels: map[string]string{
				oceanbaseconst.LabelRefOBCluster:    m.BackupPolicy.Labels[oceanbaseconst.LabelRefOBCluster],
				oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
				oceanbaseconst.LabelRefUID:          string(m.BackupPolicy.GetUID()),
				oceanbaseconst.LabelBackupType:      string(jobType),
			},
		},
		Spec: v1alpha1.OBTenantBackupSpec{
			Path:             path,
			Type:             jobType,
			TenantName:       tenantRecordName,
			TenantSecret:     tenantSecret,
			ObClusterName:    m.BackupPolicy.Spec.ObClusterName,
			EncryptionSecret: m.BackupPolicy.Spec.DataBackup.EncryptionSecret,
		},
	}
	return m.Client.Create(m.Ctx, backupJob)
}

func (m *ObTenantBackupPolicyManager) createBackupJobIfNotExists(jobType apitypes.BackupJobType) error {
	noRunningJobs, err := m.noRunningJobs(jobType)
	if err != nil {
		m.Logger.Error(err, "Failed to check if there is running backup job")
		return nil
	}
	if noRunningJobs {
		return m.createBackupJob(jobType)
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) noRunningJobs(jobType apitypes.BackupJobType) (bool, error) {
	var runningJobs v1alpha1.OBTenantBackupList
	err := m.Client.List(m.Ctx, &runningJobs,
		client.MatchingLabels{
			oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
			oceanbaseconst.LabelBackupType:      string(jobType),
		},
		client.InNamespace(m.BackupPolicy.Namespace))
	if err != nil {
		return false, err
	}
	for _, item := range runningJobs.Items {
		if item.Spec.Type == jobType {
			switch item.Status.Status {
			case "":
				fallthrough
			case constants.BackupJobStatusInitializing:
				fallthrough
			case constants.BackupJobStatusRunning:
				return false, nil
			}
		}
	}
	return true, nil
}

// getTenantRecord return tenant info from status if exists, otherwise query from database view
func (m *ObTenantBackupPolicyManager) getTenantRecord(useCache bool) (*model.OBTenant, error) {
	if useCache && m.BackupPolicy.Status.TenantInfo != nil {
		return m.BackupPolicy.Status.TenantInfo, nil
	}
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	var tenantRecordName string
	if m.BackupPolicy.Spec.TenantName != "" {
		tenantRecordName = m.BackupPolicy.Spec.TenantName
	} else {
		tenantRecordName, err = m.getTenantRecordName()
		if err != nil {
			return nil, err
		}
	}
	tenants, err := con.ListTenantWithName(m.Ctx, tenantRecordName)
	if err != nil {
		return nil, err
	}
	if len(tenants) == 0 {
		return nil, errors.Errorf("tenant %s not found", tenantRecordName)
	}
	return tenants[0], nil
}

func (m *ObTenantBackupPolicyManager) configureBackupCleanPolicy() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	cleanPolicy, err := con.ListBackupCleanPolicy(m.Ctx)
	if err != nil {
		return err
	}
	policyName := "default"
	if len(cleanPolicy) == 0 {
		err = con.AddCleanBackupPolicy(m.Ctx, policyName, cleanConfig.RecoveryWindow)
		if err != nil {
			return err
		}
	} else {
		for _, policy := range cleanPolicy {
			if policy.RecoveryWindow != cleanConfig.RecoveryWindow {
				err = con.RemoveCleanBackupPolicy(m.Ctx, policy.PolicyName)
				if err != nil {
					return err
				}
				err = con.AddCleanBackupPolicy(m.Ctx, policyName, cleanConfig.RecoveryWindow)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) getTenantRecordName() (string, error) {
	if m.BackupPolicy.Status.TenantCR != nil {
		return m.BackupPolicy.Status.TenantCR.Spec.TenantName, nil
	}
	if m.BackupPolicy.Status.TenantName != "" {
		return m.BackupPolicy.Status.TenantName, nil
	}
	tenant := &v1alpha1.OBTenant{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.BackupPolicy.Namespace,
		Name:      m.BackupPolicy.Spec.TenantCRName,
	}, tenant)
	if err != nil {
		return "", err
	}
	return tenant.Spec.TenantName, nil
}
