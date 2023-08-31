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

	oceanbaseconst "github.com/oceanbase/ob-operator/pkg/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/operation"
	"github.com/pkg/errors"
)

// OBTenantBackupReconciler reconciles a OBTenantBackup object
type OBTenantBackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	con *operation.OceanbaseOperationManager
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
	case v1alpha1.BackupJobTypeFull:
		fallthrough
	case v1alpha1.BackupJobTypeIncr:
		switch crJob.Status.Status {
		case "":
			fallthrough
		case v1alpha1.BackupJobStatusInitializing:
			crJob.Status.Status = v1alpha1.BackupJobStatusRunning
			return ctrl.Result{}, r.createBackupJobInOB(ctx, crJob)
		case v1alpha1.BackupJobStatusRunning:
			return ctrl.Result{}, r.maintainRunningBackupJob(ctx, crJob)
		default:
			// Completed, Failed, Canceled, do nothing
			return ctrl.Result{}, nil
		}

	case v1alpha1.BackupJobTypeArchive:
		return ctrl.Result{}, r.maintainRunningArchiveLogJob(ctx, crJob)
	case v1alpha1.BackupJobTypeClean:
		return ctrl.Result{}, r.maintainRunningBackupCleanJob(ctx, crJob)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
	latest, err := con.CreateAndReturnBackupJob(job.Spec.Type)
	if err != nil {
		logger.Error(err, "failed to create and return backup job")
		return err
	}

	job.Status.BackupJob = latest
	return r.Status().Update(ctx, job)
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
		if job.Spec.Type == v1alpha1.BackupJobTypeFull || job.Spec.Type == v1alpha1.BackupJobTypeIncr {
			latest, err := con.QueryLatestBackupJobOfType(job.Spec.Type)
			if err != nil {
				return err
			}
			job.Status.BackupJob = latest
			targetJob = latest
		}
		// archive log and data clean job should not be here
	} else {
		modelJob, err := con.QueryBackupJobWithId(job.Status.BackupJob.JobId)
		if err != nil {
			return err
		}
		if modelJob == nil {
			return errors.New(fmt.Sprintf("backup job with id %d not found", job.Status.BackupJob.JobId))
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
		job.Status.Status = v1alpha1.BackupJobStatusSuccessful
	case "FAILED":
		job.Status.Status = v1alpha1.BackupJobStatusFailed
	case "CANCELED":
		job.Status.Status = v1alpha1.BackupJobStatusCanceled
	}
	return r.Client.Status().Update(ctx, job)
}

func (r *OBTenantBackupReconciler) maintainRunningBackupCleanJob(ctx context.Context, job *v1alpha1.OBTenantBackup) error {
	logger := log.FromContext(ctx)
	con, err := r.getObOperationClient(ctx, job)
	if err != nil {
		logger.Error(err, "failed to get ob operation client")
		return err
	}

	latest, err := con.QueryLatestBackupCleanJob()
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
			job.Status.Status = v1alpha1.BackupJobStatusSuccessful
		case "FAILED":
			job.Status.Status = v1alpha1.BackupJobStatusFailed
		case "CANCELED":
			job.Status.Status = v1alpha1.BackupJobStatusCanceled
		case "DOING":
			job.Status.Status = v1alpha1.BackupJobStatusRunning
		}
		return r.Client.Status().Update(ctx, job)
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

	latest, err := con.QueryLatestArchiveLogJob()
	if err != nil {
		logger.Error(err, "failed to query latest archive log job")
		return err
	}
	if latest != nil {
		job.Status.ArchiveLogJob = latest
		job.Status.StartedAt = latest.StartScnDisplay
		job.Status.EndedAt = latest.CheckpointScnDisplay
		switch latest.Status {
		case "STOP":
			job.Status.Status = v1alpha1.BackupJobStatusStopped
		case "DOING":
			job.Status.Status = v1alpha1.BackupJobStatusRunning
		}
		return r.Client.Status().Update(ctx, job)
	}

	return nil
}

func (r *OBTenantBackupReconciler) getObOperationClient(ctx context.Context, job *v1alpha1.OBTenantBackup) (*operation.OceanbaseOperationManager, error) {
	if r.con != nil {
		return r.con, nil
	}
	logger := log.FromContext(ctx)
	clusterName, _ := job.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: job.Namespace,
		Name:      clusterName,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := resource.GetTenantOperationClient(r.Client, &logger, obcluster, job.Spec.TenantName)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	r.con = con
	return con, nil
}
