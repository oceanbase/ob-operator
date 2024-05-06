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

package obtenantbackup

import (
	"fmt"

	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (m *OBTenantBackupManager) getObOperationClient() (*operation.OceanbaseOperationManager, error) {
	var err error
	job := m.Resource
	obcluster := &v1alpha1.OBCluster{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      job.Spec.ObClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := resourceutils.GetTenantRootOperationClient(m.Client, m.Logger, obcluster, job.Spec.TenantName, job.Spec.TenantSecret)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (m *OBTenantBackupManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newestJob := &v1alpha1.OBTenantBackup{}
		err := m.Client.Get(m.Ctx, types.NamespacedName{
			Namespace: m.Resource.GetNamespace(),
			Name:      m.Resource.GetName(),
		}, newestJob)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		newestJob.Status = m.Resource.Status
		return m.Client.Status().Update(m.Ctx, newestJob)
	})
}

func (m *OBTenantBackupManager) maintainRunningBackupJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "Failed to get ob operation client")
		return err
	}
	var targetJob *model.OBBackupJob
	if job.Status.BackupJob == nil {
		// occasionally happen, try to fetch the job from OB view
		if job.Spec.Type == constants.BackupJobTypeFull || job.Spec.Type == constants.BackupJobTypeIncr {
			latest, err := con.GetLatestBackupJobOfType(job.Spec.Type)
			if err != nil {
				return err
			}
			job.Status.BackupJob = latest
			targetJob = latest
		}
		// archive log and data clean job should not be here
	} else {
		modelJob, err := con.GetBackupJobWithId(job.Status.BackupJob.JobID)
		if err != nil {
			return err
		}
		if modelJob == nil {
			return fmt.Errorf("backup job with id %d not found", job.Status.BackupJob.JobID)
		}
		job.Status.BackupJob = modelJob
		targetJob = modelJob
	}
	job.Status.StartedAt = targetJob.StartTimestamp
	if targetJob.EndTimestamp != nil {
		job.Status.EndedAt = *targetJob.EndTimestamp
	}
	switch targetJob.Status {
	case "COMPLETED":
		job.Status.Status = constants.BackupJobStatusSuccessful
	case "FAILED":
		job.Status.Status = constants.BackupJobStatusFailed
	case "CANCELED":
		job.Status.Status = constants.BackupJobStatusCanceled
	}
	return nil
}

func (m *OBTenantBackupManager) maintainRunningBackupCleanJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "Failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestBackupCleanJob()
	if err != nil {
		logger.Error(err, "Failed to query latest backup clean job")
		return err
	}
	if latest != nil {
		job.Status.DataCleanJob = latest
		job.Status.StartedAt = latest.StartTimestamp
		if latest.EndTimestamp != nil {
			job.Status.EndedAt = *latest.EndTimestamp
		}
		switch latest.Status {
		case "COMPLETED":
			job.Status.Status = constants.BackupJobStatusSuccessful
		case "FAILED":
			job.Status.Status = constants.BackupJobStatusFailed
		case "CANCELED":
			job.Status.Status = constants.BackupJobStatusCanceled
		case "DOING":
			job.Status.Status = constants.BackupJobStatusRunning
		}
	}
	return nil
}

func (m *OBTenantBackupManager) maintainRunningArchiveLogJob() error {
	logger := m.Logger
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		logger.Error(err, "Failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestArchiveLogJob()
	if err != nil {
		logger.Error(err, "Failed to query latest archive log job")
		return err
	}
	if latest != nil {
		job.Status.ArchiveLogJob = latest
		if latest.StartScnDisplay != nil {
			job.Status.StartedAt = *latest.StartScnDisplay
		}
		job.Status.EndedAt = latest.CheckpointScnDisplay
		switch latest.Status {
		case "STOP":
			job.Status.Status = constants.BackupJobStatusStopped
		case "DOING":
			job.Status.Status = constants.BackupJobStatusRunning
		case "SUSPEND":
			job.Status.Status = constants.BackupJobStatusSuspend
		}
	}
	return nil
}
