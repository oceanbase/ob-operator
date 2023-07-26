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
	"time"

	"github.com/pkg/errors"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/cable"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
	statefulappCore "github.com/oceanbase/ob-operator/pkg/controllers/statefulapp/const"
	"github.com/oceanbase/ob-operator/pkg/util"
)

func (ctrl *OBClusterCtrl) IsNewCluster(statefulApp cloudv1.StatefulApp) (*cloudv1.StatefulApp, bool) {
	var err error
	for _, cluster := range ctrl.OBCluster.Spec.Topology {
		if cluster.Cluster == myconfig.ClusterName {
			statefulAppCtrl := NewStatefulAppCtrl(ctrl, statefulApp)
			// TODO: check owner
			statefulAppName := converter.GenerateStatefulAppName(ctrl.OBCluster.Name)
			statefulApp, err = statefulAppCtrl.GetStatefulAppByName(statefulAppName)
			if err != nil {
				if kubeerrors.IsNotFound(err) {
					return &statefulApp, true
				}
			}
		}
	}
	return &statefulApp, false
}

func (ctrl *OBClusterCtrl) NewCluster(statefulApp cloudv1.StatefulApp) error {
	// create StatefulApp
	statefulAppCtrl := NewStatefulAppCtrl(ctrl, statefulApp)
	statefulApp, err := statefulAppCtrl.CreateStatefulApp()
	if err != nil {
		return err
	}

	// create RootService
	rsCtrl := NewRootServiceCtrl(ctrl)
	_, err = rsCtrl.CreateRootService()
	if err != nil {
		return err
	}

	// create OBZone
	obZoneCtrl := NewOBZoneCtrl(ctrl)
	_, err = obZoneCtrl.CreateOBZone()
	if err != nil {
		return err
	}

	// update status
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ResourcePrepareing, "", "")
}

func (ctrl *OBClusterCtrl) ResourcePrepareingEffectorForBootstrap(statefulApp cloudv1.StatefulApp) error {
	var err error
	if statefulApp.Status.ClusterStatus == statefulappCore.Ready {
		// update status
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ResourceReady, "", "")
	} else {
		// time out, delete StatefulApp
		_, err = ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
		if err != nil {
			return err
		}
	}
	return err
}

// ResourceReadyEffectorForBootstrap will get iplsit by zone, and send iplsit to cable server, cable will start ob server
func (ctrl *OBClusterCtrl) ResourceReadyEffectorForBootstrap(statefulApp cloudv1.StatefulApp) error {
	// time out, delete StatefulApp
	status, err := ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
	if err != nil {
		return err
	}

	if !status {
		// check StatefulApp status
		err = ctrl.StatefulAppStatusCheckForBootstrap(statefulApp)
		if err != nil {
			return err
		}

		subsets := statefulApp.Status.Subsets

		// check cable status
		err = cable.CableStatusCheck(subsets)
		if err != nil {
			// TODO: need to check, why the cable get ip failed
			klog.Errorln("cable get ip failed, ", "OBCluster status ", util.CovertToJSON(ctrl.OBCluster.Status), "StatefulApp status", util.CovertToJSON(statefulApp.Status))
			return err
		}

		// generate rsList
		rsList := cable.GenerateRSListFromSubset(subsets)

		version, err := ctrl.GetCurrentVersion(statefulApp)
		if err != nil {
			klog.Errorln("bootstrap server get Version failed")
			version = observerconst.OBClusterV3
		}
		// start observer
		go cable.OBServerStart(ctrl.OBCluster, subsets, rsList, version)

		// update status
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.OBServerPrepareing, "", "")
	}

	return nil
}

func (ctrl *OBClusterCtrl) OBServerPrepareingEffectorForBootstrap(statefulApp cloudv1.StatefulApp) error {
	// time out, delete StatefulApp
	status, err := ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
	if err != nil {
		return err
	}

	if !status {
		// check StatefulApp status
		err = ctrl.StatefulAppStatusCheckForBootstrap(statefulApp)
		if err != nil {
			return err
		}

		subsets := statefulApp.Status.Subsets

		// check observer status
		err = cable.OBServerStatusCheck(ctrl.OBCluster.Name, subsets)
		if err != nil {
			return nil
		}

		// update status
		return ctrl.UpdateOBClusterAndZoneStatus(observerconst.OBServerReady, "", "")
	}

	return nil
}

func (ctrl *OBClusterCtrl) OBServerReadyEffectorForBootstrap(statefulApp cloudv1.StatefulApp) error {
	// time out, delete StatefulApp
	status, err := ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
	if err != nil {
		return err
	}

	if !status {
		// check StatefulApp status
		err = ctrl.StatefulAppStatusCheckForBootstrap(statefulApp)
		if err != nil {
			return err
		}

		subsets := statefulApp.Status.Subsets

		// make obcluster bootstrap args
		obclusterBootstrapArgs, err := cable.GenerateOBClusterBootstrapArgs(subsets)
		if err != nil {
			// delete StatefulApp
			statefulAppCtrl := NewStatefulAppCtrl(ctrl, statefulApp)
			err = statefulAppCtrl.DeleteStatefulApp()
			if err != nil {
				return err
			}
			klog.Errorln("generate OBCluster bootstrap args failed, delete StatefulApp", statefulApp.Name, "to recreate")
		}

		// update status
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.OBClusterBootstraping, "", "")
		if err != nil {
			return err
		}

		// run bootstrap SQL
		go ctrl.BootstrapForOB(statefulApp, obclusterBootstrapArgs)
	}

	return nil
}

func (ctrl *OBClusterCtrl) BootstrapForOB(statefulApp cloudv1.StatefulApp, SQL string) error {
	var sqlOperator *sql.SqlOperator
	var err error
	for i := 0; i < 60; i++ {
		sqlOperator, err = ctrl.GetSqlOperatorFromStatefulApp(statefulApp)
		if err != nil {
			klog.Errorf("get sql operator for bootstrap failed")
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	sqlOperator.ConnectProperties.Timeout = 600
	err = sqlOperator.BootstrapForOB(SQL)
	if err != nil {
		klog.Errorf("bootstrap failed %v", err)
		return errors.New("bootstrap ob failed")
	}

	klog.Infoln("OBCluster", ctrl.OBCluster.Name, "run bootstrap sql finish")
	_ = ctrl.UpdateOBClusterStatusBootstrapReady()
	return nil
}

func (ctrl *OBClusterCtrl) UpdateOBClusterStatusBootstrapReady() error {
	// update status
	return ctrl.UpdateOBClusterAndZoneStatus(observerconst.OBClusterBootstrapReady, "", "")
}

func (ctrl *OBClusterCtrl) OBClusterBootstraping(statefulApp cloudv1.StatefulApp) error {
	// time out, delete StatefulApp
	_, err := ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
	return err
}

func (ctrl *OBClusterCtrl) OBClusterBootstrapReady(statefulApp cloudv1.StatefulApp) error {
	sqlOperator, err := ctrl.GetSqlOperatorFromStatefulApp(statefulApp)
	if err != nil {
		return errors.Wrap(err, "get sql operator when bootstrap")
	}
	// time out, delete StatefulApp
	status, err := ctrl.CheckTimeoutAndKillForBootstrap(statefulApp)
	if err != nil {
		return err
	}

	if !status {
		subsets := statefulApp.Status.Subsets
		obServerList := sqlOperator.GetOBServer()

		obServerBootstrapSucceed := converter.IsAllOBServerActive(obServerList, ctrl.OBCluster.Spec.Topology)
		if obServerBootstrapSucceed {
			// update OBServer Pod Readiness
			err = cable.CableReadinessUpdate(subsets)
			if err != nil {
				return err
			}
			// update status
			return ctrl.UpdateOBClusterAndZoneStatus(observerconst.OBClusterReady, "", "")
		}

		klog.Infoln("wait for OBCluster", ctrl.OBCluster.Name, "Bootstraping finish")
	}

	return nil
}

func (ctrl *OBClusterCtrl) StatefulAppStatusCheckForBootstrap(statefulApp cloudv1.StatefulApp) error {
	if statefulApp.Status.ClusterStatus != statefulappCore.Ready {
		// update status
		err := ctrl.UpdateOBClusterAndZoneStatus(observerconst.ResourcePrepareing, "", "")
		if err != nil {
			return err
		}
		klog.Infoln("StatefulApp is Prepareing, update OBCluster", ctrl.OBCluster.Name, "to", observerconst.ResourcePrepareing)
	}
	return nil
}

func (ctrl *OBClusterCtrl) CheckTimeoutAndKillForBootstrap(statefulApp cloudv1.StatefulApp) (bool, error) {
	creationTimestamp := statefulApp.CreationTimestamp.Unix()
	nowTimestamp := time.Now().Unix()
	timeoutSeconds := ctrl.OBCluster.Spec.BootstrapTimeoutSeconds
	if timeoutSeconds == 0 {
		timeoutSeconds = observerconst.BootstrapTimeout
	}
	if nowTimestamp-creationTimestamp > timeoutSeconds {
		klog.Infoln("OBCluster Bootstraping timeout, now time", nowTimestamp, "create time", creationTimestamp)
		statefulAppCtrl := NewStatefulAppCtrl(ctrl, statefulApp)
		// delete StatefulApp
		err := statefulAppCtrl.DeleteStatefulApp()
		if err != nil {
			return true, err
		}
		// delete RootService
		rsName := converter.GenerateRootServiceName(ctrl.OBCluster.Name)
		rsCtrl := NewRootServiceCtrl(ctrl)
		rsCurrent, _ := rsCtrl.GetRootServiceByName(ctrl.OBCluster.Namespace, rsName)
		_ = rsCtrl.DeleteRootService(rsCurrent)
		// delete OBZone
		obZoneName := converter.GenerateOBZoneName(ctrl.OBCluster.Name)
		obZoneCtrl := NewOBZoneCtrl(ctrl)
		obZoneCurrent, _ := obZoneCtrl.GetOBZoneByName(ctrl.OBCluster.Namespace, obZoneName)
		_ = obZoneCtrl.DeleteOBZone(obZoneCurrent)
		// update status
		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, "", "")
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}
