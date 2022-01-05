/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package core

import (
	"context"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	statefulappconst "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core/judge"
)

// StatefulAppReconciler reconciles a StatefulApp object
type StatefulAppReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/events/event.go
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps/sfinalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *StatefulAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.StatefulApp{}
	err := r.CRClient.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			// Object not found, return.
			// Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object, requeue the request.
		return reconcile.Result{}, err
	}
	// custom logic
	return r.StatefulAppCoordinator(*instance)
}

// StatefulAppCoordinator is the entry function for control-plan logic
func (r *StatefulAppReconciler) StatefulAppCoordinator(statefulApp cloudv1.StatefulApp) (ctrl.Result, error) {
	var err error
	subsetCtrl := NewSubsetCtrl(r.CRClient, r.Recorder, statefulApp)

	subsetsSpec := statefulApp.Spec.Subsets
	subsetsCurrentNameList := subsetCtrl.GetSubsetsNameList()

	// ScaleUP, add subset
	// ScaleDown, delete subsets
	// Maintain, update cluster status
	scaleState, subsetSpecName := judge.SubsetScaleJundge(subsetsSpec, subsetsCurrentNameList)
	switch scaleState {
	case statefulappconst.ScaleUP:
		// create subset
		err = subsetCtrl.CreateSubset(subsetSpecName, subsetsSpec)
	case statefulappconst.ScaleDown:
		// delete subset
		err = subsetCtrl.DeleteSubset(subsetSpecName, subsetsCurrentNameList)
	case statefulappconst.Maintain:
		err = r.SubsetsCoordinator(statefulApp, subsetsSpec)
	}

	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
