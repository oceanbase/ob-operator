package core

import (
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) AddOBZone(clusterIP, obZoneName string) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)

	klog.Infoln("AddOBZone: clusterStatus", clusterStatus)
	expectedOBZoneList := ctrl.OBCluster.Spec.Topology
	klog.Infoln("AddOBZone: expectedOBZoneList ", expectedOBZoneList)

	for _, zone := range expectedOBZoneList {
		klog.Infoln("AddOBZone: item ", zone)
	}

	return nil
}

