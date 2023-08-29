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
	"time"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	tenant, err := m.getTenantInfo()
	if err != nil {
		return err
	}
	m.BackupPolicy.Status.TenantInfo = tenant
	// update status ahead of regular task
	return m.Client.Status().Update(m.Ctx, m.BackupPolicy)
}

func (m *ObTenantBackupPolicyManager) StartBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantInfo()
	if err != nil {
		return err
	}
	if tenantInfo.LogMode == "NOARCHIVELOG" {
		err = con.EnableArchiveLogForTenant()
		if err != nil {
			return err
		}
	}
	cleanConfig := &m.BackupPolicy.Spec.DataClean
	cleanPolicy, err := con.QueryBackupCleanPolicy()
	if err != nil {
		return err
	}
	policyName := "default"
	if len(cleanPolicy) == 0 {
		// the name of the policy can only be 'default', and the recovery window can only be 1d-7d
		err = con.AddCleanBackupPolicy(policyName, cleanConfig.RecoveryWindow)
		if err != nil {
			return err
		}
	} else {
		for _, policy := range cleanPolicy {
			if policy.RecoveryWindow != cleanConfig.RecoveryWindow {
				err = con.RemoveCleanBackupPolicy(policy.PolicyName)
				if err != nil {
					return err
				}
				err = con.AddCleanBackupPolicy(policyName, cleanConfig.RecoveryWindow)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	// create backup job of full type
	return m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeFull)
}

func (m *ObTenantBackupPolicyManager) StopBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	// ignore the error
	err = con.DisableArchiveLogForTenant()
	if err != nil {
		return err
	}
	err = con.StopBackupJobOfTenant()
	if err != nil {
		return err
	}
	cleanPolicyName := "default"
	err = con.RemoveCleanBackupPolicy(cleanPolicyName)
	if err != nil {
		return err
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) CheckAndSpawnJobs() error {
	latestFull, err := m.getLatestBackupJob(v1alpha1.BackupJobTypeFull)
	if err != nil {
		return err
	}
	if latestFull == nil {
		return m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeFull)
	}
	if latestFull.Status == "COMPLETED" {
		var lastFullBackupFinishedAt time.Time
		if latestFull.EndTimestamp == nil {
			// TODO: check if this is possible: COMPLETED job with nil end timestamp
			lastFullBackupFinishedAt, err = time.Parse(time.DateTime, latestFull.StartTimestamp)
		} else {
			lastFullBackupFinishedAt, err = time.Parse(time.DateTime, *latestFull.EndTimestamp)
		}
		if err != nil {
			m.Logger.Error(err, "Failed to parse end timestamp of completed backup job")
		}
		timeNow := time.Now()
		fullCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.FullCrontab)
		if err != nil {
			// TODO: Check pattern of crontab with admission webhook
			m.Logger.Error(err, "Failed to parse full backup crontab")
			return nil
		}
		nextFullTime := fullCron.Next(lastFullBackupFinishedAt)
		if nextFullTime.Before(timeNow) {
			// Full backup and incremental backup can not be executed at the same time
			// create Full type backup job and return here
			return m.createBackupJob(v1alpha1.BackupJobTypeFull)
		}

		// considering incremental backup
		// create incremental backup if there is a completed full/incremental backup job
		incrementalCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.IncrCrontab)
		if err != nil {
			m.Logger.Error(err, "Failed to parse full backup crontab")
			return nil
		}
		latestIncr, err := m.getLatestBackupJob(v1alpha1.BackupJobTypeIncr)
		if err != nil {
			return err
		}
		if latestIncr != nil {
			if latestIncr.Status == "COMPLETED" {
				var lastIncrBackupFinishedAt time.Time
				if latestIncr.EndTimestamp == nil {
					// TODO: check if this is possible
					lastIncrBackupFinishedAt, err = time.Parse(time.DateTime, latestIncr.StartTimestamp)
				} else {
					lastIncrBackupFinishedAt, err = time.Parse(time.DateTime, *latestIncr.EndTimestamp)
				}
				if err != nil {
					m.Logger.Error(err, "Failed to parse end timestamp of completed backup job")
				}

				nextIncrTime := incrementalCron.Next(lastIncrBackupFinishedAt)
				if nextIncrTime.Before(timeNow) {
					err = m.createBackupJob(v1alpha1.BackupJobTypeIncr)
					if err != nil {
						return err
					}
				}
			} else if latestIncr.Status == "INIT" || latestIncr.Status == "DOING" {
				// do nothing
			} else {
				m.Logger.Info("Incremental BackupJob are in status " + latestIncr.Status)
			}
		} else {
			nextIncrTime := incrementalCron.Next(lastFullBackupFinishedAt)
			if nextIncrTime.Before(timeNow) {
				err = m.createBackupJob(v1alpha1.BackupJobTypeIncr)
				if err != nil {
					return err
				}
			}
		}
	} else if latestFull.Status == "INIT" || latestFull.Status == "DOING" {
		// do nothing
	} else {
		m.Logger.Info("BackupJob are in status " + latestFull.Status)
	}
	return nil
}

func (m *ObTenantBackupPolicyManager) getLatestBackupJob(jobType v1alpha1.BackupJobType) (*model.OBBackupJob, error) {
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	jobs, err := con.QueryLatestBackupJob(jobType)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	return jobs[0], nil
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

func (m *ObTenantBackupPolicyManager) createBackupJob(jobType v1alpha1.BackupJobType) error {
	backupJob := &v1alpha1.OBTenantBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.BackupPolicy.Name + "-" + string(jobType) + "-" + time.Now().Format("20060102150405"),
			Namespace: m.BackupPolicy.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         m.BackupPolicy.APIVersion,
				Kind:               m.BackupPolicy.Kind,
				Name:               m.BackupPolicy.Name,
				UID:                m.BackupPolicy.GetUID(),
				BlockOwnerDeletion: getRef(true),
			}},
			Labels: map[string]string{
				oceanbaseconst.LabelRefOBCluster:    m.BackupPolicy.Labels[oceanbaseconst.LabelRefOBCluster],
				oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
				oceanbaseconst.LabelRefUID:          string(m.BackupPolicy.GetUID()),
			},
		},
		Spec: v1alpha1.OBTenantBackupSpec{
			Type:       jobType,
			TenantName: m.BackupPolicy.Spec.TenantName,
		},
	}
	return m.Client.Create(m.Ctx, backupJob)
}

func (m *ObTenantBackupPolicyManager) createBackupJobIfNotExists(jobType v1alpha1.BackupJobType) error {
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

func (m *ObTenantBackupPolicyManager) noRunningJobs(jobType v1alpha1.BackupJobType) (bool, error) {
	var runningJobs v1alpha1.OBTenantBackupList
	err := m.Client.List(m.Ctx, &runningJobs,
		client.MatchingLabels{
			oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
		},
		client.MatchingFieldsSelector{
			Selector: fields.AndSelectors(
				fields.OneTermEqualSelector("spec.tenantName", m.BackupPolicy.Spec.TenantName),
				fields.OneTermEqualSelector("spec.type", string(jobType)),
				fields.OneTermNotEqualSelector("status.status", string(v1alpha1.BackupJobStatusFailed)),
				fields.OneTermNotEqualSelector("status.status", string(v1alpha1.BackupJobStatusSuccessful)),
			),
		},
		client.InNamespace(m.BackupPolicy.Namespace))
	if err != nil {
		return false, err
	}
	return len(runningJobs.Items) == 0, nil
}

// getTenantInfo return tenant info from status if exists, otherwise query from database view
func (m *ObTenantBackupPolicyManager) getTenantInfo() (*model.OBTenant, error) {
	if m.BackupPolicy.Status.TenantInfo != nil {
		return m.BackupPolicy.Status.TenantInfo, nil
	}
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	tenants, err := con.QueryTenantWithName(m.BackupPolicy.Spec.TenantName)
	if err != nil {
		return nil, err
	}
	if len(tenants) == 0 {
		return nil, errors.Errorf("tenant %s not found", m.BackupPolicy.Spec.TenantName)
	}
	return tenants[0], nil
}
