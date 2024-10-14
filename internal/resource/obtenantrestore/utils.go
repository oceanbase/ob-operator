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
	apitypes "github.com/oceanbase/ob-operator/api/types"
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
	bakPath = m.getDestPath(source.BakDataSource)
	archivePath = m.getDestPath(source.ArchiveSource)

	if bakPath == "" || archivePath == "" {
		return "", errors.New("Unexpected error: both bakPath and archivePath must be set")
	}

	return strings.Join([]string{bakPath, archivePath}, ","), nil
}

func (m *ObTenantRestoreManager) getDestPath(dest *apitypes.BackupDestination) string {
	if dest.Type == constants.BackupDestTypeNFS || resourceutils.IsZero(dest.Type) {
		return "file://" + path.Join(oceanbaseconst.BackupPath, dest.Path)
	}
	if dest.OSSAccessSecret == "" {
		return ""
	}
	secret := &v1.Secret{}
	err := m.Client.Get(m.Ctx, types.NamespacedName{
		Namespace: m.Resource.GetNamespace(),
		Name:      dest.OSSAccessSecret,
	}, secret)
	if err != nil {
		m.PrintErrEvent(err)
		return ""
	}
	destPath := strings.Join([]string{dest.Path, "access_id=" + string(secret.Data["accessId"]), "access_key=" + string(secret.Data["accessKey"])}, "&")
	if dest.Type == constants.BackupDestTypeCOS {
		destPath += ("&appid=" + string(secret.Data["appId"]))
	} else if dest.Type == constants.BackupDestTypeS3 {
		destPath += ("&s3_region=" + string(secret.Data["s3Region"]))
	}
	return destPath
}
