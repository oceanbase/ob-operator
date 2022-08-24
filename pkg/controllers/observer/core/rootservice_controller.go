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

type RootServiceCtrl struct {
	OBClusterCtrl OBClusterCtrl
}

type RootServiceCtrlOperator interface {
	CreateRootService() (cloudv1.RootService, error)
	GetRootServiceByName(namespace, name string) (cloudv1.RootService, error)
	UpdateRootServiceStatus(rs cloudv1.RootService) error
	DeleteRootService(rs cloudv1.RootService) error
}

func NewRootServiceCtrl(obClusterCtrl *OBClusterCtrl) RootServiceCtrlOperator {
	return &RootServiceCtrl{
		OBClusterCtrl: *obClusterCtrl,
	}
}

func (ctrl *RootServiceCtrl) CreateRootService() (cloudv1.RootService, error) {
	rootService := converter.GenerateRootServiceObject(ctrl.OBClusterCtrl.OBCluster)
	rootServiceExecuter := resource.NewRootServiceResource(ctrl.OBClusterCtrl.Resource)
	err := rootServiceExecuter.Create(context.TODO(), rootService)
	if err != nil {
		return rootService, err
	}
	return rootService, nil
}

func (ctrl *RootServiceCtrl) GetRootServiceByName(namespace, name string) (cloudv1.RootService, error) {
	rootServiceExecuter := resource.NewRootServiceResource(ctrl.OBClusterCtrl.Resource)
	rs, err := rootServiceExecuter.Get(context.TODO(), namespace, name)
	if err != nil {
		return rs.(cloudv1.RootService), err
	}
	return rs.(cloudv1.RootService), nil
}

func (ctrl *RootServiceCtrl) UpdateRootServiceStatus(rs cloudv1.RootService) error {
	rootServiceExecuter := resource.NewRootServiceResource(ctrl.OBClusterCtrl.Resource)
	err := rootServiceExecuter.UpdateStatus(context.TODO(), rs)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *RootServiceCtrl) DeleteRootService(rs cloudv1.RootService) error {
	rootServiceExecuter := resource.NewRootServiceResource(ctrl.OBClusterCtrl.Resource)
	err := rootServiceExecuter.Delete(context.TODO(), rs)
	if err != nil {
		return err
	}
	return nil
}
