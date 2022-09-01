package core

import (
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func (ctrl *OBClusterCtrl) AddAndStartOBZone(clusterIP string) error {
	// clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)

	clusterSpec := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	expectedOBZoneList := clusterSpec.Zone

	obZoneList := sql.GetOBZone(clusterIP)

	isExist := false
	for _, zone := range expectedOBZoneList {
		for _, existZone := range obZoneList {
			// 说明该zone已经存在
			if zone.Name == existZone.Zone {
				isExist = true
			}
		}
		if !isExist {
			// add zone
			err := sql.AddZone(clusterIP, zone.Name)
			if err != nil {
				return err
			}
		}
		isExist = false
	}

	isReady := false
	for _, zone := range expectedOBZoneList {
		for _, existZone := range obZoneList {
			// 说明该zone已经ready
			if zone.Name == existZone.Zone && existZone.Info == observerconst.OBZoneActive {
				isReady = true
			}
		}
		if !isReady {
			// start zone
			err := sql.StartZone(clusterIP, zone.Name)
			if err != nil {
				return err
			}
		}
		isReady = false
	}
	return nil
}
