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
	resobparameter "github.com/oceanbase/ob-operator/internal/resource/obparameter"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// OBParameterReconciler reconciles a OBParameter object
type OBParameterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OBParameter object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBParameterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	obparameter := &v1alpha1.OBParameter{}
	err := r.Client.Get(ctx, req.NamespacedName, obparameter)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// obparameter not found, just return
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Get obparameter error")
		return ctrl.Result{}, err
	}

	// TODO add finalizers

	// create cluster manager
	obparameterManager := &resobparameter.OBParameterManager{
		Ctx:         ctx,
		OBParameter: obparameter,
		Client:      r.Client,
		Logger:      &logger,
		Recorder:    telemetry.NewRecorder(ctx, r.Recorder),
	}
	coordinator := coordinator.NewCoordinator(obparameterManager, &logger)
	return coordinator.Coordinate()
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBParameterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBParameter{}).
		WithEventFilter(preds).
		Complete(r)
}
