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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	apitypes "github.com/oceanbase/ob-operator/api/types"
)

const (
	envPrefix = "ODP_"
)

// Resource name prefixes
const (
	cmPrefix            = "cm-"
	svcPrefix           = "svc-"
	proxyRoSecretPrefix = "sec-ro-"
)

// Additional label constants
const (
	LabelWithConfigMap    = "obproxy.oceanbase.com/with-config-map"
	LabelProxyClusterName = "obproxy.oceanbase.com/proxy-cluster-name"
)

// Annotation constants
const (
	AnnotationServiceType    = "obproxy.oceanbase.com/service-type"
	AnnotationServiceIP      = "obproxy.oceanbase.com/service-ip"
	AnnotationProxySysSecret = "obproxy.oceanbase.com/proxy-sys-secret"
)

// Container ports
const (
	SqlPort        = 2883
	PrometheusPort = 2884
)

func (m *OBProxyManager) buildOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: m.OBProxy.APIVersion,
		Kind:       m.OBProxy.Kind,
		Name:       m.OBProxy.Name,
		UID:        m.OBProxy.GetUID(),
	}
}

func (m *OBProxyManager) buildOwnerReferenceList() []metav1.OwnerReference {
	return []metav1.OwnerReference{m.buildOwnerReference()}
}

func (m *OBProxyManager) buildCommonLabels() map[string]string {
	clusterNS := m.OBProxy.Spec.OBCluster.Namespace
	if clusterNS == "" {
		clusterNS = m.OBProxy.Namespace
	}

	labels := map[string]string{
		LabelOBProxyInstance:       m.OBProxy.Name,
		LabelRefOBCluster:          m.OBProxy.Spec.OBCluster.Name,
		LabelRefOBClusterNamespace: clusterNS,
	}
	return labels
}

func (m *OBProxyManager) buildConfigMap() *corev1.ConfigMap {
	cmName := cmPrefix + m.OBProxy.Name
	labels := m.buildCommonLabels()

	data := make(map[string]string)
	for _, param := range m.OBProxy.Spec.Parameters {
		key := strings.ToUpper(envPrefix + param.Name)
		data[key] = param.Value
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmName,
			Namespace:       m.OBProxy.Namespace,
			Labels:          labels,
			OwnerReferences: m.buildOwnerReferenceList(),
		},
		Data: data,
	}

	return cm
}

func (m *OBProxyManager) buildService() *corev1.Service {
	svcName := svcPrefix + m.OBProxy.Name
	labels := m.buildCommonLabels()

	svcType := corev1.ServiceTypeClusterIP
	if m.OBProxy.Spec.ServiceType != "" {
		svcType = corev1.ServiceType(m.OBProxy.Spec.ServiceType)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       m.OBProxy.Namespace,
			Labels:          labels,
			OwnerReferences: m.buildOwnerReferenceList(),
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "sql",
					Port:       SqlPort,
					TargetPort: intstr.FromInt(SqlPort),
				},
				{
					Name:       "prometheus",
					Port:       PrometheusPort,
					TargetPort: intstr.FromInt(PrometheusPort),
				},
			},
			Selector: map[string]string{
				LabelOBProxyInstance: m.OBProxy.Name,
			},
			Type: svcType,
		},
	}

	return svc
}

func (m *OBProxyManager) getProxySysSecret() (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.OBProxy.Namespace,
		Name:      m.OBProxy.Spec.ProxySysSecret,
	}, secret)
	if err != nil {
		return nil, errors.Wrap(err, "get proxy sys secret")
	}
	return secret, nil
}

func (m *OBProxyManager) buildDeployment(rsList string, svc *corev1.Service, proxyRoSecret, proxySysSecret *corev1.Secret) *appsv1.Deployment {
	labels := m.buildCommonLabels()
	labels[LabelProxyClusterName] = m.OBProxy.Spec.ProxyClusterName

	podLabels := map[string]string{
		LabelOBProxyInstance: m.OBProxy.Name,
	}

	resources := m.buildResourceRequirements()

	proxyClusterName := m.OBProxy.Spec.ProxyClusterName
	if proxyClusterName == "" {
		proxyClusterName = m.OBProxy.Name
	}

	container := corev1.Container{
		Name:  "obproxy",
		Image: m.OBProxy.Spec.Image,
		Ports: []corev1.ContainerPort{
			{
				Name:          "sql",
				ContainerPort: SqlPort,
			},
			{
				Name:          "prometheus",
				ContainerPort: PrometheusPort,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "RS_LIST",
				Value: rsList,
			},
			{
				Name:  "APP_NAME",
				Value: proxyClusterName,
			},
			{
				Name:  "OB_CLUSTER",
				Value: m.OBProxy.Spec.OBCluster.Name,
			},
			{
				Name: "PROXYRO_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: proxyRoSecret.Name,
						},
						Key: "password",
					},
				},
			},
			{
				Name: "PROXYSYS_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: proxySysSecret.Name,
						},
						Key: "password",
					},
				},
			},
		},
		Resources: resources,
	}

	cmName := m.getConfigMapName()
	container.EnvFrom = []corev1.EnvFromSource{
		{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
			},
		},
	}

	if m.OBProxy.Spec.Resource != nil && !m.OBProxy.Spec.Resource.Memory.IsZero() {
		memoryMB := m.OBProxy.Spec.Resource.Memory.Value() * 95 / 100 / (1 << 20)
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "ODP_PROXY_MEM_LIMITED",
			Value: fmt.Sprintf("%dMB", memoryMB),
		})
	}

	replicas := m.OBProxy.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            m.OBProxy.Name,
			Namespace:       m.OBProxy.Namespace,
			Labels:          labels,
			OwnerReferences: m.buildOwnerReferenceList(),
			Annotations: map[string]string{
				AnnotationServiceType:    string(svc.Spec.Type),
				AnnotationServiceIP:      svc.Spec.ClusterIP,
				AnnotationProxySysSecret: proxySysSecret.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					Containers:         []corev1.Container{container},
					NodeSelector:       m.OBProxy.Spec.NodeSelector,
					Affinity:           m.OBProxy.Spec.Affinity,
					Tolerations:        m.OBProxy.Spec.Tolerations,
					ServiceAccountName: m.OBProxy.Spec.ServiceAccount,
				},
			},
		},
	}

	return deploy
}

func (m *OBProxyManager) buildResourceRequirements() corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Limits: corev1.ResourceList{},
	}

	if m.OBProxy.Spec.Resource != nil {
		if !m.OBProxy.Spec.Resource.Cpu.IsZero() {
			resources.Limits[corev1.ResourceCPU] = m.OBProxy.Spec.Resource.Cpu
		}
		if !m.OBProxy.Spec.Resource.Memory.IsZero() {
			resources.Limits[corev1.ResourceMemory] = m.OBProxy.Spec.Resource.Memory
		}
	}

	return resources
}

func (m *OBProxyManager) buildCopiedProxyROSecret(sourceSecret *corev1.Secret) *corev1.Secret {
	secretName := proxyRoSecretPrefix + m.OBProxy.Name
	labels := m.buildCommonLabels()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            secretName,
			Namespace:       m.OBProxy.Namespace,
			Labels:          labels,
			OwnerReferences: m.buildOwnerReferenceList(),
		},
		Data: sourceSecret.Data,
	}

	return secret
}

func (m *OBProxyManager) getConfigMapName() string {
	return cmPrefix + m.OBProxy.Name
}

func (m *OBProxyManager) getServiceName() string {
	return svcPrefix + m.OBProxy.Name
}

func (m *OBProxyManager) getProxyROSecretName() string {
	return proxyRoSecretPrefix + m.OBProxy.Name
}

func (m *OBProxyManager) updateConfigMapData(cm *corev1.ConfigMap, parameters []apitypes.Parameter) *corev1.ConfigMap {
	cmCopy := cm.DeepCopy()
	cmCopy.Data = make(map[string]string)
	for _, param := range parameters {
		key := strings.ToUpper(envPrefix + param.Name)
		cmCopy.Data[key] = param.Value
	}
	return cmCopy
}

