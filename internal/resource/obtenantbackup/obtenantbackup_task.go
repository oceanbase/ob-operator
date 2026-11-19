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

package obtenantbackup

import (
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task_register $GOFILE

var taskMap = builder.NewTaskHub[*OBTenantBackupManager]()

func CreateBackupJobInOB(m *OBTenantBackupManager) tasktypes.TaskError {
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		m.Logger.Error(err, "failed to get ob operation client")
		return err
	}
	submit := func(db *operation.OceanbaseOperationManager) error {
		_, err := db.CreateAndReturnBackupJob(m.Ctx, string(job.Spec.Type))
		return err
	}
	if job.Spec.EncryptionSecret != "" {
		password, readErr := resourceutils.ReadBackupEncryptionPassword(m.Client, job.Namespace, job.Spec.EncryptionSecret)
		if readErr != nil {
			m.Recorder.Event(job, "Warning", "ReadBackupEncryptionSecretFailed", readErr.Error())
			return readErr
		}
		err = con.WithBackupSession(m.Ctx, func(db *operation.OceanbaseOperationManager) error {
			if setErr := db.SetBackupPassword(m.Ctx, password); setErr != nil {
				m.Recorder.Event(job, "Warning", "SetBackupPasswordFailed", setErr.Error())
				return setErr
			}
			return submit(db)
		})
	} else {
		err = submit(con)
	}
	if err != nil {
		m.Logger.Error(err, "failed to create and return backup job")
		m.Recorder.Event(job, "Warning", "CreateAndReturnBackupJobFailed", err.Error())
		return err
	}

	m.Recorder.Event(job, "Create", "", "Create backup job successfully")
	return nil
}
