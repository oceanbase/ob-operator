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
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
)

// OBResourceRescueReconciler reconciles a OBResourceRescue object
type OBResourceRescueReconciler struct {
	client.Client
	Dynamic  dynamic.Interface
	Scheme   *runtime.Scheme
	Recorder telemetry.Recorder
}

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
			logger.Error(err, "Failed to get REST mapping", "gvk", gvk)
			return ctrl.Result{}, err
		}

		uns, err := r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Get(ctx, rescue.Spec.TargetResName, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "Failed to get the target resource", "resource kind", rescue.Spec.TargetKind, "resource name", rescue.Spec.TargetResName)
			return ctrl.Result{}, err
		}

		switch rescue.Spec.Type {
		case "delete":
			uns.SetFinalizers(nil)
			_, err := r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Update(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "Failed to update finalizers of the target resource")
				return ctrl.Result{}, err
			}
			if uns.GetDeletionTimestamp() == nil {
				err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Delete(ctx, rescue.Spec.TargetResName, metav1.DeleteOptions{})
				if err != nil {
					logger.Error(err, "Failed to delete the target resource")
					return ctrl.Result{}, err
				}
			}
		case "reset":
			err := errors.Join(
				unstructured.SetNestedField(uns.Object, nil, "status", "operationContext"),
				unstructured.SetNestedField(uns.Object, rescue.Spec.TargetStatus, "status", "status"),
			)
			if err != nil {
				logger.Error(err, "Failed to reset fields of the target resource")
				return ctrl.Result{}, err
			}
			_, err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).UpdateStatus(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "Failed to update status of the target resource")
				return ctrl.Result{}, err
			}
		case "retry":
			// operationContext.TaskStatus = taskstatus.Pending
			// operationContext.FailureRule.RetryCount = 0
			context, exist, err := unstructured.NestedMap(uns.Object, "status", "operationContext")
			if err != nil {
				logger.Error(err, "Failed to get operationContext fields of the target resource")
				return ctrl.Result{}, nil
			}
			if !exist {
				logger.Info("OperationContext not found", "resource kind", uns.GetKind(), "resource name", uns.GetName())
				return ctrl.Result{}, nil
			}
			_, exist, err = unstructured.NestedMap(context, "failureRule")
			if err != nil {
				logger.Error(err, "Failed to get failureStrategy field of the target resource")
				return ctrl.Result{}, nil
			}
			if !exist {
				logger.Info("FailureStrategy not found", "resource kind", uns.GetKind(), "resource name", uns.GetName())
				return ctrl.Result{}, nil
			}

			// Only bool, int64, float64, string, []interface{}, map[string]interface{}, json.Number and nil are allowed to be set.
			var retryCount int64
			err = errors.Join(
				unstructured.SetNestedField(uns.Object, taskstatus.Pending, "status", "operationContext", "taskStatus"),
				unstructured.SetNestedField(uns.Object, retryCount, "status", "operationContext", "failureRule", "retryCount"),
			)
			if err != nil {
				logger.Error(err, "Failed to set operationContext fields of the target resource")
				return ctrl.Result{}, err
			}
			_, err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).UpdateStatus(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "Failed to update status of the target resource")
				return ctrl.Result{}, err
			}
		case "skip":
			// operationContext.TaskStatus = taskstatus.Successful
			// When coordinator finds that the task status is `successful`, it will go on the following steps.
			_, exist, err := unstructured.NestedMap(uns.Object, "status", "operationContext")
			if err != nil {
				logger.Error(err, "Failed to get operationContext fields of the target resource")
				return ctrl.Result{}, nil
			}
			if !exist {
				logger.Info("OperationContext not found", "resource kind", uns.GetKind(), "resource name", uns.GetName())
				return ctrl.Result{}, nil
			}
			err = unstructured.SetNestedField(uns.Object, taskstatus.Successful, "status", "operationContext", "taskStatus")
			if err != nil {
				logger.Error(err, "Failed to reset fields of the target resource")
				return ctrl.Result{}, err
			}
			_, err = r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).UpdateStatus(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "Failed to update status of the target resource")
				return ctrl.Result{}, err
			}
		case "ignore-deletion":
			annotations := uns.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations[oceanbaseconst.AnnotationsIgnoreDeletion] = "true"
			uns.SetAnnotations(annotations)
			_, err := r.Dynamic.Resource(mapping.Resource).Namespace(rescue.GetNamespace()).Update(ctx, uns, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "Failed to update annotations of the target resource")
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
