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
	"github.com/pkg/errors"

	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
)

func (ctrl *OBClusterCtrl) AddAndStartOBZone(clusterIP string) error {
	// clusterStatus := converter.GetClusterStatusFromOBTopologyStatus(ctrl.OBCluster.Status.Topology)

	clusterSpec := converter.GetClusterSpecFromOBTopology(ctrl.OBCluster.Spec.Topology)
	expectedOBZoneList := clusterSpec.Zone

	sqlOperator, err := ctrl.GetSqlOperator()
	if err != nil {
		return errors.Wrap(err, "get sql operator when add and start obzone")
	}

	obZoneList := sqlOperator.GetOBZone()

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
			err := sqlOperator.AddZone(zone.Name)
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
			err := sqlOperator.StartZone(zone.Name)
			if err != nil {
				return err
			}
		}
		isReady = false
	}
	return nil
}
