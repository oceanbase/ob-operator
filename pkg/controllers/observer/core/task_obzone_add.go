package core

import (
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func (ctrl *OBClusterCtrl) AddAndStartOBZone(clusterIP string) error {
	clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)
	expectedOBZoneList := ctrl.OBCluster.Spec.Topology[0].Zone

	isExist := false
	for _, zone := range expectedOBZoneList {
		for _, readyZone := range clusterStatus.Zone {
			// 说明该zone已经ready
			if zone.Name == readyZone.Name {
				isExist = true
			}
		}
		if !isExist {
			// add zone
			err := sql.AddZone(clusterIP, zone.Name)
			if err != nil {
				return nil
			}

			//start zone
			err = sql.StartZone(clusterIP, zone.Name)
			if err != nil {
				return err
			}
		}
		isExist = false
	}
	return nil
}
