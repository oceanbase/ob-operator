/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package oceanbase

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/param"
	oberr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

// Check references in the requested namespace before creating any credentials.
// Return only names/status; object-storage credentials never enter API responses.
func validateSharedStorageDependencies(ctx context.Context, p *param.CreateOBClusterParam) error {
	if p.DeploymentMode != "shared_storage" {
		return nil
	}
	k8s := client.GetClient()
	return checkSharedStorageDependencies(ctx, p, k8s.DynamicClient, k8s.ClientSet.CoreV1())
}

func checkSharedStorageDependencies(ctx context.Context, p *param.CreateOBClusterParam, resources dynamic.Interface, core coreclient.CoreV1Interface) error {
	ls, err := resources.Resource(schema.GroupVersionResource{
		Group: "oceanbase.oceanbase.com", Version: "v1alpha1", Resource: "oblogserviceclusters",
	}).Namespace(p.Namespace).Get(ctx, p.LogServiceRef.Name, metav1.GetOptions{})
	if err != nil {
		return oberr.NewBadRequest("Cannot read logServiceRef in the selected namespace: " + err.Error())
	}
	status, _, _ := unstructured.NestedString(ls.Object, "status", "status")
	if ls.GetDeletionTimestamp() != nil || status != "running" {
		return oberr.NewBadRequest("logServiceRef must reference a running OBLogServiceCluster")
	}
	secret, err := core.Secrets(p.Namespace).Get(ctx, p.SharedStorageInfo.SecretRef.Name, metav1.GetOptions{})
	if err != nil {
		return oberr.NewBadRequest("Cannot read sharedStorageInfo.secretRef in the selected namespace: " + err.Error())
	}
	if len(secret.Data["access_id"]) == 0 || len(secret.Data["access_key"]) == 0 {
		return oberr.NewBadRequest("sharedStorageInfo.secretRef must contain non-empty access_id and access_key")
	}
	return nil
}
