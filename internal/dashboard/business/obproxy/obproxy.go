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

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/obproxy"
	"github.com/oceanbase/ob-operator/internal/dashboard/utils"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func ListOBProxies(ctx context.Context, ns string, listOptions metav1.ListOptions) ([]obproxy.OBProxyOverview, error) {
	if listOptions.LabelSelector == "" {
		listOptions.LabelSelector = LabelOBProxy
	} else {
		listOptions.LabelSelector += "," + LabelOBProxy
	}
	deployments, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).List(ctx, listOptions)
	if err != nil {
		return nil, httpErr.NewInternal("Failed to list obproxies, err msg: " + err.Error())
	}
	obproxies := make([]obproxy.OBProxyOverview, 0, len(deployments.Items))
	for _, deploy := range deployments.Items {
		obproxies = append(obproxies, *buildOBProxyOverview(&deploy))
	}
	return obproxies, nil
}

func GetOBProxy(ctx context.Context, ns, name string) (*obproxy.OBProxy, error) {
	deployment, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to get obproxy, err msg: " + err.Error())
	}
	return buildOBProxy(ctx, deployment)
}

func CreateOBProxy(ctx context.Context, param *obproxy.CreateOBProxyParam) (*obproxy.OBProxy, error) {
	clusterNs := param.OBCluster.Namespace
	clusterName := param.OBCluster.Name
	obcluster, err := clients.ClusterClient.Get(ctx, clusterNs, clusterName, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewBadRequest("OBCluster not found")
		}
		return nil, httpErr.NewInternal("Failed to get obcluster, err msg: " + err.Error())
	}
	// TODO: Decrypted the password
	proxySysSecret, err := createPasswordSecret(ctx, param.Namespace, proxySysSecretPrefix+param.Name, param.ProxySysPassword)
	if err != nil {
		return nil, err
	}
	var proxyRoSecret *corev1.Secret
	proxyRoSecret, err = copyPasswordSecret(ctx, clusterNs, obcluster.Spec.UserSecrets.ProxyRO, param.Namespace, proxyRoSecretPrefix+param.Name)
	if err != nil {
		return nil, err
	}
	cm, err := createConfigMap(ctx, param.Namespace, param.Name, param)
	if err != nil {
		return nil, err
	}
	svc, err := createOBProxyService(ctx, param.Namespace, param.Name, corev1.ServiceType(param.ServiceType))
	if err != nil {
		return nil, err
	}
	innerParam := buildDeploymentParam{
		cm:             cm,
		svc:            svc,
		cluster:        obcluster,
		proxyRoSecret:  proxyRoSecret,
		proxySysSecret: proxySysSecret,
	}
	deployBody, err := buildOBProxyDeployment(ctx, param, &innerParam)
	if err != nil {
		return nil, err
	}
	deployment, err := client.GetClient().ClientSet.AppsV1().Deployments(param.Namespace).Create(ctx, deployBody, metav1.CreateOptions{})
	if err != nil {
		if kubeerrors.IsAlreadyExists(err) {
			return nil, httpErr.NewBadRequest("OBProxy already exists")
		}
		return nil, httpErr.NewInternal("Failed to create obproxy, err msg: " + err.Error())
	}
	return buildOBProxy(ctx, deployment)
}

func PatchOBProxy(ctx context.Context, ns, name string, param *obproxy.PatchOBProxyParam) (*obproxy.OBProxy, error) {
	deploy, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("OBProxy not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy, err msg: " + err.Error())
	}
	if param.ServiceType != nil {
		svc, err := updateOBProxyService(ctx, ns, name, corev1.ServiceType(*param.ServiceType))
		if err != nil {
			return nil, err
		}
		deploy.Annotations[AnnotationServiceType] = string(svc.Spec.Type)
	}
	updated := false
	if param.Image != nil {
		deploy.Spec.Template.Spec.Containers[0].Image = *param.Image
		updated = true
	}
	if param.Replicas != nil {
		deploy.Spec.Replicas = param.Replicas
		updated = true
	}
	if param.Resource != nil {
		quantCPU := resource.NewQuantity(param.Resource.Cpu, resource.DecimalSI)
		quantMemory, err := resource.ParseQuantity(fmt.Sprintf("%dGi", param.Resource.MemoryGB))
		if err != nil {
			return nil, httpErr.NewBadRequest("Failed to parse memory quantity")
		}
		deploy.Spec.Template.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
			corev1.ResourceCPU:    *quantCPU,
			corev1.ResourceMemory: quantMemory,
		}
		updated = true
	}

	parametersUpdated := false
	if param.Parameters != nil {
		changed, err := doesParametersChanged(ctx, ns, name, param)
		if err != nil {
			return nil, err
		}
		if changed {
			_, err := updateConfigMap(ctx, ns, name, param)
			if err != nil {
				return nil, err
			}
			parametersUpdated = true
		}
	}
	odp, err := buildOBProxy(ctx, deploy)
	if err != nil {
		return nil, err
	}
	if parametersUpdated {
		type Parameter struct {
			VariableName string `json:"Variable_name"`
			Value        string `json:"Value"`
		}
		for _, pod := range odp.Pods {
			conn, err := utils.GetOBConnectionByHost(ctx, odp.Namespace, pod.PodIP, "root", "proxysys", odp.ProxySysSecret, 2883)
			if err != nil {
				return nil, httpErr.NewInternal("Failed to get oceanbase connection by host " + pod.PodIP)
			}
			for _, param := range param.Parameters {
				err = conn.ExecWithDefaultTimeout(ctx, fmt.Sprintf("ALTER proxyconfig SET %s = ?;", param.Key), param.Value)
				if err != nil {
					return nil, httpErr.NewInternal("Failed to update obproxy config, err msg: " + err.Error())
				}
			}
		}
	}
	if updated || parametersUpdated {
		deployment, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).Update(ctx, deploy, metav1.UpdateOptions{})
		if err != nil {
			return nil, httpErr.NewInternal("Failed to update obproxy, err msg: " + err.Error())
		}
		return buildOBProxy(ctx, deployment)
	}
	return odp, nil
}

func DeleteOBProxy(ctx context.Context, ns, name string) (*obproxy.OBProxy, error) {
	deploy, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("OBProxy not found")
		}
	}
	deleted, _ := buildOBProxy(ctx, deploy)
	_, err = deleteOBProxyService(ctx, ns, name)
	if err != nil {
		return nil, err
	}
	err = client.GetClient().ClientSet.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to delete obproxy, err msg: " + err.Error())
	}
	_, err = deleteConfigMap(ctx, ns, name)
	if err != nil {
		return nil, err
	}
	err = client.GetClient().ClientSet.CoreV1().Secrets(ns).Delete(ctx, proxySysSecretPrefix+name, metav1.DeleteOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to delete obproxy secret, err msg: " + err.Error())
	}
	err = client.GetClient().ClientSet.CoreV1().Secrets(ns).Delete(ctx, proxyRoSecretPrefix+name, metav1.DeleteOptions{})
	if err != nil {
		return nil, httpErr.NewInternal("Failed to delete obproxy secret, err msg: " + err.Error())
	}
	return deleted, nil
}

func ListOBProxyParameters(ctx context.Context, ns string, name string) ([]obproxy.ConfigItem, error) {
	items := make([]obproxy.ConfigItem, 0)
	deploy, err := client.GetClient().ClientSet.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound("OBProxy not found")
		}
		return nil, httpErr.NewInternal("Failed to get obproxy, err msg: " + err.Error())
	}
	odp, err := buildOBProxy(ctx, deploy)
	if err != nil {
		return nil, err
	}
	for _, pod := range odp.Pods {
		conn, err := utils.GetOBConnectionByHost(ctx, odp.Namespace, pod.PodIP, "root", "proxysys", odp.ProxySysSecret, 2883)
		if err != nil {
			logrus.Infof("Failed to get oceanbase connection by host %s", pod.PodIP)
			continue
		}
		err = conn.QueryList(ctx, &items, "SHOW PROXYCONFIG;")
		if err != nil {
			return nil, httpErr.NewInternal("Failed to list obproxy config, err msg: " + err.Error())
		}
		return items, nil
	}
	return items, nil
}
