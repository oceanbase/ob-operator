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
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/oceanbase/ob-operator/api/constants"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

// Restore progress:
// 1. create unit (in tenant manager)
// 2. create resource pool (in tenant manager)
// 3. trigger restore job
// 4. wait for finishing
// 5. upgrade tenant if needed
// 6. activate or replay log or do nothing

// OBTenantRestore tasks

func (m *ObTenantRestoreManager) StartRestoreJobInOB() tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	restoreSpec := m.Resource.Spec.Source
	sourceUri, err := m.getSourceUri()
	if err != nil {
		return err
	}

	if restoreSpec.BakEncryptionSecret != "" {
		password, err := resourceutils.ReadPassword(m.Client, m.Resource.Namespace, restoreSpec.BakEncryptionSecret)
		if err != nil {
			m.Recorder.Event(m.Resource, v1.EventTypeWarning, "ReadRestorePasswordFailed", err.Error())
			return err
		}
		err = con.SetRestorePassword(password)
		if err != nil {
			m.Recorder.Event(m.Resource, v1.EventTypeWarning, "SetRestorePasswordFailed", err.Error())
			return err
		}
	}

	if restoreSpec.Until.Unlimited {
		err = con.StartRestoreUnlimited(m.Resource.Spec.TargetTenant, sourceUri, m.Resource.Spec.Option)
		if err != nil {
			return err
		}
	} else {
		if restoreSpec.Until.Timestamp != nil {
			err = con.StartRestoreWithLimit(m.Resource.Spec.TargetTenant, sourceUri, m.Resource.Spec.Option, "TIME", *restoreSpec.Until.Timestamp)
			if err != nil {
				return err
			}
		} else if restoreSpec.Until.Scn != nil {
			err = con.StartRestoreWithLimit(m.Resource.Spec.TargetTenant, sourceUri, m.Resource.Spec.Option, "SCN", *restoreSpec.Until.Scn)
			if err != nil {
				return err
			}
		} else {
			return errors.New("Restore until must have a limit key, scn and timestamp are both nil now")
		}
	}
	return nil
}

func (m *ObTenantRestoreManager) StartLogReplay() tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	if m.Resource.Spec.PrimaryTenant != nil {
		restoreSource, err := resourceutils.GetTenantRestoreSource(m.Ctx, m.Client, m.Logger, m.Resource.Namespace, *m.Resource.Spec.PrimaryTenant)
		if err != nil {
			return err
		}
		err = con.SetParameter("LOG_RESTORE_SOURCE", restoreSource, &param.Scope{
			Name:  "TENANT",
			Value: m.Resource.Spec.TargetTenant,
		})
		if err != nil {
			m.Logger.Error(err, "Failed to set log restore source")
			return err
		}
	}
	replayUntil := m.Resource.Spec.Source.ReplayLogUntil
	if replayUntil == nil || replayUntil.Unlimited {
		err = con.ReplayStandbyLog(m.Resource.Spec.TargetTenant, "UNLIMITED")
	} else if replayUntil.Timestamp != nil {
		err = con.ReplayStandbyLog(m.Resource.Spec.TargetTenant, fmt.Sprintf("TIME='%s'", *replayUntil.Timestamp))
	} else if replayUntil.Scn != nil {
		err = con.ReplayStandbyLog(m.Resource.Spec.TargetTenant, fmt.Sprintf("SCN=%s", *replayUntil.Scn))
	} else {
		return errors.New("Replay until with limit must have a limit key, scn and timestamp are both nil now")
	}
	return err
}

func (m *ObTenantRestoreManager) ActivateStandby() tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	return con.ActivateStandby(m.Resource.Spec.TargetTenant)
}

func (m *ObTenantRestoreManager) getClusterSysClient() (*operation.OceanbaseOperationManager, error) {
	if m.con != nil {
		return m.con, nil
	}
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
	m.con = con
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
