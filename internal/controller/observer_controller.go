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
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	resobserver "github.com/oceanbase/ob-operator/internal/resource/observer"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// OBServerReconciler reconciles a OBServer object
type OBServerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OBServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *OBServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	observer := &v1alpha1.OBServer{}
	err := r.Client.Get(ctx, req.NamespacedName, observer)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// observer not found, just return
			return ctrl.Result{}, nil
		}
		logger.Error(err, "get observer error")
		return ctrl.Result{}, err
	}

	// create observer manager
	observerManager := &resobserver.OBServerManager{
		Ctx:      ctx,
		OBServer: observer,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}

	// execute finalizers
	finalizerName := fmt.Sprintf("observer.oceanbase.com.finalizers.%s", observer.Name)
	if !observer.ObjectMeta.DeletionTimestamp.IsZero() {
		needExecuteFinalizer := false
		for _, finalizer := range observer.ObjectMeta.Finalizers {
			if finalizer == finalizerName {
				needExecuteFinalizer = true
				break
			}
		}
		if needExecuteFinalizer {
			err = resobserver.DeleteOBServerInCluster(observerManager)
			if err != nil {
				logger.Error(err, "delete observer failed")
				return ctrl.Result{}, errors.Wrapf(err, "delete observer %s failed", observer.Name)
			}
		}
	}
	coordinator := coordinator.NewCoordinator(observerManager, &logger)
	result, err := coordinator.Coordinate()
	if err != nil {
		return result, err
	}
	if result.RequeueAfter > 5*time.Second {
		result.RequeueAfter = 5 * time.Second
	}
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBServer{}).
		Owns(&corev1.Pod{}).
		WithEventFilter(preds).
		Complete(r)
}
