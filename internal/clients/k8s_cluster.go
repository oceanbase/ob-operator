/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package clients

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	k8sv1alpha1 "github.com/oceanbase/ob-operator/api/k8sv1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients/schema"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func ListAllK8sClusters(ctx context.Context) (*k8sv1alpha1.K8sClusterList, error) {
	client := client.GetClient()
	var k8sClusterList k8sv1alpha1.K8sClusterList
	obj, err := client.DynamicClient.Resource(schema.K8sClusterGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "List k8s clusters")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &k8sClusterList)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured object to k8s cluster list")
	}
	return &k8sClusterList, nil
}

func DeleteK8sCluster(ctx context.Context, name string) error {
	return K8sClusterClient.Delete(ctx, "", name, metav1.DeleteOptions{})
}

func GetK8sCluster(ctx context.Context, name string) (*k8sv1alpha1.K8sCluster, error) {
	client := client.GetClient()
	var k8sCluster k8sv1alpha1.K8sCluster
	obj, err := client.DynamicClient.Resource(schema.K8sClusterGVR).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Get k8s cluster")
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &k8sCluster)
	if err != nil {
		return nil, errors.Wrap(err, "Convert unstructured object to k8s cluster")
	}
	return &k8sCluster, nil
}

func UpdateK8sCluster(ctx context.Context, k8sCluster *k8sv1alpha1.K8sCluster) (*k8sv1alpha1.K8sCluster, error) {
	return K8sClusterClient.Update(ctx, k8sCluster, metav1.UpdateOptions{})
}

func CreateK8sCluster(ctx context.Context, k8sCluster *k8sv1alpha1.K8sCluster) (*k8sv1alpha1.K8sCluster, error) {
	return K8sClusterClient.Create(ctx, k8sCluster, metav1.CreateOptions{})
}
