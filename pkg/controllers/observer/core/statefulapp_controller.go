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
	myconfig "github.com/oceanbase/ob-operator/pkg/config"
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

type StatefulAppCtrl struct {
	OBClusterCtrl OBClusterCtrl
	StatefulApp   cloudv1.StatefulApp
}

type StatefulAppCtrlOperator interface {
	CreateStatefulApp() (cloudv1.StatefulApp, error)
	GetStatefulAppByName(name string) (cloudv1.StatefulApp, error)
	UpdateStatefulApp() error
	DeleteStatefulApp() error
}

func NewStatefulAppCtrl(obClusterCtrl *OBClusterCtrl, statefulApp cloudv1.StatefulApp) StatefulAppCtrlOperator {
	return &StatefulAppCtrl{
		OBClusterCtrl: *obClusterCtrl,
		StatefulApp:   statefulApp,
	}
}

func (ctrl *StatefulAppCtrl) CreateStatefulApp() (cloudv1.StatefulApp, error) {
	var cluster cloudv1.Cluster
	for _, cluster = range ctrl.OBClusterCtrl.OBCluster.Spec.Topology {
		if cluster.Cluster == myconfig.ClusterName {
			break
		}
	}
	statefulApp := converter.GenerateStatefulAppObject(cluster, ctrl.OBClusterCtrl.OBCluster)
	statefulAppExecuter := resource.NewStatefulAppResource(ctrl.OBClusterCtrl.Resource)
	err := statefulAppExecuter.Create(context.TODO(), statefulApp)
	if err != nil {
		return statefulApp, err
	}
	return statefulApp, nil
}

func (ctrl *StatefulAppCtrl) GetStatefulAppByName(name string) (cloudv1.StatefulApp, error) {
	statefulAppExecuter := resource.NewStatefulAppResource(ctrl.OBClusterCtrl.Resource)
	statefulApp, err := statefulAppExecuter.Get(context.TODO(), ctrl.OBClusterCtrl.OBCluster.Namespace, name)
	if err != nil {
		return cloudv1.StatefulApp{}, err
	}
	return statefulApp.(cloudv1.StatefulApp), nil
}

func (ctrl *StatefulAppCtrl) UpdateStatefulApp() error {
	statefulAppExecuter := resource.NewStatefulAppResource(ctrl.OBClusterCtrl.Resource)
	err := statefulAppExecuter.Update(context.TODO(), ctrl.StatefulApp)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *StatefulAppCtrl) DeleteStatefulApp() error {
	statefulAppExecuter := resource.NewStatefulAppResource(ctrl.OBClusterCtrl.Resource)
	err := statefulAppExecuter.Delete(context.TODO(), ctrl.StatefulApp)
	if err != nil {
		return err
	}
	return nil
}
