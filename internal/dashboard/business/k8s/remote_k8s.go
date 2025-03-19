/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package k8s

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sv1alpha1 "github.com/oceanbase/ob-operator/api/k8sv1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	"github.com/oceanbase/ob-operator/internal/dashboard/business/common"
	modelcommon "github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

func ListRemoteK8sClusters(ctx context.Context) ([]k8s.K8sClusterInfo, error) {
	k8sClusterList, err := clients.ListAllK8sClusters(ctx)
	if err != nil {
		return nil, err
	}
	k8sClusters := make([]k8s.K8sClusterInfo, 0)
	for _, k8sCluster := range k8sClusterList.Items {
		k8sClusters = append(k8sClusters, k8s.K8sClusterInfo{
			Name:        k8sCluster.Name,
			Description: k8sCluster.Spec.Description,
			CreatedAt:   k8sCluster.CreationTimestamp.Unix(),
		})
	}
	return k8sClusters, nil
}

func GetRemoteK8sCluster(ctx context.Context, name string) (*k8s.K8sClusterInfo, error) {
	k8sCluster, err := clients.GetK8sCluster(ctx, name)
	if err != nil {
		return nil, err
	}
	return &k8s.K8sClusterInfo{
		Name:        k8sCluster.Name,
		Description: k8sCluster.Spec.Description,
		CreatedAt:   k8sCluster.CreationTimestamp.Unix(),
	}, nil
}

func DeleteRemoteK8sCluster(ctx context.Context, name string) error {
	return clients.DeleteK8sCluster(ctx, name)
}

func UpdateRemoteK8sCluster(ctx context.Context, name string, param *k8s.UpdateK8sClusterParam) (*k8s.K8sClusterInfo, error) {
	k8sCluster, err := clients.GetK8sCluster(ctx, name)
	if err != nil {
		return nil, err
	}
	k8sCluster.Spec.Description = param.Description
	if param.KubeConfig != "" {
		k8sCluster.Spec.KubeConfig = param.KubeConfig
	}
	updatedK8sCluster, err := clients.UpdateK8sCluster(ctx, k8sCluster)
	if err != nil {
		return nil, err
	}
	k8sInfo := &k8s.K8sClusterInfo{
		Name:        updatedK8sCluster.Name,
		Description: updatedK8sCluster.Spec.Description,
		CreatedAt:   updatedK8sCluster.CreationTimestamp.Unix(),
	}
	return k8sInfo, nil
}

func CreateRemoteK8sCluster(ctx context.Context, param *k8s.CreateK8sClusterParam) (*k8s.K8sClusterInfo, error) {
	k8sCluster := &k8sv1alpha1.K8sCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        param.Name,
			Annotations: map[string]string{},
		},
		Spec: k8sv1alpha1.K8sClusterSpec{
			Name:        param.Name,
			Description: param.Description,
			KubeConfig:  param.KubeConfig,
		},
	}
	createdK8sCluster, err := clients.CreateK8sCluster(ctx, k8sCluster)
	if err != nil {
		return nil, err
	}
	k8sInfo := &k8s.K8sClusterInfo{
		Name:        createdK8sCluster.Name,
		Description: createdK8sCluster.Spec.Description,
		CreatedAt:   createdK8sCluster.CreationTimestamp.Unix(),
	}
	return k8sInfo, nil
}

func ListRemoteK8sClusterNodes(ctx context.Context, clusterName string) ([]response.K8sNode, error) {
	c, err := GetClientForK8sCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Get client for k8s cluster %s", clusterName))
	}
	return ListK8sClusterNodes(ctx, c)
}

func ListRemoteK8sClusterEvents(ctx context.Context, clusterName string, queryEventParam *param.QueryEventParam) ([]response.K8sEvent, error) {
	c, err := GetClientForK8sCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Get client for k8s cluster %s", clusterName))
	}
	return ListK8sClusterEvents(ctx, c, queryEventParam)
}

func UpdateRemoteK8sClusterNodeLabels(ctx context.Context, clusterName string, nodeName string, labels []modelcommon.KVPair) (*response.K8sNode, error) {
	c, err := GetClientForK8sCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Get client for k8s cluster %s", clusterName))
	}

	node, err := c.ClientSet.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	node.Labels = common.KVsToMap(labels)
	node, err = c.ClientSet.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return NewK8sNodeResponse(node), err
}

func UpdateRemoteK8sClusterNodeTaints(ctx context.Context, clusterName string, nodeName string, taints []k8s.Taint) (*response.K8sNode, error) {
	c, err := GetClientForK8sCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Get client for k8s cluster %s", clusterName))
	}
	node, err := c.ClientSet.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	nodeTaints := make([]corev1.Taint, 0)
	for _, taint := range taints {
		nodeTaints = append(nodeTaints, corev1.Taint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: corev1.TaintEffect(taint.Effect),
		})
	}
	node.Spec.Taints = nodeTaints
	node, err = c.ClientSet.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return NewK8sNodeResponse(node), err
}
func BatchUpdateRemoteK8sClusterNodes(ctx context.Context, clusterName string, updateNodesParam *param.BatchUpdateNodesParam) error {
	c, err := GetClientForK8sCluster(ctx, clusterName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Get client for k8s cluster %s", clusterName))
	}
	return BatchUpdateK8sClusterNodes(ctx, c, updateNodesParam)
}
