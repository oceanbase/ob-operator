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
	"strconv"
	"strings"
	"time"

	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	constants "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*ObTenantBackupPolicyManager]()

func ConfigureServerForBackup(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Configure Server For Backup")
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantRecord(true)
	if err != nil {
		return err
	}
	setArchiveDest := func() error {
		if tenantInfo.LogMode == "NOARCHIVELOG" {
			err = con.SetLogArchiveDestForTenant(m.Ctx, m.getArchiveDestSettingValue())
			if err != nil {
				return err
			}
		} else {
			latestArchiveJob, err := con.GetLatestArchiveLogJob(m.Ctx)
			if err != nil {
				return err
			}
			if latestArchiveJob == nil || latestArchiveJob.Status != "DOING" {
				err = con.SetLogArchiveDestForTenant(m.Ctx, m.getArchiveDestSettingValue())
				if err != nil {
					return err
				}
			}
			// TODO: Stop running log archive job and modify destination?
			// Log archive jobs won't terminate if no error happens
		}
		return nil
	}
	// Maintain log archive parameters
	configs, err := con.ListArchiveLogParameters(m.Ctx)
	if err != nil {
		return err
	}
	if len(configs) == 0 {
		err = setArchiveDest()
		if err != nil {
			return err
		}
	} else {
		archiveSpec := m.BackupPolicy.Spec.LogArchive
		archivePath := m.getArchiveDestPath()
		for _, config := range configs {
			switch {
			case config.Name == "path" && config.Value != archivePath:
				fallthrough
			case config.Name == "piece_switch_interval" && config.Value != archiveSpec.SwitchPieceInterval:
				fallthrough
			case config.Name == "binding" && config.Value != strings.ToUpper(string(archiveSpec.Binding)):
				err = setArchiveDest()
				if err != nil {
					return err
				}
			default:
				// configurations match, do nothing
			}
		}
		if archiveSpec.Concurrency != 0 {
			err = con.SetLogArchiveConcurrency(m.Ctx, archiveSpec.Concurrency)
			if err != nil {
				return err
			}
		}
	}
	setBackupDest := func() error {
		latestRunning, err := con.GetLatestRunningBackupJob(m.Ctx)
		if err != nil {
			return err
		}
		if latestRunning == nil {
			err = con.SetDataBackupDestForTenant(m.Ctx, m.getBackupDestPath())
			if err != nil {
				return err
			}
		}
		// TODO: Stop running backup job and modify the destination?
		return nil
	}
	// Maintain backup parameters
	backupConfigs, err := con.ListBackupParameters(m.Ctx)
	if err != nil {
		return err
	}
	backupPath := m.getBackupDestPath()
	if len(backupConfigs) == 0 {
		err = setBackupDest()
		if err != nil {
			return err
		}
	} else {
		for _, config := range backupConfigs {
			// Can not modify backup destination when there is a running backup job
			if config.Name == "data_backup_dest" && config.Value != backupPath {
				err = setBackupDest()
				if err != nil {
					return err
				}
			}
		}
	}
	err = m.configureBackupCleanPolicy()
	if err != nil {
		return err
	}
	return nil
}

func StartBackup(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantRecord(true)
	if err != nil {
		return err
	}
	if tenantInfo.LogMode == "NOARCHIVELOG" {
		err = con.EnableArchiveLogForTenant(m.Ctx)
		if err != nil {
			return err
		}
	}
	err = m.createBackupJobIfNotExists(constants.BackupJobTypeArchive)
	if err != nil {
		return err
	}
	err = m.createBackupJobIfNotExists(constants.BackupJobTypeClean)
	if err != nil {
		return err
	}
	// Initialization: wait for archive log job to start
	archiveRunning := false
	for !archiveRunning {
		time.Sleep(10 * time.Second)
		latestArchiveJob, err := con.GetLatestArchiveLogJob(m.Ctx)
		if err != nil {
			return err
		}
		if latestArchiveJob != nil && latestArchiveJob.Status == "DOING" {
			archiveRunning = true
		}
	}
	// create backup job of full type
	return m.createBackupJobIfNotExists(constants.BackupJobTypeFull)
}

func StopBackup(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantRecord(false)
	if err != nil {
		return err
	}
	if tenantInfo.LogMode != "NOARCHIVELOG" {
		err = con.DisableArchiveLogForTenant(m.Ctx)
		if err != nil {
			return err
		}
	}

	err = con.StopBackupJobOfTenant(m.Ctx)
	if err != nil {
		return err
	}
	err = con.CancelCleanBackup(m.Ctx)
	if err != nil {
		return err
	}
	cleanPolicyName := "default"
	err = con.RemoveCleanBackupPolicy(m.Ctx, cleanPolicyName)
	if err != nil {
		return err
	}
	return nil
}

func CheckAndSpawnJobs(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	var backupPath string
	if m.BackupPolicy.Spec.DataBackup.Destination.Type == constants.BackupDestTypeOSS {
		backupPath = m.BackupPolicy.Spec.DataBackup.Destination.Path
	} else {
		backupPath = m.getBackupDestPath()
	}
	// Avoid backup failure due to destination modification
	latestFull, err := m.getLatestBackupJobOfTypeAndPath(constants.BackupJobTypeFull, backupPath)
	if err != nil {
		return err
	}
	if latestFull == nil || latestFull.Status == "CANCELED" {
		return m.createBackupJobIfNotExists(constants.BackupJobTypeFull)
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
			m.Logger.Error(err, "Failed to parse full backup crontab")
			return nil
		}
		nextFullTime := fullCron.Next(lastFullBackupFinishedAt)
		if nextFullTime.Before(timeNow) {
			// Full backup and incremental backup can not be executed at the same time
			// create Full type backup job and return here
			return m.createBackupJobIfNotExists(constants.BackupJobTypeFull)
		}

		// considering incremental backup
		// create incremental backup if there is a completed full/incremental backup job
		incrementalCron, err := cron.ParseStandard(m.BackupPolicy.Spec.DataBackup.IncrementalCrontab)
		if err != nil {
			m.Logger.Error(err, "Failed to parse full backup crontab")
			return nil
		}
		latestIncr, err := m.getLatestBackupJobOfTypeAndPath(constants.BackupJobTypeIncr, backupPath)
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
					err = m.createBackupJobIfNotExists(constants.BackupJobTypeIncr)
					if err != nil {
						return err
					}
				}
			} else if latestIncr.Status == "INIT" || latestIncr.Status == "DOING" {
				// do nothing
				_ = latestIncr
			} else {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Incremental BackupJob are in status " + latestIncr.Status)
			}
		} else {
			nextIncrTime := incrementalCron.Next(lastFullBackupFinishedAt)
			if nextIncrTime.Before(timeNow) {
				err = m.createBackupJobIfNotExists(constants.BackupJobTypeIncr)
				if err != nil {
					return err
				}
			}
		}
	} else if latestFull.Status == "INIT" || latestFull.Status == "DOING" {
		// do nothing
		_ = latestFull
	} else {
		m.Logger.V(oceanbaseconst.LogLevelDebug).Info("BackupJob are in status " + latestFull.Status)
	}
	return nil
}

func CleanOldBackupJobs(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	// JobKeepWindow is not set, do nothing
	if m.BackupPolicy.Spec.JobKeepWindow == "" {
		return nil
	}
	var jobs v1alpha1.OBTenantBackupList
	fieldSelector := fields.ParseSelectorOrDie(".status.status=" + string(constants.BackupJobStatusSuccessful))
	labelSelector, err := labels.Parse(fmt.Sprintf("%s in (%s, %s)", oceanbaseconst.LabelBackupType, constants.BackupJobTypeFull, constants.BackupJobTypeIncr))
	if err != nil {
		return err
	}
	err = m.Client.List(m.Ctx, &jobs,
		client.MatchingLabels{
			oceanbaseconst.LabelRefBackupPolicy: m.BackupPolicy.Name,
		},
		client.MatchingLabelsSelector{
			Selector: labelSelector,
		},
		client.MatchingFieldsSelector{
			Selector: fieldSelector,
		},
		client.InNamespace(m.BackupPolicy.Namespace))
	if err != nil {
		return err
	}
	if len(jobs.Items) == 0 {
		return nil
	}
	keepWindowDays, err := strconv.Atoi(strings.TrimRight(m.BackupPolicy.Spec.JobKeepWindow, "d"))
	if err != nil {
		return err
	}
	keepWindowDuration := time.Duration(keepWindowDays*24) * time.Hour
	for i, job := range jobs.Items {
		if job.Status.BackupJob.EndTimestamp != nil {
			finishedAt, err := time.ParseInLocation(time.DateTime, *job.Status.BackupJob.EndTimestamp, time.Local)
			if err != nil {
				return err
			}
			if finishedAt.Add(keepWindowDuration).Before(time.Now()) {
				err = m.Client.Delete(m.Ctx, &jobs.Items[i])
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func PauseBackup(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	err = con.SetLogArchiveDestState(m.Ctx, string(constants.LogArchiveDestStateDefer))
	if err != nil {
		return err
	}

	err = con.StopBackupJobOfTenant(m.Ctx)
	if err != nil {
		return err
	}
	err = con.CancelCleanBackup(m.Ctx)
	if err != nil {
		return err
	}
	cleanPolicyName := "default"
	err = con.RemoveCleanBackupPolicy(m.Ctx, cleanPolicyName)
	if err != nil {
		return err
	}
	m.Recorder.Event(m.BackupPolicy, v1.EventTypeNormal, "PauseBackup", "Pause backup policy")
	return nil
}

func ResumeBackup(m *ObTenantBackupPolicyManager) tasktypes.TaskError {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	err = con.SetLogArchiveDestState(m.Ctx, string(constants.LogArchiveDestStateEnable))
	if err != nil {
		return err
	}
	err = m.configureBackupCleanPolicy()
	if err != nil {
		m.Logger.Info("Failed to configure backup clean policy", "error", err)
	}
	archiveRunning := false
	for !archiveRunning {
		time.Sleep(10 * time.Second)
		latestArchiveJob, err := con.GetLatestArchiveLogJob(m.Ctx)
		if err != nil {
			return err
		}
		if latestArchiveJob != nil && latestArchiveJob.Status == "DOING" {
			archiveRunning = true
		}
	}
	m.Recorder.Event(m.BackupPolicy, v1.EventTypeNormal, "ResumeBackup", "Resume backup policy")
	return m.createBackupJobIfNotExists(constants.BackupJobTypeFull)
}
