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
	"errors"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	ctlconfig "github.com/oceanbase/ob-operator/internal/controller/config"
	"github.com/oceanbase/ob-operator/internal/telemetry"
)

// OBResourceRescueReconciler reconciles a OBResourceRescue object
type OBResourceRescueReconciler struct {
	client.Client
	Dynamic  dynamic.Interface
	Scheme   *runtime.Scheme
	Recorder telemetry.Recorder
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obresourcerescues/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *OBResourceRescueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	rescue := v1alpha1.OBResourceRescue{}
	if err := r.Client.Get(ctx, req.NamespacedName, &rescue); err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if rescue.Status.Status == "Successful" {
		return ctrl.Result{}, nil
	}

	switch rescue.Spec.TargetKind {
	case "OBCluster", "OBParameter", "OBServer", "OBZone", "OBTenant", "OBTenantBackupPolicy", "OBTenantBackup", "OBTenantRestore", "OBTenantOperation":
		gvStr := v1alpha1.GroupVersion.String()
		if rescue.Spec.TargetGV != "" {
			gvStr = rescue.Spec.TargetGV
		}

		gvk := schema.FromAPIVersionAndKind(gvStr, rescue.Spec.TargetKind)
		mapping, err := r.Client.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			logger.Error(err, "failed to get REST mapping")
			return ctrl.Result{}, err
		}

		uns, err := r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Get(ctx, rescue.Spec.TargetResName, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "failed to get the target resource")
			return ctrl.Result{}, err
		}

		switch rescue.Spec.Type {
		case "delete":
			uns.SetFinalizers(nil)
			_, err := r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Update(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "failed to update finalizers of the target resource")
				return ctrl.Result{}, err
			}
			if uns.GetDeletionTimestamp() == nil {
				err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Delete(ctx, rescue.Spec.TargetResName, metav1.DeleteOptions{})
				if err != nil {
					logger.Error(err, "failed to delete the target resource")
					return ctrl.Result{}, err
				}
			}
		case "reset":
			err := errors.Join(
				unstructured.SetNestedField(uns.Object, nil, "status", "operationContext"),
				unstructured.SetNestedField(uns.Object, rescue.Spec.TargetStatus, "status", "status"),
			)
			if err != nil {
				logger.Error(err, "failed to reset fields of the target resource")
				return ctrl.Result{}, err
			}
			_, err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).UpdateStatus(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "failed to update status of the target resource")
				return ctrl.Result{}, err
			}
		}

		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := r.Client.Get(ctx, req.NamespacedName, &rescue); err != nil {
				if kubeerrors.IsNotFound(err) {
					return nil
				}
				return err
			}
			rescue.Status.Status = "Successful"
			return r.Client.Status().Update(ctx, &rescue)
		})

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBResourceRescueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	kubeconfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	r.Dynamic, err = dynamic.NewForConfig(kubeconfig)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		WithEventFilter(preds).
		For(&v1alpha1.OBResourceRescue{}).
		Complete(r)
}

func NewOBResourceRescueReconciler(mgr ctrl.Manager) *OBResourceRescueReconciler {
	return &OBResourceRescueReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: telemetry.NewRecorder(context.Background(), mgr.GetEventRecorderFor(ctlconfig.OBResourceRescueControllerName)),
	}
}
