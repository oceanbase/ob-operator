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
	ctrlruntime "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha2"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	k8sclient "github.com/oceanbase/ob-operator/pkg/k8s/client"
)

type cacheEntry struct {
	LatestGeneration int64
	Client           ctrlruntime.Client
}

var K8sClientCache sync.Map

func GetCachedCtrlRuntimeClientFromK8sCluster(ctx context.Context, k8sCluster string) (ctrlruntime.Client, error) {
	creds := &v1alpha2.K8sClusterCredentialList{}
	err := clients.K8sClusterCredentialClient.List(ctx, "", creds, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelK8sCluster + "=" + k8sCluster,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list k8s cluster credentials")
	}
	if len(creds.Items) == 0 {
		return nil, errors.New("no k8s cluster credential found for k8s cluster " + k8sCluster)
	}
	if len(creds.Items) > 1 {
		return nil, errors.New("more than one credentials found for k8s cluster " + k8sCluster)
	}
	cred := creds.Items[0]
	return GetCachedCtrlRuntimeClient(ctx, &cred)
}

func GetCachedCtrlRuntimeClientFromCredName(ctx context.Context, credentialName string) (ctrlruntime.Client, error) {
	cred, err := clients.K8sClusterCredentialClient.Get(ctx, "", credentialName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get k8s cluster credential")
	}
	return GetCachedCtrlRuntimeClient(ctx, cred)
}

func GetCachedCtrlRuntimeClient(ctx context.Context, cred *v1alpha2.K8sClusterCredential) (ctrlruntime.Client, error) {
	if client, ok := K8sClientCache.Load(cred.Name); ok {
		entry := client.(cacheEntry)
		if entry.LatestGeneration >= cred.Generation {
			return entry.Client, nil
		}
	}

	config, err := k8sclient.GetConfigFromBytes([]byte(cred.Spec.KubeConfig))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get config from kubeconfig field of %s", cred.Name)
	}
	client, err := k8sclient.GetCtrlRuntimeClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create k8s client")
	}
	K8sClientCache.Store(cred.Name, cacheEntry{LatestGeneration: cred.Generation, Client: client})
	return client, nil
}
