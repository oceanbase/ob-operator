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

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiconsts "github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	res "github.com/oceanbase/ob-operator/internal/resource/obclusteroperation"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// OBClusterOperationReconciler reconciles a OBClusterOperation object
type OBClusterOperationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusteroperations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *OBClusterOperationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	op := &v1alpha1.OBClusterOperation{}
	err := r.Client.Get(ctx, req.NamespacedName, op)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get cluster operation")
		return ctrl.Result{}, err
	}

	switch op.Status.Status {
	case apiconsts.ClusterOpStatusSucceeded, apiconsts.ClusterOpStatusFailed:
		if op.ShouldBeCleaned() {
			if err := r.Client.Delete(ctx, op); err != nil {
				logger.Error(err, "Failed to delete stale cluster operation")
				return ctrl.Result{}, err
			}
		}
	}

	// create cluster operation manager
	clusterOpManager := &res.OBClusterOperationManager{
		Ctx:      ctx,
		Resource: op,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}

	cood := coordinator.NewCoordinator(clusterOpManager, &logger)
	return cood.Coordinate()
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBClusterOperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBClusterOperation{}).
		Complete(r)
}
