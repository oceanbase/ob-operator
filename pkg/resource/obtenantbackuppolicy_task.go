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
	"strings"
	"time"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const backupVolumePath = oceanbaseconst.BackupPath

func (m *ObTenantBackupPolicyManager) ConfigureServerForBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantInfo()
	if err != nil {
		return err
	}

	// Maintain log archive parameters
	configs, err := con.QueryArchiveLogParameters()
	if err != nil {
		return err
	}
	archiveSpec := m.BackupPolicy.Spec.LogArchive
	archivePath := m.getArchiveDestPath()
	for _, config := range configs {
		switch {
		case config.Name == "path" && config.Value != archivePath:
			fallthrough
		case config.Name == "piece_switch_interval" && config.Value != archiveSpec.SwitchPieceInterval:
			fallthrough
		case config.Name == "binding" && config.Value != strings.ToUpper(string(archiveSpec.Binding)):
			if tenantInfo.LogMode == "NOARCHIVELOG" {
				err = con.SetLogArchiveDestForTenant(m.getArchiveDestSettingValue())
				if err != nil {
					return err
				}
			} else {
				latestArchiveJob, err := con.QueryLatestArchiveLogJob()
				if err != nil {
					return err
				}
				if latestArchiveJob == nil || latestArchiveJob.Status != "DOING" {
					err = con.SetLogArchiveDestForTenant(m.getArchiveDestSettingValue())
					if err != nil {
						return err
					}
				} else {
					// TODO: Stop running log archive job and modify destination?
					// Log archive jobs won't terminate if no error happens
				}
			}
		default:
			// configurations match, do nothing
		}
	}
	if archiveSpec.Concurrency != 0 {
		err = con.SetLogArchiveConcurrency(archiveSpec.Concurrency)
		if err != nil {
			return err
		}
	}

	// Maintain backup parameters
	backupConfigs, err := con.QueryBackupParameters()
	if err != nil {
		return err
	}
	backupPath := m.getBackupDestPath()
	for _, config := range backupConfigs {
		// Can not modify backup destination when there is a running backup job
		if config.Name == "data_backup_dest" && config.Value != backupPath {
			latestRunning, err := con.QueryLatestRunningBackupJob()
			if err != nil {
				return err
			}
			if latestRunning == nil {
				err = con.SetDataBackupDestForTenant(m.getBackupDestPath())
				if err != nil {
					return err
				}
			} else {
				// TODO: Stop running backup job and modify the destination?
			}
		}
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
	return nil
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
	err = m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeArchive)
	if err != nil {
		return err
	}
	err = m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeClean)
	if err != nil {
		return err
	}
	// create backup job of full type
	return m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeFull)
}

func (m *ObTenantBackupPolicyManager) StopBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}

	_ = con.DisableArchiveLogForTenant()

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
	backupPath := m.getBackupDestPath()
	// Avoid backup failure due to destination modification
	latestFull, err := m.getLatestBackupJobOfTypeAndPath(v1alpha1.BackupJobTypeFull, backupPath)
	if err != nil {
		return err
	}
	if latestFull == nil || latestFull.Status == "CANCELED" {
		return m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeFull)
	}
	if latestFull.Status == "COMPLETED" {
		var lastFullBackupFinishedAt time.Time
		if latestFull.EndTimestamp == nil {
			// TODO: check if this is possible: COMPLETED job with nil end timestamp
			lastFullBackupFinishedAt, err = time.ParseInLocation(time.DateTime, latestFull.StartTimestamp, time.Local)
		} else {
			lastFullBackupFinishedAt, err = time.ParseInLocation(time.DateTime, *latestFull.EndTimestamp, time.Local)
		}
		if err != nil {
			m.Logger.Error(err, "Failed to parse end timestamp of completed backup job")
			return nil
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
			return m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeFull)
		} else {
			m.BackupPolicy.Status.NextFullBackupAt = metav1.NewTime(nextFullTime)
		}

		// considering incremental backup
		// create incremental backup if there is a completed full/incremental backup job
		incrementalCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.IncrementalCrontab)
		if err != nil {
			m.Logger.Error(err, "Failed to parse full backup crontab")
			return nil
		}
		latestIncr, err := m.getLatestBackupJobOfTypeAndPath(v1alpha1.BackupJobTypeIncr, backupPath)
		if err != nil {
			return err
		}
		if latestIncr != nil {
			if latestIncr.Status == "COMPLETED" || latestIncr.Status == "CANCELED" {
				var lastIncrBackupFinishedAt time.Time
				if latestIncr.EndTimestamp == nil {
					// TODO: check if this is possible
					lastIncrBackupFinishedAt, err = time.ParseInLocation(time.DateTime, latestIncr.StartTimestamp, time.Local)
				} else {
					lastIncrBackupFinishedAt, err = time.ParseInLocation(time.DateTime, *latestIncr.EndTimestamp, time.Local)
				}
				if err != nil {
					m.Logger.Error(err, "Failed to parse end timestamp of completed backup job")
				}

				nextIncrTime := incrementalCron.Next(lastIncrBackupFinishedAt)
				if nextIncrTime.Before(timeNow) {
					err = m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeIncr)
					if err != nil {
						return err
					}
				} else {
					m.BackupPolicy.Status.NextIncrBackupAt = metav1.NewTime(nextIncrTime)
				}
			} else if latestIncr.Status == "INIT" || latestIncr.Status == "DOING" {
				// do nothing
			} else {
				m.Logger.Info("Incremental BackupJob are in status " + latestIncr.Status)
			}
		} else {
			nextIncrTime := incrementalCron.Next(lastFullBackupFinishedAt)
			if nextIncrTime.Before(timeNow) {
				err = m.createBackupJobIfNotExists(v1alpha1.BackupJobTypeIncr)
				if err != nil {
					return err
				}
			} else {
				m.BackupPolicy.Status.NextIncrBackupAt = metav1.NewTime(nextIncrTime)
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
	return con.QueryLatestBackupJobOfType(jobType)
}

func (m *ObTenantBackupPolicyManager) getLatestBackupJobOfTypeAndPath(jobType v1alpha1.BackupJobType, path string) (*model.OBBackupJob, error) {
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	return con.QueryLatestBackupJobOfTypeAndPath(jobType, path)
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
	archiveSpec := m.BackupPolicy.Spec.LogArchive
	targetDest := archiveSpec.Destination
	if targetDest.Type == v1alpha1.BackupDestTypeNFS || isZero(targetDest.Type) {
		var dest string
		if targetDest.Path == "" {
			dest = "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, "log_archive")
		} else {
			dest = "file://" + path.Join(backupVolumePath, m.BackupPolicy.Spec.TenantName, targetDest.Path)
		}
		return dest

	} else {
		return targetDest.Path
	}
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
	if targetDest.Type == v1alpha1.BackupDestTypeNFS || isZero(targetDest.Type) {
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
	m.Logger.Info("Create Backup Job", "type", jobType)
	var path string
	switch jobType {
	case v1alpha1.BackupJobTypeClean:
		fallthrough
	case v1alpha1.BackupJobTypeIncr:
		fallthrough
	case v1alpha1.BackupJobTypeFull:
		path = m.getBackupDestPath()

	case v1alpha1.BackupJobTypeArchive:
		path = m.getArchiveDestPath()
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
				BlockOwnerDeletion: getRef(true),
			}},
			Labels: map[string]string{
				oceanbaseconst.LabelRefOBCluster:    m.BackupPolicy.Labels[oceanbaseconst.LabelRefOBCluster],
				oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
				oceanbaseconst.LabelRefUID:          string(m.BackupPolicy.GetUID()),
				oceanbaseconst.LabelTenantName:      m.BackupPolicy.Spec.TenantName,
				oceanbaseconst.LabelBackupType:      string(jobType),
			},
		},
		Spec: v1alpha1.OBTenantBackupSpec{
			Path:       path,
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
	m.Logger.Info("runningJobs?", "type", noRunningJobs)
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
			oceanbaseconst.LabelTenantName:      m.BackupPolicy.Spec.TenantName,
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
			case v1alpha1.BackupJobStatusInitializing:
				fallthrough
			case v1alpha1.BackupJobStatusRunning:
				return false, nil
			}
		}
	}
	return true, nil
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
