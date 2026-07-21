/*
Copyright (c) 2025 OceanBase
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

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	resoblogservicenode "github.com/oceanbase/ob-operator/internal/resource/oblogservicenode"
	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/coordinator"
)

type OBLogServiceNodeReconciler struct {
	Client    client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
}

func (r *OBLogServiceNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	resource := &v1alpha1.OBLogServiceNode{}
	err := r.Client.Get(ctx, req.NamespacedName, resource)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Get OBLogServiceNode error")
		return ctrl.Result{}, err
	}

	manager := &resoblogservicenode.OBLogServiceNodeManager{
		Ctx:       ctx,
		Resource:  resource,
		Client:    r.Client,
		APIReader: r.APIReader,
		Logger:    &logger,
		Recorder:  telemetry.NewRecorder(ctx, r.Recorder),
	}
	coord := coordinator.NewCoordinator(manager, &logger)
	return coord.Coordinate()
}

func (r *OBLogServiceNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.APIReader == nil {
		r.APIReader = mgr.GetAPIReader()
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.OBLogServiceNode{}).
		Owns(&corev1.Pod{}).
		WithEventFilter(preds).
		Complete(r)
}
