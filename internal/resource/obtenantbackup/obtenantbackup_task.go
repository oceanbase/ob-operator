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
	"github.com/oceanbase/ob-operator/pkg/task/builder"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

//go:generate task-register $GOFILE

var taskMap = builder.NewTaskHub[*OBTenantBackupManager]()

func CreateBackupJobInOB(m *OBTenantBackupManager) tasktypes.TaskError {
	job := m.Resource
	con, err := m.getObOperationClient()
	if err != nil {
		m.Logger.Error(err, "failed to get ob operation client")
		return err
	}
	if job.Spec.EncryptionSecret != "" {
		password, err := resourceutils.ReadPassword(m.Client, job.Namespace, job.Spec.EncryptionSecret)
		if err != nil {
			m.Logger.Error(err, "failed to read backup encryption secret")
			m.Recorder.Event(job, "Warning", "ReadBackupEncryptionSecretFailed", err.Error())
		} else if password != "" {
			err = con.SetBackupPassword(password)
			if err != nil {
				m.Logger.Error(err, "failed to set backup password")
				m.Recorder.Event(job, "Warning", "SetBackupPasswordFailed", err.Error())
			}
		}
	}
	_, err = con.CreateAndReturnBackupJob(job.Spec.Type)
	if err != nil {
		m.Logger.Error(err, "failed to create and return backup job")
		m.Recorder.Event(job, "Warning", "CreateAndReturnBackupJobFailed", err.Error())
		return err
	}

	// job.Status.BackupJob = latest
	m.Recorder.Event(job, "Create", "", "create backup job successfully")
	return nil
}
