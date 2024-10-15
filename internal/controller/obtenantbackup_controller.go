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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	resobbackup "github.com/oceanbase/ob-operator/internal/resource/obtenantbackup"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// OBTenantBackupReconciler reconciles a OBTenantBackup object
type OBTenantBackupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBTenantBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	crJob := &v1alpha1.OBTenantBackup{}
	if err := r.Get(ctx, req.NamespacedName, crJob); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Do not reconcile if the backup job is already completed
	switch crJob.Spec.Type {
	case constants.BackupJobTypeFull, constants.BackupJobTypeIncr:
		switch crJob.Status.Status {
		case constants.BackupJobStatusCanceled, constants.BackupJobStatusSuccessful, constants.BackupJobStatusFailed:
			return ctrl.Result{}, nil
		}
	}

	mgr := &resobbackup.OBTenantBackupManager{
		Ctx:      ctx,
		Resource: crJob,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}

	result, err := coordinator.NewCoordinator(mgr, &logger).Coordinate()
	if result.RequeueAfter < time.Second*5 {
		result.RequeueAfter = time.Second * 5
	}

	return result, err
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
		WithEventFilter(preds).
		Complete(r)
}
