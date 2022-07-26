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
	"k8s.io/klog/v2"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
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

func (ctrl *OBClusterCtrl) OBZoneScaleUP(statefulApp cloudv1.StatefulApp, status string) error {
	//generate new StatefulApp for new Zone
	klog.Infoln("----------------------------UpdateOBZoneTopology----------------------------")
	// create StatefulApp
	newStatefulApp := converter.UpdateZoneForStatefulApp(ctrl.OBCluster.Spec.Topology, statefulApp)

	statefulAppCtrl := NewStatefulAppCtrl(ctrl, newStatefulApp)
	err := statefulAppCtrl.UpdateStatefulApp()
	if err != nil {
		return err
	}

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

	err = ctrl.UpdateOBZoneStatus(statefulApp)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: UpdateOBZoneStatus err ", err)
	}

	err = ctrl.UpdateOBClusterAndZoneStatus(status, "", "")
	if err != nil {
		klog.Infoln("OBZoneScaleUP: UpdateOBClusterAndZoneStatus err ", err)
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

	err, zoneName, podIP := converter.GetInfoForAddServerByZone(clusterIP, statefulApp)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: GetInfoForAddServerByZone err ", err)
		return err
	}

	err = ctrl.AddOBServer(clusterIP, zoneName, podIP, statefulApp)
	if err != nil {
		klog.Infoln("OBZoneScaleUP: AddOBServer err ", err)
		return err
	}
	return nil
}
