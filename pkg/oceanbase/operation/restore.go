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

package operation

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

func (m *OceanbaseOperationManager) SetRestorePassword(password string) error {
	err := m.ExecWithDefaultTimeout(sql.SetRestorePassword, password)
	if err != nil {
		m.Logger.Error(err, "Got exception when set restore password")
		return errors.Wrap(err, "Set restore password")
	}
	return nil
}

func (m *OceanbaseOperationManager) StartRestoreWithLimit(tenantName, uri, restoreOption string, limitKey, limitValue any) error {
	sqlStatement := fmt.Sprintf(sql.StartRestoreWithLimit, tenantName, limitKey)
	err := m.ExecWithTimeout(config.TenantRestoreTimeOut, sqlStatement, uri, limitValue, restoreOption)
	if err != nil {
		m.Logger.Error(err, "Got exception when start restore with limit")
		return errors.Wrap(err, "Start restore with limit")
	}
	return nil
}

func (m *OceanbaseOperationManager) StartRestoreUnlimited(tenantName, uri, restoreOption string) error {
	err := m.ExecWithTimeout(config.TenantRestoreTimeOut, fmt.Sprintf(sql.StartRestoreUnlimited, tenantName), uri, restoreOption)
	if err != nil {
		m.Logger.Error(err, "Got exception when start restore unlimited")
		return errors.Wrap(err, "Start restore unlimited")
	}
	return nil
}

func (m *OceanbaseOperationManager) CancelRestoreOfTenant(tenantName string) error {
	err := m.ExecWithDefaultTimeout(fmt.Sprintf(sql.CancelRestore, tenantName))
	if err != nil {
		m.Logger.Error(err, "Got exception when cancel restore of tenant")
		return errors.Wrap(err, "Cancel restore of tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) ReplayStandbyLog(tenantName, untilLimit string) error {
	sqlStatement := fmt.Sprintf(sql.ReplayStandbyLog, untilLimit)
	err := m.ExecWithDefaultTimeout(sqlStatement, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when replay standby log")
		return errors.Wrap(err, "Replay standby log")
	}
	return nil
}

func (m *OceanbaseOperationManager) ActivateStandby(tenantName string) error {
	err := m.ExecWithDefaultTimeout(sql.ActivateStandby, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when activate standby")
		return errors.Wrap(err, "Activate standby")
	}
	return nil
}

func (m *OceanbaseOperationManager) ListRestoreProgress() ([]*model.RestoreProgress, error) {
	progressInfos := make([]*model.RestoreProgress, 0)
	err := m.QueryList(&progressInfos, sql.QueryRestoreProgress)
	if err != nil {
		m.Logger.Error(err, "Got exception when query restore progress")
		return nil, errors.Wrap(err, "List restore progress")
	}
	return progressInfos, nil
}

func (m *OceanbaseOperationManager) ListRestoreHistory() ([]*model.RestoreHistory, error) {
	restoreHistory := make([]*model.RestoreHistory, 0)
	err := m.QueryList(&restoreHistory, sql.QueryRestoreHistory)
	if err != nil {
		m.Logger.Error(err, "Got exception when query restore history")
		return nil, errors.Wrap(err, "List restore history")
	}
	return restoreHistory, nil
}

func (m *OceanbaseOperationManager) GetLatestRestoreProgressOfTenant(tenant string) (*model.RestoreProgress, error) {
	latest := make([]*model.RestoreProgress, 0)
	err := m.QueryList(&latest, sql.GetLatestRestoreProgress, tenant)
	if err != nil {
		m.Logger.Error(err, "Got exception when query latest restore progress")
		return nil, errors.Wrap(err, "Get latest restore progress")
	}
	if len(latest) == 0 {
		return nil, nil
	}
	return latest[0], nil
}

func (m *OceanbaseOperationManager) GetLatestRestoreHistoryOfTenant(tenant string) (*model.RestoreHistory, error) {
	latest := make([]*model.RestoreHistory, 0)
	err := m.QueryList(&latest, sql.GetLatestRestoreHistory, tenant)
	if err != nil {
		m.Logger.Error(err, "Got exception when query latest restore history")
		return nil, errors.Wrap(err, "Get latest restore history")
	}
	if len(latest) == 0 {
		return nil, nil
	}
	return latest[0], nil
}
