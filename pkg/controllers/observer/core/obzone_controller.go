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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
	"k8s.io/klog/v2"
)

type OBZoneCtrl struct {
	OBClusterCtrl OBClusterCtrl
}

type OBZoneCtrlOperator interface {
	CreateOBZone() (cloudv1.OBZone, error)
	GetOBZoneByName(namespace, name string) (cloudv1.OBZone, error)
	UpdateOBZone(zone cloudv1.OBZone) error
	UpdateOBZoneStatus(zone cloudv1.OBZone) error
	DeleteOBZone(zone cloudv1.OBZone) error
}

func NewOBZoneCtrl(obClusterCtrl *OBClusterCtrl) OBZoneCtrlOperator {
	return &OBZoneCtrl{
		OBClusterCtrl: *obClusterCtrl,
	}
}

func (ctrl *OBZoneCtrl) CreateOBZone() (cloudv1.OBZone, error) {
	obZone := converter.GenerateOBZoneObject(ctrl.OBClusterCtrl.OBCluster)
	obZoneExecuter := resource.NewOBZoneResource(ctrl.OBClusterCtrl.Resource)
	err := obZoneExecuter.Create(context.TODO(), obZone)
	if err != nil {
		return obZone, err
	}
	return obZone, nil
}

func (ctrl *OBZoneCtrl) GetOBZoneByName(namespace, name string) (cloudv1.OBZone, error) {
	obZoneExecuter := resource.NewOBZoneResource(ctrl.OBClusterCtrl.Resource)
	rs, err := obZoneExecuter.Get(context.TODO(), namespace, name)
	if err != nil {
		return rs.(cloudv1.OBZone), err
	}
	return rs.(cloudv1.OBZone), nil
}

func (ctrl *OBZoneCtrl) UpdateOBZone(zone cloudv1.OBZone) error {
	obZoneExecuter := resource.NewOBZoneResource(ctrl.OBClusterCtrl.Resource)
	err := obZoneExecuter.Update(context.TODO(), zone)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBZoneCtrl) UpdateOBZoneStatus(zone cloudv1.OBZone) error {
	obZoneExecuter := resource.NewOBZoneResource(ctrl.OBClusterCtrl.Resource)
	err := obZoneExecuter.UpdateStatus(context.TODO(), zone)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBZoneCtrl) DeleteOBZone(zone cloudv1.OBZone) error {
	obZoneExecuter := resource.NewOBZoneResource(ctrl.OBClusterCtrl.Resource)
	err := obZoneExecuter.Delete(context.TODO(), zone)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) OBZoneScaleUP(statefulApp cloudv1.StatefulApp) error {
	//generate new StatefulApp for new Zone
	klog.Infoln("----------------------------OBZoneScaleUP----------------------------")

	// generate new OBZone for new ObZone
	obZoneName := converter.GenerateOBZoneName(ctrl.OBCluster.Name)
	obZoneCtrl := NewOBZoneCtrl(ctrl)
	obZoneCurrent, err := obZoneCtrl.GetOBZoneByName(ctrl.OBCluster.Namespace, obZoneName)
	if err != nil {
		return err
	}
	newOBZone := converter.UpdateOBZoneSpec(obZoneCurrent, ctrl.OBCluster.Spec.Topology)

	// update OBZone replica
	err = obZoneCtrl.UpdateOBZone(newOBZone)
	if err != nil {
		return err
	}

	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: GetServiceClusterIPByName err ", err)
		return err
	}

	// add and start ob zone
	err = ctrl.AddAndStartOBZone(clusterIP)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: AddAndStartOBZone err ", err)
		return err
	}

	err = ctrl.UpdateOBZoneStatus(statefulApp)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: UpdateOBZoneStatus err ", err)
	}

	err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ZoneScaleUP, "", "")
	if err != nil {
		klog.Infoln("OBZoneScaleUP: UpdateOBClusterAndZoneStatus err ", err)
		return err
	}

	return ctrl.UpdateOBZoneStatusFromDB(clusterIP)

}

func (ctrl *OBClusterCtrl) GetInfoForDelZone(clusterIP string, clusterSpec cloudv1.Cluster, statefulApp cloudv1.StatefulApp) (error, string) {
	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when get info for del zone"), ""
	}

	obZoneList := sqlOperator.GetOBZone()
	if len(obZoneList) == 0 {
		return errors.New(observerconst.DataBaseError), ""
	}
	zoneNodeMap := converter.GenerateZoneNodeMapByOBZoneList(obZoneList)

	for _, obZone := range obZoneList {
		zoneSpec := converter.GetZoneSpecFromClusterSpec(obZone.Zone, clusterSpec)
		if zoneNodeMap[obZone.Zone] != nil && zoneSpec.Name == "" {
			return nil, obZone.Zone
		}
	}

	return errors.New("none zone need del"), ""
}

func (ctrl *OBClusterCtrl) OBZoneScaleDown(statefulApp cloudv1.StatefulApp) error {
	klog.Infoln("----------------------------OBZoneScaleDown----------------------------")

	obZoneName := converter.GenerateOBZoneName(ctrl.OBCluster.Name)
	obZoneCtrl := NewOBZoneCtrl(ctrl)

	obZoneCurrent, err := obZoneCtrl.GetOBZoneByName(ctrl.OBCluster.Namespace, obZoneName)
	if err != nil {
		return err
	}
	newOBZone := converter.UpdateOBZoneSpec(obZoneCurrent, ctrl.OBCluster.Spec.Topology)

	// update OBZone replica
	err = obZoneCtrl.UpdateOBZone(newOBZone)
	if err != nil {
		return err
	}

	clusterIP, err := ctrl.GetServiceClusterIPByName(ctrl.OBCluster.Namespace, ctrl.OBCluster.Name)
	if err != nil {
		klog.Infoln("OBZoneScaleDown: GetServiceClusterIPByName err ", err)
		return err
	}

	// get info for delete obzone
	clusterSpec := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	err, zoneName := ctrl.GetInfoForDelZone(clusterIP, clusterSpec, statefulApp)
	// nil : need to delete zone
	if err == nil {
		// del obzone
		klog.Infoln("need to delete obzone ")
		err = ctrl.DeleteOBZone(clusterIP, zoneName, statefulApp)
		if err != nil {
			klog.Infoln("OBZoneScaleDown: Delete OBZone err ", err)
			return err
		}

		err = ctrl.UpdateOBZoneStatus(statefulApp)
		if err != nil {
			klog.Infoln("OBZoneScaleDown: UpdateOBZoneStatus err ", err)
		}

		err = ctrl.UpdateOBClusterAndZoneStatus(observerconst.ZoneScaleDown, "", "")
		if err != nil {
			klog.Infoln("OBZoneScaleDown: UpdateOBClusterAndZoneStatus err ", err)
			return err
		}
	}

	return ctrl.UpdateOBZoneStatusFromDB(clusterIP)

}
