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
	"strings"
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
	"github.com/oceanbase/ob-operator/internal/clientcache"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
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
		logger.Error(err, "Get observer error")
		return ctrl.Result{}, err
	}

	// create observer manager
	observerManager := &resobserver.OBServerManager{
		Ctx:          ctx,
		OBServer:     observer,
		Client:       r.Client,
		Logger:       &logger,
		Recorder:     telemetry.NewRecorder(ctx, r.Recorder),
		K8sResClient: r.Client,
	}

	if observer.Spec.K8sCluster != "" {
		resClient, err := clientcache.GetCachedCtrlRuntimeClientFromK8sName(ctx, observer.Spec.K8sCluster)
		if err != nil {
			logger.Error(err, "Failed to get get client from k8s cluster "+observer.Spec.K8sCluster)
			return ctrl.Result{}, err
		}
		observerManager.K8sResClient = resClient
	}

	// execute finalizers
	finalizerName := strings.Join([]string{oceanbaseconst.FinalizerOBServer, observer.Name}, ".")
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
				logger.Error(err, "Delete observer failed")
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
