/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/oceanbase/ob-operator/pkg/resource"
	"github.com/oceanbase/ob-operator/pkg/telemetry"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
)

// OBTenantBackupReconciler reconciles a OBTenantBackup object
type OBTenantBackupReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	Telemetry telemetry.Telemetry

	telemetryOnce sync.Once
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBTenantBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	crJob := &v1alpha1.OBTenantBackup{}
	if err := r.Get(ctx, req.NamespacedName, crJob); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch crJob.Spec.Type {
	case constants.BackupJobTypeFull:
		fallthrough
	case constants.BackupJobTypeIncr:
		switch crJob.Status.Status {
		case "":
			fallthrough
		case constants.BackupJobStatusInitializing:
			crJob.Status.Status = constants.BackupJobStatusRunning
			return ctrl.Result{
				RequeueAfter: time.Second * 5,
			}, r.createBackupJobInOB(ctx, crJob)
		case constants.BackupJobStatusRunning:
			return ctrl.Result{
				RequeueAfter: time.Second * 5,
			}, r.maintainRunningBackupJob(ctx, crJob)
		default:
			// Completed, Failed, Canceled, do nothing
			return ctrl.Result{}, nil
		}

	case constants.BackupJobTypeArchive:
		return ctrl.Result{
			RequeueAfter: time.Second * 5,
		}, r.maintainRunningArchiveLogJob(ctx, crJob)
	case constants.BackupJobTypeClean:
		return ctrl.Result{
			RequeueAfter: time.Second * 5,
		}, r.maintainRunningBackupCleanJob(ctx, crJob)
	}

	return ctrl.Result{
		RequeueAfter: time.Second * 5,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	jobStatusKey := ".status.status"
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.OBTenantBackup{}, jobStatusKey, func(rawObj client.Object) []string {
		job := rawObj.(*v1alpha1.OBTenantBackup)
		return []string{string(job.Status.Status)}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBTenantBackup{}).
		Complete(r)
}

func (r *OBTenantBackupReconciler) createBackupJobInOB(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	logger := log.FromContext(ctx)
	con, err := r.getObOperationClient(ctx, job)
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}
	if job.Spec.EncryptionSecret != "" {
		password, err := resource.ReadPassword(r.Client, job.Namespace, job.Spec.EncryptionSecret)
		if err != nil {
			logger.Error(err, "failed to read backup encryption secret")
			r.getTelemetry(ctx).Event(job, "Warning", "ReadBackupEncryptionSecretFailed", err.Error())
		} else if password != "" {
			err = con.SetBackupPassword(password)
			if err != nil {
				logger.Error(err, "failed to set backup password")
				r.getTelemetry(ctx).Event(job, "Warning", "SetBackupPasswordFailed", err.Error())
			}
		}
	}
	latest, err := con.CreateAndReturnBackupJob(job.Spec.Type)
	if err != nil {
		logger.Error(err, "failed to create and return backup job")
		r.getTelemetry(ctx).Event(job, "Warning", "CreateAndReturnBackupJobFailed", err.Error())
		return err
	}

	job.Status.BackupJob = latest
	err = r.retryUpdateStatus(ctx, job)
	if err != nil {
		logger.Error(err, "failed to update status")
		r.getTelemetry(ctx).Event(job, "Warning", "UpdateStatusFailed", err.Error())
		return err
	}
	r.getTelemetry(ctx).Event(job, "Create", "", "create backup job successfully")
	return nil
}

// TODO: Calculate the progress of running jobs

func (r *OBTenantBackupReconciler) maintainRunningBackupJob(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	logger := log.FromContext(ctx)
	con, err := r.getObOperationClient(ctx, job)
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
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
	return r.retryUpdateStatus(ctx, job)
}

func (r *OBTenantBackupReconciler) maintainRunningBackupCleanJob(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	logger := log.FromContext(ctx)
	con, err := r.getObOperationClient(ctx, job)
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestBackupCleanJob()
	if err != nil {
		logger.Error(err, "failed to query latest backup clean job")
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
		return r.retryUpdateStatus(ctx, job)
	}

	return nil
}

func (r *OBTenantBackupReconciler) maintainRunningArchiveLogJob(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	logger := log.FromContext(ctx)
	con, err := r.getObOperationClient(ctx, job)
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}

	latest, err := con.GetLatestArchiveLogJob()
	if err != nil {
		logger.Error(err, "failed to query latest archive log job")
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
		return r.retryUpdateStatus(ctx, job)
	}

	return nil
}

func (r *OBTenantBackupReconciler) getObOperationClient(ctx context.Context, job *v1alpha1.OBTenantBackup) (*operation.OceanbaseOperationManager, error) {
	var err error
	logger := log.FromContext(ctx)
	obcluster := &v1alpha1.OBCluster{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      job.Spec.ObClusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := resource.GetTenantRootOperationClient(r.Client, &logger, obcluster, job.Spec.TenantName, job.Spec.TenantSecret)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (r *OBTenantBackupReconciler) retryUpdateStatus(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newestJob := &v1alpha1.OBTenantBackup{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: job.GetNamespace(),
			Name:      job.GetName(),
		}, newestJob)
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		newestJob.Status = job.Status
		return r.Status().Update(ctx, newestJob)
	})
}

func (r *OBTenantBackupReconciler) getTelemetry(ctx context.Context) telemetry.Telemetry {
	if r.Telemetry != nil {
		return r.Telemetry
	}
	r.telemetryOnce.Do(func() {
		r.Telemetry = telemetry.NewTelemetry(ctx, r.Recorder)
	})
	return r.Telemetry
}
