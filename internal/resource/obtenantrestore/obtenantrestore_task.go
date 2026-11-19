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
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/param"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
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

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*ObTenantRestoreManager]()

func StartRestoreJobInOB(m *ObTenantRestoreManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	description := restoreDescription(m.Resource)
	if m.Resource.Annotations[restoreSubmission] == description {
		return m.awaitSubmittedRestore(con, description)
	}
	restoreSpec := m.Resource.Spec.Source
	sourceUri, err := m.getSourceUri()
	if err != nil {
		return err
	}

	var key string
	var limit any
	if !restoreSpec.Until.Unlimited {
		switch {
		case restoreSpec.Until.Timestamp != nil:
			if strings.TrimSpace(*restoreSpec.Until.Timestamp) == "" {
				return errors.New("restore cutoff time is empty")
			}
			key, limit = "TIME", *restoreSpec.Until.Timestamp
		case restoreSpec.Until.Scn != nil:
			key = "SCN"
			limit, err = operation.RestoreUntilArgument(key, *restoreSpec.Until.Scn)
			if err != nil {
				return err
			}
		default:
			return errors.New("restore cutoff must specify TIME or SCN")
		}
	}
	if restoreSpec.BakEncryptionSecret != "" {
		password, err := resourceutils.ReadPassword(m.Client, m.Resource.Namespace, restoreSpec.BakEncryptionSecret)
		if err != nil {
			m.Recorder.Event(m.Resource, v1.EventTypeWarning, "ReadRestorePasswordFailed", err.Error())
			return err
		}
		err = con.SetRestorePassword(m.Ctx, password)
		if err != nil {
			m.Recorder.Event(m.Resource, v1.EventTypeWarning, "SetRestorePasswordFailed", err.Error())
			return err
		}
	}

	claimed, err := m.claimRestoreSubmission()
	if err != nil {
		return err
	}
	if !claimed {
		return m.awaitSubmittedRestore(con, description)
	}
	err = con.StartTrackedRestore(m.Ctx, m.Resource.Spec.TargetTenant, sourceUri, m.Resource.Spec.Option, key, limit, description)
	if err != nil {
		var sqlErr *mysql.MySQLError
		if errors.As(err, &sqlErr) {
			return err
		}
		// A disconnected client can miss a successful server reply.
		if observed := m.awaitSubmittedRestore(con, description); observed != nil {
			return errors.Wrapf(err, "no matching physical restore confirmed: %v", observed)
		}
	}

	return nil
}

func StartLogReplay(m *ObTenantRestoreManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	if m.Resource.Spec.PrimaryTenant != nil {
		restoreSource, err := resourceutils.GetTenantRestoreSource(m.Ctx, m.Client, m.Logger, m.Resource.Namespace, *m.Resource.Spec.PrimaryTenant)
		if err != nil {
			return err
		}
		err = con.SetParameter(m.Ctx, "LOG_RESTORE_SOURCE", restoreSource, &param.Scope{
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
		err = con.ReplayStandbyLog(m.Ctx, m.Resource.Spec.TargetTenant, "UNLIMITED")
	} else if replayUntil.Timestamp != nil {
		err = con.ReplayStandbyLog(m.Ctx, m.Resource.Spec.TargetTenant, fmt.Sprintf("TIME='%s'", *replayUntil.Timestamp))
	} else if replayUntil.Scn != nil {
		err = con.ReplayStandbyLog(m.Ctx, m.Resource.Spec.TargetTenant, fmt.Sprintf("SCN=%s", *replayUntil.Scn))
	} else {
		return errors.New("Replay until with limit must have a limit key, scn and timestamp are both nil now")
	}
	return err
}

func ActivateStandby(m *ObTenantRestoreManager) tasktypes.TaskError {
	con, err := m.getClusterSysClient()
	if err != nil {
		return err
	}
	return con.ActivateStandby(m.Ctx, m.Resource.Spec.TargetTenant)
}
