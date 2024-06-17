/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/constant"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
	"github.com/oceanbase/ob-operator/internal/dashboard/utils"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func getDeploymentStatus(deploy *appsv1.Deployment) string {
	if deploy.Status.Replicas == deploy.Status.AvailableReplicas {
		return "Running"
	}
	return "Pending"
}

func buildOBProxyOverview(deploy *appsv1.Deployment) *obproxy.OBProxyOverview {
	overview := &obproxy.OBProxyOverview{
		Name:      deploy.Name,
		Namespace: deploy.Namespace,
		OBCluster: obproxy.K8sObject{
			Namespace: deploy.Labels[LabelForNamespace],
			Name:      deploy.Labels[LabelForOBCluster],
		},
		ProxyClusterName: deploy.Labels[LabelProxyClusterName],
		Image:            deploy.Spec.Template.Spec.Containers[0].Image,
		Replicas:         *deploy.Spec.Replicas,
		Status:           getDeploymentStatus(deploy),
		CreationTime:     deploy.CreationTimestamp.Unix(),
		ServiceIP:        deploy.Annotations[AnnotationServiceIP],
	}
	return overview
}

func buildOBProxy(ctx context.Context, deploy *appsv1.Deployment) (*obproxy.OBProxy, error) {
	cm, err := getConfigMap(ctx, deploy.Namespace, deploy.Name)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get obproxy config map")
	}
	svc, err := getOBProxyService(ctx, deploy.Namespace, deploy.Name)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get obproxy service")
	}
	pods, err := client.GetClient().ClientSet.CoreV1().Pods(deploy.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: LabelOBProxy + "=" + deploy.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to list obproxy pods")
	}

	odp := &obproxy.OBProxy{
		OBProxyOverview: *buildOBProxyOverview(deploy),
		ProxySysSecret:  deploy.Annotations[AnnotationProxySysSecret],
		Service: response.K8sService{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: strings.Join(svc.Spec.ExternalIPs, ","),
			Ports:      []response.K8sServicePort{},
		},
		Resource: common.ResourceSpec{
			Cpu:      deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().Value(),
			MemoryGB: deploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().ScaledValue(resource.Giga),
		},
		Parameters: []common.KVPair{},
		Pods:       []response.K8sPodInfo{},
	}
	for _, port := range svc.Spec.Ports {
		odp.Service.Ports = append(odp.Service.Ports, response.K8sServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort.IntVal,
		})
	}
	for k, v := range cm.Data {
		odp.Parameters = append(odp.Parameters, common.KVPair{
			Key:   strings.ToLower(strings.ReplaceAll(k, envPrefix, "")),
			Value: v,
		})
	}
	// TODO: Move pods fetching to another function?
	for _, pod := range pods.Items {
		podInfo := response.K8sPodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			NodeName:   pod.Spec.NodeName,
			PodIP:      pod.Status.PodIP,
			Status:     string(pod.Status.Phase),
			Message:    pod.Status.Message,
			Reason:     pod.Status.Reason,
			StartTime:  pod.Status.StartTime.Format(time.DateTime),
			Containers: []response.ContainerInfo{},
		}
		for _, container := range pod.Spec.Containers {
			containerInfo := response.ContainerInfo{
				Name:  container.Name,
				Image: container.Image,
				Ports: []int32{},
				Requests: common.ResourceSpec{
					Cpu:      container.Resources.Requests.Cpu().Value(),
					MemoryGB: container.Resources.Requests.Memory().ScaledValue(resource.Giga),
				},
				Limits: common.ResourceSpec{
					Cpu:      container.Resources.Limits.Cpu().Value(),
					MemoryGB: container.Resources.Limits.Memory().ScaledValue(resource.Giga),
				},
			}
			if len(pod.Status.ContainerStatuses) > 0 {
				containerInfo.RestartCount = pod.Status.ContainerStatuses[0].RestartCount
				containerInfo.Ready = pod.Status.ContainerStatuses[0].Ready
				if pod.Status.ContainerStatuses[0].State.Running != nil {
					containerInfo.StartTime = pod.Status.ContainerStatuses[0].State.Running.StartedAt.Format(time.DateTime)
				}
			}
			for _, port := range container.Ports {
				containerInfo.Ports = append(containerInfo.Ports, port.ContainerPort)
			}
			podInfo.Containers = append(podInfo.Containers, containerInfo)
		}
		odp.Pods = append(odp.Pods, podInfo)
	}
	return odp, nil
}

type buildDeploymentParam struct {
	cm             *corev1.ConfigMap
	svc            *corev1.Service
	proxyRoSecret  *corev1.Secret
	proxySysSecret *corev1.Secret
	cluster        *v1alpha1.OBCluster
}

func buildOBProxyDeployment(ctx context.Context, param *obproxy.CreateOBProxyParam, b *buildDeploymentParam) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	deploy.Name = param.Name
	deploy.Namespace = param.Namespace
	deploy.Labels = map[string]string{
		constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
		LabelOBProxy:            param.Name,
		LabelForOBCluster:       b.cluster.Name,
		LabelForNamespace:       b.cluster.Namespace,
		LabelWithConfigMap:      b.cm.Name,
		LabelProxyClusterName:   param.ProxyClusterName,
	}
	deploy.Annotations = map[string]string{
		AnnotationServiceType:    param.ServiceType,
		AnnotationServiceIP:      b.svc.Spec.ClusterIP,
		AnnotationProxySysSecret: b.proxySysSecret.Name,
	}
	quantCPU := resource.NewQuantity(param.Resource.Cpu, resource.DecimalSI)
	quantMemory, err := resource.ParseQuantity(fmt.Sprintf("%dGi", param.Resource.MemoryGB))
	if err != nil {
		return nil, httpErr.NewBadRequest("Failed to parse memory quantity")
	}
	// Get RS_LIST from oceanbase database
	conn, err := utils.GetOBConnection(ctx, b.cluster, "root", "sys", b.cluster.Spec.UserSecrets.Root)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get oceanbase connection")
	}
	defer conn.Close()
	parameters, err := conn.GetParameter(ctx, "rootservice_list", nil)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get rootservice list")
	}
	if len(parameters) == 0 {
		return nil, httpErr.NewInternal("Empty rootservice list")
	}
	rsList := strings.ReplaceAll(parameters[0].Value, ":2882", "")

	deploy.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				LabelOBProxy: param.Name,
			},
		},
		Replicas: &param.Replicas,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					LabelOBProxy:            param.Name,
					constant.LabelManagedBy: constant.DASHBOARD_APP_NAME,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:  "obproxy",
					Image: param.Image,
					Ports: []corev1.ContainerPort{{
						Name:          "sql",
						ContainerPort: 2883,
					}, {
						Name:          "prometheus",
						ContainerPort: 2884,
					}},
					EnvFrom: []corev1.EnvFromSource{{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: b.cm.Name,
							},
						},
					}},
					Env: []corev1.EnvVar{{
						Name:  "APP_NAME",
						Value: param.ProxyClusterName,
					}, {
						Name:  "OB_CLUSTER",
						Value: b.cluster.Spec.ClusterName,
					}, {
						Name:  "RS_LIST",
						Value: rsList,
					}, {
						Name: "PROXYRO_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: b.proxyRoSecret.Name,
								},
								Key: "password",
							},
						},
					}, {
						Name: "PROXYSYS_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: b.proxySysSecret.Name,
								},
								Key: "password",
							},
						},
					}, {
						Name:  "ODP_PROXY_MEM_LIMITED",
						Value: fmt.Sprintf("%dMB", quantMemory.Value()*95/100/(1<<20)),
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *quantCPU,
							corev1.ResourceMemory: quantMemory,
						},
					},
				}},
			},
		},
	}
	return deploy, nil
}
