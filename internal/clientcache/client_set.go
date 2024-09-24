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

package clientcache

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/internal/clients"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type clientSetEntry struct {
	LatestGeneration int64
	Client           *k8sclient.Client
}

var clientSetCache sync.Map

func GetClientSetFromK8sName(ctx context.Context, k8sClusterName string) (*k8sclient.Client, error) {
	k8sCluster, err := clients.K8sClusterClient.Get(ctx, "", k8sClusterName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get k8s cluster credential")
	}
	if client, ok := clientSetCache.Load(k8sCluster.Name); ok {
		entry := client.(clientSetEntry)
		if entry.LatestGeneration >= k8sCluster.Generation {
			return entry.Client, nil
		}
	}
	kubeConfig, err := k8sCluster.DecodeKubeConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode kubeconfig")
	}
	client, err := k8sclient.GetClientFromBytes(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create k8s client")
	}
	clientSetCache.Store(k8sCluster.Name, clientSetEntry{LatestGeneration: k8sCluster.Generation, Client: client})
	return client, nil
}
