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

package oceanbase

import (
	"context"
	"sort"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	"github.com/oceanbase/ob-operator/internal/clients"
	obagentconst "github.com/oceanbase/ob-operator/internal/const/obagent"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	secretconst "github.com/oceanbase/ob-operator/internal/const/secret"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	observerstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

type ConnectionManager struct {
	obcluster     *v1alpha1.OBCluster
	connectionMap map[string]*operation.OceanbaseOperationManager
	ctx           context.Context
	mu            sync.Mutex
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(ctx context.Context, obcluster *v1alpha1.OBCluster) *ConnectionManager { // Change to logr.Logger
	return &ConnectionManager{
		ctx:       ctx,
		obcluster: obcluster,
	}
}

func (cm *ConnectionManager) readPassword() (string, error) {
	secret := &corev1.Secret{}
	client := client.GetClient()
	secret, err := client.ClientSet.CoreV1().Secrets(cm.obcluster.Namespace).Get(cm.ctx, cm.obcluster.Spec.UserSecrets.Monitor, v1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(err, "Failed to get secret")
	}
	return string(secret.Data[secretconst.PasswordKeyName]), err
}

func (cm *ConnectionManager) GetSysReadonlyConnection() (*operation.OceanbaseOperationManager, error) {
	return cm.GetSysReadonlyConnectionByIP("")
}

func (cm *ConnectionManager) GetSysReadonlyConnectionByIP(svrIP string) (*operation.OceanbaseOperationManager, error) {
	observerList, err := clients.ListOBServersOfOBCluster(cm.ctx, cm.obcluster)

	if err != nil {
		return nil, errors.Wrap(err, "Get observers")
	}
	if len(observerList.Items) == 0 {
		return nil, errors.Errorf("No observer belongs to cluster %s", cm.obcluster.Name)
	}

	sort.Slice(observerList.Items, func(i, j int) bool {
		return observerList.Items[i].Status.Status == observerstatus.Running && observerList.Items[j].Status.Status != observerstatus.Running
	})

	var s *connector.OceanBaseDataSource
	password, err := cm.readPassword()
	if err != nil {
		return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", cm.obcluster.Name)
	}

	addresses := make([]string, 0, len(observerList.Items)+1)
	if svrIP != "" {
		addresses = append(addresses, svrIP)
	}
	for _, observer := range observerList.Items {
		if observer.Status.GetConnectAddr() == svrIP {
			continue
		}
		addresses = append(addresses, observer.Status.GetConnectAddr())
	}
	for _, address := range addresses {
		s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, obagentconst.MonitorUser, oceanbaseconst.SysTenant, password, oceanbaseconst.DefaultDatabase)
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		if err != nil {
			continue
		}
		clientLogger := logr.FromContextOrDiscard(cm.ctx)
		sysClient.Logger = &clientLogger
		var checkConnectionErr error
		if cm.obcluster.Status.Status != clusterstatus.New && sysClient != nil {
			_, checkConnectionErr = sysClient.ListServers(cm.ctx)
		}
		if sysClient != nil && checkConnectionErr == nil {
			return sysClient, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all servers", cm.obcluster.Name)
}
