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

package obtenantrestore

import (
	"path"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/constants"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

func (m *ObTenantRestoreManager) getClusterSysClient() (*operation.OceanbaseOperationManager, error) {
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      m.Resource.Spec.TargetCluster,
	}, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	con, err := resourceutils.GetSysOperationClient(m.Client, m.Logger, obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get oceanbase operation manager")
	}
	return con, nil
}

func (m *ObTenantRestoreManager) getSourceUri() (string, error) {
	source := m.Resource.Spec.Source
	if source.SourceUri != "" {
		return source.SourceUri, nil
	}
	var bakPath, archivePath string
	if source.BakDataSource != nil && source.BakDataSource.Type == constants.BackupDestTypeOSS {
		accessId, accessKey, err := m.readAccessCredentials(source.BakDataSource.OSSAccessSecret)
		if err != nil {
			return "", err
		}
		bakPath = strings.Join([]string{source.BakDataSource.Path, "access_id=" + accessId, "access_key=" + accessKey}, "&")
	} else {
		bakPath = "file://" + path.Join(oceanbaseconst.BackupPath, source.BakDataSource.Path)
	}

	if source.ArchiveSource != nil && source.ArchiveSource.Type == constants.BackupDestTypeOSS {
		accessId, accessKey, err := m.readAccessCredentials(source.ArchiveSource.OSSAccessSecret)
		if err != nil {
			return "", err
		}
		archivePath = strings.Join([]string{source.ArchiveSource.Path, "access_id=" + accessId, "access_key=" + accessKey}, "&")
	} else {
		archivePath = "file://" + path.Join(oceanbaseconst.BackupPath, source.ArchiveSource.Path)
	}

	if bakPath == "" || archivePath == "" {
		return "", errors.New("Unexpected error: both bakPath and archivePath must be set")
	}

	return strings.Join([]string{bakPath, archivePath}, ","), nil
}

func (m *ObTenantRestoreManager) readAccessCredentials(secretName string) (accessId, accessKey string, err error) {
	secret := &v1.Secret{}
	err = m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.Namespace,
		Name:      secretName,
	}, secret)
	if err != nil {
		return "", "", err
	}
	accessId = string(secret.Data["accessId"])
	accessKey = string(secret.Data["accessKey"])
	return accessId, accessKey, nil
}
