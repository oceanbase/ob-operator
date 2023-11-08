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

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/resource"
	"github.com/oceanbase/ob-operator/pkg/telemetry"
)

// OBZoneReconciler reconciles a OBZone object
type OBZoneReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obzones/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OBZone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBZoneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	obzone := &v1alpha1.OBZone{}
	err := r.Client.Get(ctx, req.NamespacedName, obzone)
	if err != nil {
		logger.Error(err, "get obzone error")
		if kubeerrors.IsNotFound(err) {
			// obzone not found, just return
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// TODO add finalizers

	// create zone manager
	obzoneManager := &resource.OBZoneManager{
		Ctx:      ctx,
		OBZone:   obzone,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}
	coordinator := resource.NewCoordinator(obzoneManager, &logger)
	return coordinator.Coordinate()
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBZoneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBZone{}).
		Complete(r)
}
