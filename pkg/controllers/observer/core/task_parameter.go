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
	"k8s.io/klog/v2"
)

func (ctrl *OBClusterCtrl) CheckAndSetParameters() error {
    sqlOperator, err := ctrl.GetSqlOperator()
    if err != nil {
        return errors.Wrap(err, "get sql operator when check and set parameter")
    }
	for _, cluster := range ctrl.OBCluster.Spec.Topology {
		for _, parameter := range cluster.Parameters {

			currentParameters := sqlOperator.GetParameter(parameter.Name)
			match := true
			for _, currentParameter := range currentParameters {
				if currentParameter.EditLevel == "READONLY" {
					klog.Infof("parameter %s is readonly, skip", parameter.Name)
					break
				} else if currentParameter.Value != parameter.Value {
					klog.Infof("found parameter %s with value %s did't match with config %s", parameter.Name, currentParameter.Value, parameter.Value)
					if currentParameter.EditLevel == "STATIC_EFFECTIVE" {
						klog.Infof("parameter %s is static effective, need restart after set parameter value", parameter.Name)
					}
					match = false
					break
				}
			}

			if !match {
				klog.Infof("set parameter %s = %s", parameter.Name, parameter.Value)
				err = sqlOperator.SetParameter(parameter.Name, parameter.Value)
				if err != nil {
					return err
				}
			}

		}
	}
	return nil
}
