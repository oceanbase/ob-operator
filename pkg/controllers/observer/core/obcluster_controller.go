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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	statefulAppCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/core"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/judge"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

// OBClusterReconciler reconciles a OBCluster object
type OBClusterReconciler struct {
	CRClient client.Client
	Scheme   *runtime.Scheme
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/events/event.go
	Recorder record.EventRecorder
}

type OBClusterCtrl struct {
	Resource  *resource.Resource
	OBCluster cloudv1.OBCluster
}

type OBClusterCtrlOperator interface {
	OBClusterCoordinator() (ctrl.Result, error)
}

// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=rootservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=rootservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=rootservices/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obzones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obzones/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=obzones/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=statefulapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=backups/finalizers,verbs=update
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cloud.oceanbase.com,resources=restores/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *OBClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the CR instance
	instance := &cloudv1.OBCluster{}
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
	obClusterCtrl := NewOBServerCtrl(r.CRClient, r.Recorder, *instance)
	return obClusterCtrl.OBClusterCoordinator()
}

func NewOBServerCtrl(client client.Client, recorder record.EventRecorder, obCluster cloudv1.OBCluster) OBClusterCtrlOperator {
	ctrlResource := resource.NewResource(client, recorder)
	return &OBClusterCtrl{
		Resource:  ctrlResource,
		OBCluster: obCluster,
	}
}

func (ctrl *OBClusterCtrl) GetSqlOperatorFromStatefulApp(statefulApp cloudv1.StatefulApp) (*sql.SqlOperator, error) {
	podCtrl := &statefulAppCore.PodCtrl{
		Resource:    ctrl.Resource,
		StatefulApp: statefulApp,
	}
	return podCtrl.GetSqlOperator()
}

func (ctrl *OBClusterCtrl) GetSqlOperator(server ...string) (*sql.SqlOperator, error) {
	var clusterIP string
	var err error
	if server != nil {
		clusterIP = server[0]
	} else {
		clusterIP, err = ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
		// get svc failed
		if err != nil {
			return nil, errors.New("failed to get service address")
		}
	}

	secretName := converter.GenerateSecretNameForDBUser(ctrl.OBCluster.Name, "sys", "admin")
	secretExecutor := resource.NewSecretResource(ctrl.Resource)
	secret, err := secretExecutor.Get(context.TODO(), ctrl.OBCluster.Namespace, secretName)
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

func (ctrl *OBClusterCtrl) OBClusterCoordinator() (ctrl.Result, error) {
	var newClusterStatus bool
	statefulApp := &cloudv1.StatefulApp{}
	statefulApp, newClusterStatus = ctrl.IsNewCluster(*statefulApp)
	// is new cluster
	if newClusterStatus {
		err := ctrl.NewCluster(*statefulApp)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	// OBCluster control-plan
	err := ctrl.OBClusterEffector(*statefulApp)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (ctrl *OBClusterCtrl) OBClusterEffector(statefulApp cloudv1.StatefulApp) error {
	var err error
	obClusterStatus := ctrl.OBCluster.Status.Status
	switch obClusterStatus {
	case observerconst.TopologyPrepareing:
		// OBCluster is not ready
		err = ctrl.TopologyPrepareingEffector(statefulApp)
	case observerconst.TopologyNotReady:
		// OBCluster is not ready
		err = ctrl.TopologyNotReadyEffector(statefulApp)
	case observerconst.TopologyReady:
		// OBCluster is ready
		err = ctrl.TopologyReadyEffector(statefulApp)
	}
	return err
}

func (ctrl *OBClusterCtrl) TopologyPrepareingEffector(statefulApp cloudv1.StatefulApp) error {
	var err error

	for _, clusterStatus := range ctrl.OBCluster.Status.Topology {
		if clusterStatus.Cluster == myconfig.ClusterName {
			switch clusterStatus.ClusterStatus {
			case observerconst.ResourcePrepareing:
				// StatefulApp is creating
				err = ctrl.ResourcePrepareingEffectorForBootstrap(statefulApp)
			case observerconst.ResourceReady:
				// StatefulApp is ready
				err = ctrl.ResourceReadyEffectorForBootstrap(statefulApp)
			case observerconst.OBServerPrepareing:
				// OBServer is staring
				err = ctrl.OBServerPrepareingEffectorForBootstrap(statefulApp)
			case observerconst.OBServerReady:
				// OBServer is running
				err = ctrl.OBServerReadyEffectorForBootstrap(statefulApp)
			case observerconst.OBClusterBootstraping:
				// OBCluster Bootstraping
				err = ctrl.OBClusterBootstraping(statefulApp)
			case observerconst.OBClusterBootstrapReady:
				// OBCluster Bootstrap ready
				err = ctrl.OBClusterBootstrapReady(statefulApp)
			case observerconst.OBClusterReady:
				// OBCluster bootstrap succeeded
				err = ctrl.OBClusterReadyForStep(observerconst.StepBootstrap, statefulApp)
			}
		}
	}
	return err
}

func (ctrl *OBClusterCtrl) TopologyNotReadyEffector(statefulApp cloudv1.StatefulApp) error {
	var err error
	for _, clusterStatus := range ctrl.OBCluster.Status.Topology {
		if clusterStatus.Cluster == myconfig.ClusterName {
			switch clusterStatus.ClusterStatus {
			case observerconst.ScaleUP:
				// OBServer Scale UP
				err = ctrl.OBServerScaleUPByZone(statefulApp)
			case observerconst.ScaleDown:
				// OBServer Scale Down
				err = ctrl.OBServerScaleDownByZone(statefulApp)
				// OBZone Scale Up
			case observerconst.ZoneScaleUP:
				err = ctrl.OBZoneScaleUP(statefulApp)
				// OBZone Scale Down
			case observerconst.ZoneScaleDown:
				err = ctrl.OBZoneScaleDown(statefulApp)
			case observerconst.NeedUpgradeCheck:
				err = ctrl.ExecUpgradePreChecker(statefulApp)
			case observerconst.UpgradeChecking:
				err = ctrl.GetPreCheckJobStatus(statefulApp)
			case observerconst.NeedExecutingPreScripts:
				err = ctrl.CheckUpgradeModeBegin(statefulApp)
			case observerconst.ExecutingPreScripts:
				err = ctrl.ExecPreScripts(statefulApp)
			case observerconst.NeedUpgrading:
				err = ctrl.PreparingForUpgrade(statefulApp)
			case observerconst.Upgrading:
				err = ctrl.ExecUpgrading(statefulApp)
			case observerconst.ExecutingPostScripts:
				err = ctrl.ExecPostScripts(statefulApp)
			case observerconst.NeedUpgradePostCheck:
				err = ctrl.PrepareForPostCheck(statefulApp)
			case observerconst.UpgradePostChecking:
				err = ctrl.ExecUpgradePostChecker(statefulApp)
			}
		}
	}
	return err
}

func (ctrl *OBClusterCtrl) ZoneNumberIsModified() (string, error) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return "", errors.Wrap(err, "get sql operator when judge zone number")
	}

	cluster := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	zoneNumberNew := len(cluster.Zone)
	if zoneNumberNew == 0 {
		return observerconst.Maintain, kubeerrors.NewServiceUnavailable("can't scale Zone to zero")
	}

	obZoneList := sqlOperator.GetOBZone()
	zoneNumberCurrent := len(obZoneList)
	if zoneNumberCurrent == 0 {
		return "", errors.New(observerconst.DataBaseError)
	}
	if zoneNumberNew > zoneNumberCurrent {
		return observerconst.ZoneScaleUP, nil
	} else if zoneNumberNew < zoneNumberCurrent {
		return observerconst.ZoneScaleDown, nil
	} else {
		return observerconst.Maintain, nil
	}
}

func (ctrl *OBClusterCtrl) TopologyReadyEffector(statefulApp cloudv1.StatefulApp) error {
	// check parameter and version in obcluster, set parameter when modified
	ctrl.CheckAndSetParameters()

	// check version update
	versionIsModified, err := judge.VersionIsModified(ctrl.OBCluster.Spec.Tag, statefulApp)
	if err != nil {
		return err
	}
	if versionIsModified {
		// TODO: support version update
		err = ctrl.OBClusterUpgrade(statefulApp)
		return err
	}

	// check resource modified
	resourcesIsModified, err := judge.ResourcesIsModified(ctrl.OBCluster.Spec.Topology, ctrl.OBCluster, statefulApp)
	if err != nil {
		return err
	}
	if resourcesIsModified {
		// TODO: support resource change
		klog.Errorln("resource changes is not supported yet")
		return nil
	}
	// check zone number modified
	zoneScaleStatus, err := ctrl.ZoneNumberIsModified()
	if err != nil {
		return err
	}
	switch zoneScaleStatus {
	case observerconst.ZoneScaleUP:
		err = ctrl.OBZoneScaleUP(statefulApp)
	case observerconst.ZoneScaleDown:
		err = ctrl.OBZoneScaleDown(statefulApp)
	case observerconst.Maintain:
		err = ctrl.OBServerCoordinator(statefulApp)
	}
	return err
}
