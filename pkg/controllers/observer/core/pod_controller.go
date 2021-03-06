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

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *OBClusterCtrl) DelPodFromStatefulAppByIP(zoneName, podIP string, statefulApp cloudv1.StatefulApp) error {
	subsetStatus := converter.GetSubsetStatusFromStatefulApp(zoneName, statefulApp)
	for _, pod := range subsetStatus.Pods {
		if pod.PodIP == podIP {
			podExecuter := resource.NewPodResource(ctrl.Resource)
			podObject, err := podExecuter.Get(context.TODO(), ctrl.OBCluster.Namespace, pod.Name)
			if err != nil {
				return err
			}
			err = podExecuter.Delete(context.TODO(), podObject)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
