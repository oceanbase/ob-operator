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
	"runtime"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
)

func (ctrl *OBClusterCtrl) DelOBServer(clusterIP, zoneName, podIP string) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	for _, zone := range clusterStatus.Zone {
		if zone.ZoneStatus == observerconst.OBServerDel {
			// judge timeout
			lastTransitionTimestamp := clusterStatus.LastTransitionTime.Unix()
			nowTimestamp := time.Now().Unix()
			if nowTimestamp-lastTransitionTimestamp > observerconst.DelServerTimeout {
				klog.Infoln("OBCluster delete server timeout, now time", nowTimestamp, "create time", lastTransitionTimestamp)
				return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
			}
			return nil
		}
	}
	for _, zone := range clusterStatus.Zone {
		// an IP is already in the process of being deleted
		// execute serially, one IP at a time
		if zone.Name == zoneName && zone.ZoneStatus != observerconst.OBServerDel {
			// del server
			go ctrl.DelOBServerExecuter(clusterIP, zoneName, podIP)
			// update status
			return ctrl.UpdateOBClusterAndZoneStatus(observerconst.ScaleDown, zoneName, observerconst.OBServerDel)
		}
	}
	return nil
}

func (ctrl *OBClusterCtrl) DelOBServerExecuter(clusterIP, zoneName, podIP string) error {
	klog.Infoln("begin delete OBServer", zoneName, podIP)

    sqlOperator, err := ctrl.GetSqlOperator()
    if err != nil {
        return errors.Wrap(err, "get sql operator when create user for operation")
    }

	// update server server_permanent_offline_time
	err = sqlOperator.SetServerOfflineTime(20)
	if err != nil {
		klog.Errorln("set server_permanent_offline_time error", zoneName, podIP)
		runtime.Goexit()
	}

	// delete server
	err = sqlOperator.DelServer(podIP)
	if err != nil {
		klog.Errorln("delete server error", zoneName, podIP)
		runtime.Goexit()
	}

	// check delete finish
	err = ctrl.TickerRSJobStatusCheck(clusterIP, podIP)
	if err != nil {
		klog.Errorln("check rs job status error", zoneName, podIP)
		runtime.Goexit()
	}

	// check server is not in db
	status := ctrl.OBServerDeletedCheck(podIP)
	if status {
		klog.Infoln("delete OBServer finish", zoneName, podIP)

		// update status
		_ = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ClusterReady, zoneName, observerconst.OBZoneReady)
	}
    return nil
}

func (ctrl *OBClusterCtrl) IsRSJobSuccess(podIP string) (bool, error) {
    sqlOperator, err := ctrl.GetSqlOperator()
    if err != nil {
        return false, errors.Wrap(err, "get sql operator when check rs job")
    }

	rsJobStatusList := sqlOperator.GetRSJobStatus(podIP)
	if len(rsJobStatusList) == 0 {
		return false, errors.New("get rs job status faild")
	}
	lastJob := rsJobStatusList[len(rsJobStatusList)-1]
	// job status is not SUCCESS
	if lastJob.JobStatus != observerconst.RSJobStatusSuccess {
		return false, nil
	}
	return true, nil
}

func (ctrl *OBClusterCtrl) TickerRSJobStatusCheck(clusterIP, podIP string) error {
    sqlOperator, err := ctrl.GetSqlOperator()
    if err != nil {
        return errors.Wrap(err, "get sql operator when create user for operation")
    }

	tick := time.Tick(observerconst.TickPeriodForRSJobStatusCheck)
	var num int
	for {
		select {
		case <-tick:
			if num > observerconst.TickNumForRSJobStatusCheck {
				return errors.New("observer starting timeout")
			}
			num = num + 1
			status, err := ctrl.IsRSJobSuccess(podIP)
			if err == nil {
				return err
			}
			if status {
				err = sqlOperator.SetServerOfflineTime(3600)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
}

func (ctrl *OBClusterCtrl) OBServerDeletedCheck(podIP string) bool {
    sqlOperator, err := ctrl.GetSqlOperator()
    if err != nil {
        return false
    }
	obServerList := sqlOperator.GetOBServer()
	for _, obServer := range obServerList {
		if obServer.SvrIP == podIP {
			return false
		}
	}
	return true
}
