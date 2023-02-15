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
	"errors"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/tenant-backup/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TenantBackupReconciler reconciles a TenantBackup object
type TenantBackupReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme

	Recorder record.EventRecorder
}

type TenantBackupCtrl struct {
	Resource     *resource.Resource
	TenantBackup cloudv1.TenantBackup
}

type TenantBackupCtrlOperator interface {
	TenantBackupCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=tenantbackups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
func (r *TenantBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.TenantBackup{}
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
	tenantBackupCtrl := NewTenantBackupCtrl(r.CRClient, r.Recorder, *instance)
	return tenantBackupCtrl.TenantBackupCoordinator()
}

func NewTenantBackupCtrl(client client.Client, recorder record.EventRecorder, tenantBackup cloudv1.TenantBackup) TenantBackupCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &TenantBackupCtrl{
		Resource:     ctrlResource,
		TenantBackup: tenantBackup,
	}
}

func (ctrl *TenantBackupCtrl) TenantBackupCoordinator() (ctrl.Result, error) {
	// TenantBackup control-plan
	err := ctrl.TenantBackupEffector()
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (ctrl *TenantBackupCtrl) TenantBackupEffector() error {
	tenantList := ctrl.TenantBackup.Spec.Tenants
	for _, tenant := range tenantList {
		err := ctrl.SingleTenantBackupEffector(tenant)
		if err != nil {
			klog.Errorf("tenant '%s' backup failed, error '%s'", tenant.Name, err)
			continue
		}
	}
	return nil
}

func (ctrl *TenantBackupCtrl) SingleTenantBackupEffector(tenant cloudv1.TenantSpec) error {
	klog.Infoln("debug: SingleTenantBackupEffector: tenant ", tenant.Name)
	exist, backupTypeList := ctrl.CheckTenantBackupExist(tenant)
	if exist {
		klog.Infoln("debug: exist ", exist, backupTypeList)
		backupOnce, finished := ctrl.CheckTenantBackupOnce(tenant, backupTypeList)
		if backupOnce && finished {
			return nil
		}
	}
	return ctrl.SingleTenantBackup(tenant)
}

func (ctrl *TenantBackupCtrl) SingleTenantBackup(tenant cloudv1.TenantSpec) error {
	err := ctrl.CheckAndSetLogArchiveDest(tenant)
	if err != nil {
		klog.Errorf("tenant '%s' check and set LogArchiveDest error '%s'", tenant.Name, err)
		return err
	}
	err = ctrl.CheckAndStartArchive(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndSetBackupDest(tenant)
	if err != nil {
		return err
	}
	err = ctrl.CheckAndDoBackup(tenant)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *TenantBackupCtrl) GetSqlOperator() (*sql.SqlOperator, error) {
	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.TenantBackup.Namespace, ctrl.TenantBackup.Spec.SourceCluster.ClusterName)
	// get svc failed
	if err != nil {
		return nil, errors.New("failed to get service address")
	}
	secretName := converter.GenerateSecretNameForDBUser(ctrl.TenantBackup.Spec.SourceCluster.ClusterName, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.TenantBackup.Namespace, secretName)
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

func (ctrl *TenantBackupCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
