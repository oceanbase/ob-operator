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
	"github.com/oceanbase/ob-operator/pkg/resource"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
)

// OBTenantReconciler reconciles a OBTenant object
type OBTenantReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenants/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OBTenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBTenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	obtenant := &v1alpha1.OBTenant{}
	err := r.Client.Get(ctx, req.NamespacedName, obtenant)
	if err != nil {
		logger.Error(err, "get obtenant error")
		if kubeerrors.IsNotFound(err) {
			// observer not found, just return
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	logger.Info("reconcile obtenant begin:  >>>>>>>>>>>>>>> * <<<<<<<<<<<<<<<", "spec", obtenant.Spec, "status", obtenant.Status)

	// create observer manager
	obtenantManager := &resource.OBTenantManager{
		Ctx:      ctx,
		OBTenant: obtenant,
		Client:   r.Client,
		Recorder: r.Recorder,
		Logger:   &logger,
	}

	coordinator := resource.NewCoordinator(obtenantManager, &logger)
	err = coordinator.Coordinate()
	if err != nil {
		if kubeerrors.IsConflict(err) {
			obtenantManager.Logger.Error(err, "retry Reconcile tenant >>>>>>>>>>> * <<<<<<<<<<<<<")
			return ctrl.Result{Requeue: true}, nil
		}
	}
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBTenant{}).
		Complete(r)
}
