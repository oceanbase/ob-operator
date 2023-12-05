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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	resbackuppolicy "github.com/oceanbase/ob-operator/internal/resource/obtenantbackuppolicy"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
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
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *OBTenantBackupPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	policy := &v1alpha1.OBTenantBackupPolicy{}
	err := r.Client.Get(ctx, req.NamespacedName, policy)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	finalizerName := "obtenantbackuppolicy.finalizers.oceanbase.com"
	// examine DeletionTimestamp to determine if the policy is under deletion
	if policy.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(policy, finalizerName) {
			controllerutil.AddFinalizer(policy, finalizerName)
			if err := r.Update(ctx, policy); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// create backup policy manager
	mgr := &resbackuppolicy.ObTenantBackupPolicyManager{
		Ctx:          ctx,
		BackupPolicy: policy,
		Client:       r.Client,
		Logger:       &logger,
		Recorder:     telemetry.NewRecorder(ctx, r.Recorder),
	}

	coordinator := coordinator.NewCoordinator(mgr, &logger)
	return coordinator.Coordinate()
}

// SetupWithManager sets up the controller with the Manager.
func (r *OBTenantBackupPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBTenantBackupPolicy{}).
		WithEventFilter(preds).
		Complete(r)
}
