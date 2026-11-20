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
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
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
	uri, err := resourceutils.BackupDestinationPath(m.Ctx, m.Client, m.Resource.Namespace, dest)
	if err != nil {
		m.PrintErrEvent(err)
	}
	return uri
}
