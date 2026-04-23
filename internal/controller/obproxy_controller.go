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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resobproxy "github.com/oceanbase/ob-operator/internal/resource/obproxy"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

// obServerWatchPredicates enqueues OBProxy when OBServer spec changes (generation) or
// status/address/labels change, so RS_LIST can refresh when observers become Ready without spec bump.
// Global GenerationChangedPredicate is not used for this watch (see SetupWithManager).
var obServerWatchPredicates = predicate.Or(
	predicate.GenerationChangedPredicate{},
	predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldO, ok1 := e.ObjectOld.(*v1alpha1.OBServer)
			newO, ok2 := e.ObjectNew.(*v1alpha1.OBServer)
			if !ok1 || !ok2 {
				return true
			}
			if oldO.Status.Status != newO.Status.Status ||
				oldO.Status.GetConnectAddr() != newO.Status.GetConnectAddr() {
				return true
			}
			oldRef := ""
			if oldO.Labels != nil {
				oldRef = oldO.Labels[oceanbaseconst.LabelRefOBCluster]
			}
			newRef := ""
			if newO.Labels != nil {
				newRef = newO.Labels[oceanbaseconst.LabelRefOBCluster]
			}
			return oldRef != newRef
		},
	},
)

// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obproxies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obproxies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obproxies/finalizers,verbs=update
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=observers,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

// OBProxyReconciler reconciles an OBProxy object.
type OBProxyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *OBProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	obproxy := &v1alpha1.OBProxy{}
	err := r.Client.Get(ctx, req.NamespacedName, obproxy)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// obproxy not found, just return
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Get obproxy error")
		return ctrl.Result{}, err
	}

	// Create obproxy manager
	obproxyManager := &resobproxy.OBProxyManager{
		Ctx:      ctx,
		OBProxy:  obproxy,
		Client:   r.Client,
		Logger:   &logger,
		Recorder: telemetry.NewRecorder(ctx, r.Recorder),
	}
	coordinator := coordinator.NewCoordinator(obproxyManager, &logger)
	return coordinator.Coordinate()
}

// obClusterStatusPredicate triggers only when OBCluster status changes,
// so OBProxy can start its create flow once OBCluster becomes Running.
var obClusterStatusPredicate = predicate.Funcs{
	CreateFunc: func(_ event.CreateEvent) bool { return false },
	DeleteFunc: func(_ event.DeleteEvent) bool { return false },
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldC, ok1 := e.ObjectOld.(*v1alpha1.OBCluster)
		newC, ok2 := e.ObjectNew.(*v1alpha1.OBCluster)
		if !ok1 || !ok2 {
			return true
		}
		return oldC.Status.Status != newC.Status.Status
	},
	GenericFunc: func(_ event.GenericEvent) bool { return false },
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBProxy{}, builder.WithPredicates(preds)).
		Watches(
			&v1alpha1.OBServer{},
			handler.EnqueueRequestsFromMapFunc(r.mapOBServerToOBProxy),
			builder.WithPredicates(obServerWatchPredicates),
		).
		Watches(
			&v1alpha1.OBCluster{},
			handler.EnqueueRequestsFromMapFunc(r.mapOBClusterToOBProxy),
			builder.WithPredicates(obClusterStatusPredicate),
		).
		Complete(r)
}

// mapOBClusterToOBProxy enqueues OBProxy reconcile requests when the referenced OBCluster status changes.
func (r *OBProxyReconciler) mapOBClusterToOBProxy(ctx context.Context, obj client.Object) []reconcile.Request {
	obcluster, ok := obj.(*v1alpha1.OBCluster)
	if !ok {
		return nil
	}

	obproxyList := &v1alpha1.OBProxyList{}
	if err := r.Client.List(ctx, obproxyList); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, proxy := range obproxyList.Items {
		ns := proxy.Spec.OBCluster.Namespace
		if ns == "" {
			ns = proxy.Namespace
		}
		if proxy.Spec.OBCluster.Name == obcluster.Name && ns == obcluster.Namespace {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      proxy.Name,
					Namespace: proxy.Namespace,
				},
			})
		}
	}
	return requests
}

// mapOBServerToOBProxy maps OBServer changes to OBProxy reconcile requests.
// When an OBServer is added/deleted or its status changes, we need to update
// the RS_LIST in all OBProxies that reference the same OBCluster.
func (r *OBProxyReconciler) mapOBServerToOBProxy(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)
	observer, ok := obj.(*v1alpha1.OBServer)
	if !ok {
		return nil
	}

	// Get the OBCluster name from OBServer labels
	obclusterName, hasLabel := observer.Labels[oceanbaseconst.LabelRefOBCluster]
	if !hasLabel || obclusterName == "" {
		return nil
	}

	logger.Info("OBServer event detected, will check related OBProxies",
		"observer", observer.Name,
		"namespace", observer.Namespace,
		"obcluster", obclusterName,
		"observerStatus", observer.Status.Status,
		"observerIP", observer.Status.ServiceIp)

	// Find all OBProxies that reference this OBCluster
	obproxyList := &v1alpha1.OBProxyList{}
	err := r.Client.List(ctx, obproxyList)
	if err != nil {
		logger.Error(err, "Failed to list OBProxies")
		return nil
	}

	var requests []reconcile.Request
	for _, obproxy := range obproxyList.Items {
		// Check if this OBProxy references the same OBCluster
		if obproxy.Spec.OBCluster.Name == obclusterName {
			// Check namespace match - OBProxy can reference OBCluster in same or different namespace
			obclusterNS := obproxy.Spec.OBCluster.Namespace
			if obclusterNS == "" {
				obclusterNS = obproxy.Namespace
			}
			// Only trigger if the OBServer is in the same namespace as the OBCluster
			if obclusterNS == observer.Namespace {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      obproxy.Name,
						Namespace: obproxy.Namespace,
					},
				})
				logger.Info("OBServer change triggers OBProxy reconciliation",
					"observer", observer.Name,
					"observerStatus", observer.Status.Status,
					"observerIP", observer.Status.ServiceIp,
					"obcluster", obclusterName,
					"obproxy", obproxy.Name,
					"obproxyNamespace", obproxy.Namespace,
					"reason", "RS_LIST may need update due to OBServer change")
			}
		}
	}

	if len(requests) > 0 {
		logger.Info("OBServer event will trigger reconciliation for OBProxies",
			"observer", observer.Name,
			"obcluster", obclusterName,
			"obproxyCount", len(requests))
	}

	return requests
}
