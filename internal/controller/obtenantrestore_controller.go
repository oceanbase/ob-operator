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

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	restenantrestore "github.com/oceanbase/ob-operator/internal/resource/obtenantrestore"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// OBTenantRestoreReconciler reconciles a OBTenantRestore object
type OBTenantRestoreReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBTenantRestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = req

	logger := log.FromContext(ctx)
	restore := &v1alpha1.OBTenantRestore{}
	err := r.Client.Get(ctx, req.NamespacedName, restore)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	mgr := &restenantrestore.ObTenantRestoreManager{
		Ctx:      ctx,
		Resource: restore,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}

	coordinator := coordinator.NewCoordinator(mgr, &logger)
	_, err = coordinator.Coordinate()
	return ctrl.Result{
		RequeueAfter: 10 * time.Second,
	}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantRestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBTenantRestore{}).
		WithEventFilter(preds).
		Complete(r)
}
