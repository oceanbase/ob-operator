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

package utils

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/connector"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func GetSysOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster) (*operation.OceanbaseOperationManager, error) {
	logger.V(oceanbaseconst.LogLevelTrace).Info("Get cluster sys client", "obCluster", obcluster)
	var manager *operation.OceanbaseOperationManager
	var err error
	_, migrateAnnoExist := GetAnnotationField(obcluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	if migrateAnnoExist && obcluster.Status.Status == clusterstatus.MigrateFromExisting {
		manager, err = getSysClientFromSourceCluster(c, logger, obcluster, oceanbaseconst.RootUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Root)
	} else {
		manager, err = getSysClient(c, logger, obcluster, oceanbaseconst.OperatorUser, oceanbaseconst.SysTenant, obcluster.Spec.UserSecrets.Operator)
	}
	return manager, err
}

func GetTenantRootOperationClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, tenantName, credential string) (*operation.OceanbaseOperationManager, error) {
	logger.V(oceanbaseconst.LogLevelTrace).Info("Get tenant root client", "obCluster", obcluster, "tenantName", tenantName, "credential", credential)
	observerList := &v1alpha1.OBServerList{}
	err := c.List(context.Background(), observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: obcluster.Name,
	}, client.InNamespace(obcluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "Get observer list")
	}
	if len(observerList.Items) == 0 {
		return nil, errors.Errorf("No observer belongs to cluster %s", obcluster.Name)
	}
	var password string
	if credential != "" {
		password, err = ReadPassword(c, obcluster.Namespace, credential)
		if err != nil {
			return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
		}
	}

	var s *connector.OceanBaseDataSource
	for _, observer := range observerList.Items {
		address := observer.Status.GetConnectAddr()
		switch obcluster.Status.Status {
		case clusterstatus.New:
			return nil, errors.New("Cluster is not bootstrapped")
		case clusterstatus.Bootstrapped:
			return nil, errors.New("Cluster is not initialized")
		default:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, password, oceanbaseconst.DefaultDatabase)
		}
		// if err is nil, db connection is already checked available
		rootClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && rootClient != nil {
			rootClient.Logger = logger
			return rootClient, nil
		}
		// err is not nil, try to use empty password
		s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", oceanbaseconst.DefaultDatabase)
		rootClient, err = operation.GetOceanbaseOperationManager(s)
		if err == nil && rootClient != nil {
			rootClient.Logger = logger
			return rootClient, nil
		}
	}
	return nil, errors.Errorf("Can not get root operation client of tenant %s in obcluster %s after checked all servers", tenantName, obcluster.Name)
}

func getSysClientFromSourceCluster(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
	sysClient, err := getSysClient(c, logger, obcluster, userName, tenantName, secretName)
	if err == nil {
		return sysClient, nil
	}
	password, err := ReadPassword(c, obcluster.Namespace, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
	}
	// when obcluster is under migrating, use address from annotation
	migrateAnnoVal, _ := GetAnnotationField(obcluster, oceanbaseconst.AnnotationsSourceClusterAddress)
	servers := strings.Split(migrateAnnoVal, ";")
	for _, server := range servers {
		addressParts := strings.Split(server, ":")
		if len(addressParts) != 2 {
			return nil, errors.New("Parse oceanbase cluster connect address failed")
		}
		sqlPort, err := strconv.ParseInt(addressParts[1], 10, 64)
		if err != nil {
			return nil, errors.New("Parse sql port of obcluster failed")
		}
		s := connector.NewOceanBaseDataSource(addressParts[0], sqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		// if err is nil, db connection is already checked available
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		if err == nil && sysClient != nil {
			sysClient.Logger = logger
			return sysClient, nil
		}
		logger.Error(err, "Get operation manager from existing obcluster")
	}
	return nil, errors.Errorf("Failed to get sys client from existing obcluster, address: %s", migrateAnnoVal)
}

func getSysClient(c client.Client, logger *logr.Logger, obcluster *v1alpha1.OBCluster, userName, tenantName, secretName string) (*operation.OceanbaseOperationManager, error) {
	observerList := &v1alpha1.OBServerList{}
	err := c.List(context.Background(), observerList, client.MatchingLabels{
		oceanbaseconst.LabelRefOBCluster: obcluster.Name,
	}, client.InNamespace(obcluster.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "Get observer list")
	}
	if len(observerList.Items) == 0 {
		return nil, errors.Errorf("No observer belongs to cluster %s", obcluster.Name)
	}

	var s *connector.OceanBaseDataSource
	password, err := ReadPassword(c, obcluster.Namespace, secretName)
	if err != nil {
		return nil, errors.Wrapf(err, "Read password to get oceanbase operation manager of cluster %s", obcluster.Name)
	}
	for _, observer := range observerList.Items {
		address := observer.Status.GetConnectAddr()
		switch obcluster.Status.Status {
		case clusterstatus.New:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", "")
		case clusterstatus.Bootstrapped:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, oceanbaseconst.RootUser, tenantName, "", oceanbaseconst.DefaultDatabase)
		default:
			s = connector.NewOceanBaseDataSource(address, oceanbaseconst.SqlPort, userName, tenantName, password, oceanbaseconst.DefaultDatabase)
		}
		sysClient, err := operation.GetOceanbaseOperationManager(s)
		var checkConnectionErr error
		if obcluster.Status.Status != clusterstatus.New && err == nil && sysClient != nil {
			sysClient.Logger = logger
			_, checkConnectionErr = sysClient.ListServers(context.Background())
		}
		if err == nil && sysClient != nil && checkConnectionErr == nil {
			sysClient.Logger = logger
			return sysClient, nil
		}
	}
	return nil, errors.Errorf("Can not get oceanbase operation manager of obcluster %s after checked all servers", obcluster.Name)
}