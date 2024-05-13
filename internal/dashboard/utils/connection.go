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

package utils

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func GetOBConnection(ctx context.Context, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
	observerList := &v1alpha1.OBServerList{}
	err := clients.ServerClient.List(ctx, obcluster.Namespace, observerList, metav1.ListOptions{
		LabelSelector: oceanbaseconst.LabelRefOBCluster + "=" + obcluster.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Get observer list")
	}
	if len(observerList.Items) == 0 {
		return nil, errors.Errorf("No observer belongs to cluster %s", obcluster.Name)
	}

	var s *connector.OceanBaseDataSource
	secret, err := client.GetClient().ClientSet.CoreV1().Secrets(obcluster.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Get secret %s", secretName)
	}

	password := string(secret.Data["password"])
	for _, observer := range observerList.Items {
		address := observer.Status.GetConnectAddr()
		s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		// if err is nil, db connection is already checked available
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && sysClient != nil {
			dummy := logr.Discard()
			sysClient.Logger = &dummy
			return sysClient, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all server", obcluster.Name)
}
