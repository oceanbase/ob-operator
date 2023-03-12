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

package tenantBackup

import (
	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	tenantBackupconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/core"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube"
	"github.com/oceanbase/ob-operator/pkg/kubeclient"
)

var (
	controllerKind = cloudv1.SchemeGroupVersion.WithKind("TenantBackup")
)

// Add creates a new Controller and adds it to the Manager with default RBAC.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	if !kube.DiscoverGVK(controllerKind) {
		return nil
	}
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &core.TenantBackupReconciler{
		CRClient: kubeclient.NewClientFromManager(mgr),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(tenantBackupconst.ControllerName),
	}
}

// add a new Controller to mgr with r
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(
		tenantBackupconst.ControllerName,
		mgr,
		controller.Options{
			Reconciler:              r,
			MaxConcurrentReconciles: tenantBackupconst.ConcurrentReconciles,
		},
	)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	// Watch for changes to TenantBackup
	err = c.Watch(
		&source.Kind{Type: &cloudv1.TenantBackup{}},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}
