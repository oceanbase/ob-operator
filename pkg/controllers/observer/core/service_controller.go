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

	corev1 "k8s.io/api/core/v1"

	"github.com/oceanbase/ob-operator/pkg/controllers/observer/core/converter"
	"github.com/oceanbase/ob-operator/pkg/infrastructure/kube/resource"
)

func (ctrl *OBClusterCtrl) CreateService(statefulAppName string) error {
	service := converter.GenerateServiceObject(ctrl.OBCluster, statefulAppName)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	err := serviceExecuter.Create(context.TODO(), service)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) CreateServiceForPrometheus(statefulAppName string) error {
	service := converter.GenerateServiceObjectForPrometheus(ctrl.OBCluster, statefulAppName)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	err := serviceExecuter.Create(context.TODO(), service)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *OBClusterCtrl) GetServiceByName(namespace, name string) (corev1.Service, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return svc.(corev1.Service), err
	}
	return svc.(corev1.Service), nil
}

func (ctrl *OBClusterCtrl) GetServiceClusterIPByName(namespace, name string) (string, error) {
	svcName := converter.GenerateServiceName(name)
	serviceExecuter := resource.NewServiceResource(ctrl.Resource)
	svc, err := serviceExecuter.Get(context.TODO(), namespace, svcName)
	if err != nil {
		return "", err
	}
	return svc.(corev1.Service).Spec.ClusterIP, nil
}
