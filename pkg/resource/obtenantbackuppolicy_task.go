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
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apitypes "github.com/oceanbase/ob-operator/api/types"

	constants "github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

const backupVolumePath = oceanbaseconst.BackupPath

func (m *ObTenantBackupPolicyManager) ConfigureServerForBackup() error {
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
			err = con.SetLogArchiveDestForTenant(m.getArchiveDestSettingValue())
			if err != nil {
				return err
			}
		} else {
			latestArchiveJob, err := con.GetLatestArchiveLogJob()
			if err != nil {
				return err
			}
			if latestArchiveJob == nil || latestArchiveJob.Status != "DOING" {
				err = con.SetLogArchiveDestForTenant(m.getArchiveDestSettingValue())
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
	configs, err := con.ListArchiveLogParameters()
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
			err = con.SetLogArchiveConcurrency(archiveSpec.Concurrency)
			if err != nil {
				return err
			}
		}
	}
	setBackupDest := func() error {
		latestRunning, err := con.GetLatestRunningBackupJob()
		if err != nil {
			return err
		}
		if latestRunning == nil {
			err = con.SetDataBackupDestForTenant(m.getBackupDestPath())
			if err != nil {
				return err
			}
		}
		// TODO: Stop running backup job and modify the destination?
		return nil
	}
	// Maintain backup parameters
	backupConfigs, err := con.ListBackupParameters()
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

func (m *ObTenantBackupPolicyManager) StartBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	tenantInfo, err := m.getTenantRecord(true)
	if err != nil {
		return err
	}
	if tenantInfo.LogMode == "NOARCHIVELOG" {
		err = con.EnableArchiveLogForTenant()
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
		latestArchiveJob, err := con.GetLatestArchiveLogJob()
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
	err = con.CancelCleanBackup()
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

func (m *ObTenantBackupPolicyManager) CleanOldBackupJobs() error {
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

func (m *ObTenantBackupPolicyManager) PauseBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	err = con.SetLogArchiveDestState(string(constants.LogArchiveDestStateDefer))
	if err != nil {
		return err
	}

	err = con.StopBackupJobOfTenant()
	if err != nil {
		return err
	}
	err = con.CancelCleanBackup()
	if err != nil {
		return err
	}
	cleanPolicyName := "default"
	err = con.RemoveCleanBackupPolicy(cleanPolicyName)
	if err != nil {
		return err
	}
	m.Recorder.Event(m.BackupPolicy, v1.EventTypeNormal, "PauseBackup", "Pause backup policy")
	return nil
}

func (m *ObTenantBackupPolicyManager) ResumeBackup() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	err = con.SetLogArchiveDestState(string(constants.LogArchiveDestStateEnable))
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
		latestArchiveJob, err := con.GetLatestArchiveLogJob()
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

func (m *ObTenantBackupPolicyManager) syncLatestJobs() error {
	con, err := m.getOperationManager()
	if err != nil {
		return err
	}
	latestArchiveJob, err := con.GetLatestArchiveLogJob()
	if err != nil {
		return err
	}
	latestCleanJob, err := con.GetLatestBackupCleanJob()
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
	return con.GetLatestBackupJobOfType(jobType)
}

func (m *ObTenantBackupPolicyManager) getLatestBackupJobOfTypeAndPath(jobType apitypes.BackupJobType, path string) (*model.OBBackupJob, error) {
	con, err := m.getOperationManager()
	if err != nil {
		return nil, err
	}
	return con.GetLatestBackupJobOfTypeAndPath(jobType, path)
}

// get operation manager to exec sql
func (m *ObTenantBackupPolicyManager) getOperationManager() (*operation.OceanbaseOperationManager, error) {
	if m.con != nil {
		return m.con, nil
	}
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
		con, err = GetTenantRootOperationClient(m.Client, m.Logger, obcluster, m.BackupPolicy.Spec.TenantName, m.BackupPolicy.Spec.TenantSecret)
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

		con, err = GetTenantRootOperationClient(m.Client, m.Logger, obcluster, tenantCR.Spec.TenantName, tenantCR.Status.Credentials.Root)
		if err != nil {
			return nil, errors.Wrap(err, "get oceanbase operation manager")
		}
	}
	m.con = con
	return con, nil
}

func (m *ObTenantBackupPolicyManager) getArchiveDestPath() string {
	targetDest := m.BackupPolicy.Spec.LogArchive.Destination
	if targetDest.Type == constants.BackupDestTypeNFS || isZero(targetDest.Type) {
		return "file://" + path.Join(backupVolumePath, targetDest.Path)
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
	if targetDest.Type == constants.BackupDestTypeNFS || isZero(targetDest.Type) {
		return "file://" + path.Join(backupVolumePath, targetDest.Path)
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
				BlockOwnerDeletion: getRef(true),
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
	tenants, err := con.ListTenantWithName(tenantRecordName)
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
	cleanPolicy, err := con.ListBackupCleanPolicy()
	if err != nil {
		return err
	}
	policyName := "default"
	if len(cleanPolicy) == 0 {
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
	return nil
}

func (m *ObTenantBackupPolicyManager) getTenantRecordName() (string, error) {
	if m.BackupPolicy.Status.TenantCR != nil {
		return m.BackupPolicy.Status.TenantCR.Spec.TenantName, nil
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
