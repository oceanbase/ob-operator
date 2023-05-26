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
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/cable"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func (ctrl *OBClusterCtrl) AsyncStartOBServer(clusterIP, zoneName, podIP string, statefulApp cloudv1.StatefulApp) error {
	go func() {
		ctrl.StartOBServer(clusterIP, zoneName, podIP, statefulApp)
		ctrl.WaitOBServerActive(clusterIP, zoneName, podIP, statefulApp)
	}()
	return nil
}

func (ctrl *OBClusterCtrl) StartOBServer(clusterIP, zoneName, podIP string, statefulApp cloudv1.StatefulApp) error {
	klog.Infoln("begin start OBServer", zoneName, podIP)

	// check cable status
	err := cable.CableStatusCheckExecuter(podIP)
	if err != nil {
		// kill pod
		_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
		_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
	}

	// get rs
	rsName := converter.GenerateRootServiceName(ctrl.OBCluster.Name)
	rsCtrl := NewRootServiceCtrl(ctrl)
	rsCurrent, err := rsCtrl.GetRootServiceByName(ctrl.OBCluster.Namespace, rsName)
	if err != nil {
		// kill pod
		_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
		_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
	}

	// update rsList first
	ctrl.UpdateRootServiceStatus(statefulApp)
	// generate rsList
	rsList := cable.GenerateRSListFromRootServiceStatus(rsCurrent.Status.Topology)
	if len(rsList) == 0 {
		klog.Info("rs list is empty")
		return errors.New("rs list is empty")
	}
	// generate start args
	version, err := ctrl.GetCurrentVersion(statefulApp)
	if err != nil {
		klog.Errorln("add server get Version failed")
		version = observerconst.OBClusterV3
	}
	obServerStartArgs := cable.GenerateOBServerStartArgs(ctrl.OBCluster, zoneName, rsList, version)
	// check OBServer is already running, for OBServer Scale UP
	err = cable.OBServerStatusCheckExecuter(ctrl.OBCluster.Name, podIP)
	// nil is OBServer is already running
	if err != nil {
		cable.OBServerStartExecuter(podIP, obServerStartArgs)
		err = TickerOBServerStatusCheck(ctrl.OBCluster.Name, podIP)
		if err != nil {
			// kill pod
			_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
			_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
		}
	}

	return err
}

func (ctrl *OBClusterCtrl) AddOBServer(clusterIP, zoneName, podIP string, statefulApp cloudv1.StatefulApp) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	for _, zone := range clusterStatus.Zone {
		if zone.ZoneStatus == observerconst.OBServerAdd {
			// judge timeout
			lastTransitionTimestamp := clusterStatus.LastTransitionTime.Unix()
			nowTimestamp := time.Now().Unix()
			if nowTimestamp-lastTransitionTimestamp > observerconst.AddServerTimeout {
				klog.Infoln("add server timeout, need delete", zoneName, podIP)
				err := ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
				if err != nil {
					return err
				}
				return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
			}
			return nil
		}
	}
	for _, zone := range clusterStatus.Zone {
		// an IP is already in the process of being added
		// execute serially, one IP at a time
		if zone.Name == zoneName && zone.ZoneStatus != observerconst.OBServerAdd {
			// add server and update obagent config
			go func() {
				ctrl.AddOBServerExecuter(clusterIP, zoneName, podIP, statefulApp)
				ctrl.ReviseConfig(podIP, zoneName)
			}()
			// update status
			return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ScaleUP, zoneName, observerconst.OBServerAdd)
		}
	}
	return nil
}

func (ctrl *OBClusterCtrl) AddOBServerExecuter(clusterIP, zoneName, podIP string, statefulApp cloudv1.StatefulApp) {
	klog.Infoln("begin add OBServer", zoneName, podIP)

	ctrl.StartOBServer(clusterIP, zoneName, podIP, statefulApp)

	sqlOperator, err := ctrl.GetSqlOperator()
	if err == nil {
		// add server
		err = sqlOperator.AddServer(zoneName, podIP)
		if err != nil {
			// kill pod
			_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
			_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
		} else {
			klog.Infoln("add OBServer finish", zoneName, podIP)
			ctrl.WaitOBServerActive(clusterIP, zoneName, podIP, statefulApp)
		}
	}

	// update status
	_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
}

func (ctrl *OBClusterCtrl) WaitOBServerActive(clusterIP, zoneName, podIP string, statefulApp cloudv1.StatefulApp) error {
	klog.Infof("wait observer %s ready", podIP)
	err := ctrl.TickerOBServerStatusCheckFromDB(clusterIP, podIP, statefulApp)
	if err != nil {
		// kill pod
		klog.Infof("observer %s still not ready", podIP)
		_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
		_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
	}

	klog.Infof("observer %s is ready, update pod status", podIP)
	// update OBServer Pod Readiness
	err = cable.CableReadinessUpdateExecuter(podIP)
	if err != nil {
		// kill pod
		klog.Infof("update pod %s status failed", podIP)
		_ = ctrl.DelPodFromStatefulAppByIP(zoneName, podIP, statefulApp)
		_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
	}
	return err
}

func TickerOBServerStatusCheck(clusterName, podIP string) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	var err error
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("observer starting timeout")
			}
			num = num + 1
			err = cable.OBServerStatusCheckExecuter(clusterName, podIP)
			if err == nil {
				return err
			}
		}
	}

}

func (ctrl *OBClusterCtrl) TickerOBServerStatusCheckFromDB(clusterIP string, podIP string, args ...cloudv1.StatefulApp) error {
	tick := time.Tick(observerconst.TickPeriodForOBServerStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForOBServerStatusCheck {
				return errors.New("observer starting timeout")
			}
			num = num + 1
			var sqlOperator *sql.SqlOperator
			var err error
			if args != nil {
				sqlOperator, err = ctrl.GetSqlOperatorFromStatefulApp(args[0])
			} else {
				sqlOperator, err = ctrl.GetSqlOperator()
			}
			if err == nil {
				obServerList := sqlOperator.GetOBServer()
				for _, obServer := range obServerList {
					if obServer.SvrIP == podIP {
						if obServer.Status == observerconst.OBServerActive && obServer.StartServiceTime > 0 {
							return nil
						}
					}
				}
			}
		}
	}
}
