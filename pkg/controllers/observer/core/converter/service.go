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

package converter

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	cloudv1 "github.com/oceanbase/ob-operator/apis/cloud/v1"
	observerconst "github.com/oceanbase/ob-operator/pkg/controllers/observer/const"
	observerutil "github.com/oceanbase/ob-operator/pkg/controllers/observer/core/util"
)

func GenerateServiceName(name string) string {
	statefulAppName := fmt.Sprintf("svc-%s", name)
	return statefulAppName
}

func GenerateServiceSpec(statefulAppName string) corev1.ServiceSpec {
	ports := make([]corev1.ServicePort, 0)
	var servicePort corev1.ServicePort
	servicePort.Name = observerconst.MysqlPortName
	servicePort.Port = observerconst.MysqlPort
	servicePort.TargetPort = intstr.FromInt(observerconst.MysqlPort)
	servicePort.Protocol = corev1.ProtocolTCP
	ports = append(ports, servicePort)

	selector := make(map[string]string)
	selector["app"] = statefulAppName

	var res corev1.ServiceSpec
	res.Ports = ports
	res.Selector = selector
	res.Type = corev1.ServiceTypeClusterIP
	return res
}

func GenerateServiceNameForPrometheus(name string) string {
	statefulAppName := fmt.Sprintf("svc-monitor-%s", name)
	return statefulAppName
}

func GenerateServiceSpecForPrometheus(statefulAppName string) corev1.ServiceSpec {
	ports := make([]corev1.ServicePort, 0)
	var servicePort corev1.ServicePort
	servicePort.Name = observerconst.MonagentPortName
	servicePort.Port = observerconst.MonagentPort
	servicePort.TargetPort = intstr.FromInt(observerconst.MonagentPort)
	servicePort.Protocol = corev1.ProtocolTCP
	ports = append(ports, servicePort)

	selector := make(map[string]string)
	selector["app"] = statefulAppName

	var res corev1.ServiceSpec
	res.Ports = ports
	res.Selector = selector
	res.Type = corev1.ServiceTypeNodePort
	return res
}

func GenerateServiceObject(obCluster cloudv1.OBCluster, statefulAppName string) corev1.Service {
	name := GenerateServiceName(obCluster.Name)
	objectMeta := observerutil.GenerateObjectMeta(obCluster, name)
	serviceSpec := GenerateServiceSpec(statefulAppName)
	service := corev1.Service{
		ObjectMeta: objectMeta,
		Spec:       serviceSpec,
	}
	return service
}

func GenerateServiceObjectForPrometheus(obCluster cloudv1.OBCluster, statefulAppName string) corev1.Service {
	name := GenerateServiceNameForPrometheus(obCluster.Name)
	objectMeta := observerutil.GenerateObjectMeta(obCluster, name)
	serviceSpec := GenerateServiceSpecForPrometheus(statefulAppName)
	service := corev1.Service{
		ObjectMeta: objectMeta,
		Spec:       serviceSpec,
	}
	return service
}
