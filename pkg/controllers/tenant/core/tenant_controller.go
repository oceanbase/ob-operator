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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	tenantconst "github.com/oceanbase/ob-operator/pkg/controllers/tenant/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	util "github.com/oceanbase/ob-operator/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type TenantCtrl struct {
	Resource *resource.Resource
	Tenant   cloudv1.Tenant
}

type TenantCtrlOperator interface {
	TenantCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenants/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the tenant CR instance
	instance := &cloudv1.Tenant{}
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
	tenantCtrl := NewTenantCtrl(r.CRClient, r.Recorder, *instance)

	// Fetch the OBCluster CR instance
	obNamespace := types.NamespacedName{
		Namespace: instance.Namespace,
		Name:      instance.Spec.ClusterName,
	}
	obInstance := &cloudv1.OBCluster{}
	err = r.CRClient.Get(ctx, obNamespace, obInstance)
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			klog.Infof("OBCluster %s not found, namespace %s", instance.Spec.ClusterName, instance.Namespace)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if obInstance.Status.Status != observerconst.ClusterReady {
		klog.Infoln("OBCluster  %s is not ready, namespace %s", instance.Spec.ClusterName, instance.Namespace)
		return reconcile.Result{}, nil
	}

	// Handle deleted tenant
	tenantFinalizerName := fmt.Sprintf("cloud.oceanbase.com.finalizers.%s", instance.Name)
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !util.ContainsString(instance.ObjectMeta.Finalizers, tenantFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, tenantFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if util.ContainsString(instance.ObjectMeta.Finalizers, tenantFinalizerName) {
			err := r.TenantDelete(r.CRClient, r.Recorder, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.ObjectMeta.Finalizers = util.RemoveString(instance.ObjectMeta.Finalizers, tenantFinalizerName)
			if err := r.CRClient.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// custom logic
	return tenantCtrl.TenantCoordinator()
}

func NewTenantCtrl(client client.Client, recorder record.EventRecorder, tenant cloudv1.Tenant) TenantCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &TenantCtrl{
		Resource: ctrlResource,
		Tenant:   tenant,
	}
}

func (r *TenantReconciler) TenantDelete(client client.Client, recorder record.EventRecorder, tenant *cloudv1.Tenant) error {
	ctrlResource := resource.NewResource(client, recorder)
	ctrl := &TenantCtrl{
		Tenant:   *tenant,
		Resource: ctrlResource,
	}
	tenantName := ctrl.Tenant.Name
	klog.Infof("Begin Delete Tenant '%s'", tenantName)
	err := ctrl.DeleteTenant()
	if err != nil {
		return err
	}
	klog.Infof("Begin Delete Pool, Tenant '%s'", tenantName)
	err = ctrl.DeletePool()
	if err != nil {
		return err
	}
	klog.Infof("Begin Delete Unit, Tenant '%s'", tenantName)
	err = ctrl.DeleteUnit()
	if err != nil {
		return err
	}
	klog.Infof("Succeed Delete Tenant '%s'", tenantName)
	return nil
}

func (ctrl *TenantCtrl) TenantCoordinator() (ctrl.Result, error) {
	err := ctrl.TenantEffector()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (ctrl *TenantCtrl) TenantEffector() error {
	var err error
	tenantStatus := ctrl.Tenant.Status.Status
	switch tenantStatus {
	case "", tenantconst.TenantCreating:
		err = ctrl.TenantCreatingEffector()
	case tenantconst.TenantRunning, tenantconst.TenantModifying:
		err = ctrl.TenantRunningEffector()
	}
	return err
}

func (ctrl *TenantCtrl) TenantCreatingEffector() error {
	var err error
	tenantName := ctrl.Tenant.Name
	tenantExist, _, err := ctrl.TenantExist(tenantName)
	if err != nil {
		klog.Errorln("Check Whether The Tenant %s Exists Error: %s", tenantName, err)
		return err
	}
	if tenantExist {
		return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
	}

	for _, zone := range ctrl.Tenant.Spec.Topology {
		err := ctrl.CheckAndCreateUnitAndPool(tenantName, zone)
		if err != nil {
			return err
		}
	}
	err = ctrl.CreateTenant(tenantName, ctrl.Tenant.Spec.Topology)
	if err != nil {
		klog.Errorln("Create Tenant '%s' Error: %s", tenantName, err)
		return err
	}
	klog.Infof("Create Tenant '%s' OK", tenantName)
	return ctrl.UpdateTenantStatus(tenantconst.TenantRunning)
}

func (ctrl *TenantCtrl) TenantRunningEffector() error {
	err := ctrl.CheckAndSetVariables()
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetUnitConfig()
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetResourcePool()
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetTenant()
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *TenantCtrl) GetSqlOperator() (*sql.SqlOperator, error) {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.Tenant.Namespace, ctrl.Tenant.Spec.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	secretName := converter.GenerateSecretNameForDBUser(ctrl.Tenant.Spec.ClusterName, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.Tenant.Namespace, secretName)
	user := "root"
	password := ""
	if err == nil {
		user = "admin"
		password = string(secret.(corev1.Secret).Data["password"])
	}

	p := &sql.DBConnectProperties{
		IP:       clusterIP,
		Port:     observerconst.MysqlPort,
		User:     user,
		Password: password,
		Database: "oceanbase",
		Timeout:  10,
	}
	so := sql.NewSqlOperator(p)
	if so.TestOK() {
		return so, nil
	}
	return nil, errors.New("failed to get sql operator")
}

func (ctrl *TenantCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
