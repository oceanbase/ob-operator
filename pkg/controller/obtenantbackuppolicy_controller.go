/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
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

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/pkg/resource"
)

// OBTenantBackupPolicyReconciler reconciles a OBTenantBackupPolicy object
type OBTenantBackupPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackuppolicies/finalizers,verbs=update

//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackup,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=oceanbase.oceanbase.com,resources=obtenantbackup/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OBTenantBackupPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *OBTenantBackupPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	policy := &v1alpha1.OBTenantBackupPolicy{}
	err := r.Client.Get(ctx, req.NamespacedName, policy)
	if err != nil {
		logger.Error(err, "get backup policy error")
		if kubeerrors.IsNotFound(err) {
			// backup policy not found, just return
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	logger.Info("reconcile backup policy:", "spec", policy.Spec, "status", policy.Status)

	// create backup policy manager
	mgr := &resource.ObTenantBackupPolicyManager{
		Ctx:          ctx,
		BackupPolicy: policy,
		Client:       r.Client,
		Recorder:     r.Recorder,
		Logger:       &logger,
	}
	coordinator := resource.NewCoordinator(mgr, &logger)
	err = coordinator.Coordinate()
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantBackupPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBTenantBackupPolicy{}).
		Complete(r)
}
