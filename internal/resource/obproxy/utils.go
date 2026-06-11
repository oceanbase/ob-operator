/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OR ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package obproxy

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
)

const (
	// LabelOBProxyInstance is the label key for OBProxy instance
	LabelOBProxyInstance = "obproxy.oceanbase.com/obproxy"
	// LabelRefOBCluster is the label key for referencing OBCluster
	LabelRefOBCluster = "obproxy.oceanbase.com/obcluster"
	// LabelRefOBClusterNamespace is the label key for referencing OBCluster namespace
	LabelRefOBClusterNamespace = "obproxy.oceanbase.com/obcluster-namespace"
)

// getOBProxySelector returns label selector for OBProxy resources
func (m *OBProxyManager) getOBProxySelector() labels.Selector {
	return labels.SelectorFromSet(labels.Set{
		LabelOBProxyInstance: m.OBProxy.Name,
	})
}

// getOBProxyDeployment gets the Deployment owned by OBProxy
func (m *OBProxyManager) getOBProxyDeployment() (*appsv1.Deployment, error) {
	deploymentList := &appsv1.DeploymentList{}
	err := m.Client.List(m.Ctx, deploymentList,
		client.MatchingLabelsSelector{Selector: m.getOBProxySelector()},
		client.InNamespace(m.OBProxy.Namespace),
	)
	if err != nil {
		return nil, errors.Wrap(err, "list obproxy deployments")
	}

	if len(deploymentList.Items) == 0 {
		return nil, nil
	}

	// Return the first deployment (should only be one)
	return &deploymentList.Items[0], nil
}

// getOBProxyService gets the Service owned by OBProxy
func (m *OBProxyManager) getOBProxyService() (*corev1.Service, error) {
	serviceList := &corev1.ServiceList{}
	err := m.Client.List(m.Ctx, serviceList,
		client.MatchingLabelsSelector{Selector: m.getOBProxySelector()},
		client.InNamespace(m.OBProxy.Namespace),
	)
	if err != nil {
		return nil, errors.Wrap(err, "list obproxy services")
	}

	if len(serviceList.Items) == 0 {
		return nil, nil
	}

	return &serviceList.Items[0], nil
}

// getOBProxyConfigMap gets the ConfigMap owned by OBProxy
func (m *OBProxyManager) getOBProxyConfigMap() (*corev1.ConfigMap, error) {
	cmList := &corev1.ConfigMapList{}
	err := m.Client.List(m.Ctx, cmList,
		client.MatchingLabelsSelector{Selector: m.getOBProxySelector()},
		client.InNamespace(m.OBProxy.Namespace),
	)
	if err != nil {
		return nil, errors.Wrap(err, "list obproxy configmaps")
	}

	if len(cmList.Items) == 0 {
		return nil, nil
	}

	return &cmList.Items[0], nil
}

// getOBCluster gets the OBCluster referenced by OBProxy
func (m *OBProxyManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	clusterNS := m.OBProxy.Spec.OBCluster.Namespace
	if clusterNS == "" {
		clusterNS = m.OBProxy.Namespace
	}

	clusterKey := types.NamespacedName{
		Namespace: clusterNS,
		Name:      m.OBProxy.Spec.OBCluster.Name,
	}

	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, clusterKey, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}

	return obcluster, nil
}

// isOBProxyReady checks if OBProxy Deployment is ready
func (m *OBProxyManager) isOBProxyReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}

	// Check if deployment has replicas
	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas == 0 {
		return false
	}

	// Check if all replicas are ready
	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
		deployment.Status.AvailableReplicas == *deployment.Spec.Replicas
}

