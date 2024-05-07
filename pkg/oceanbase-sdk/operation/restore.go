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
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/config"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) SetRestorePassword(ctx context.Context, password string) error {
	err := m.ExecWithDefaultTimeout(ctx, sql.SetRestorePassword, password)
	if err != nil {
		m.Logger.Error(err, "Got exception when set restore password")
		return errors.Wrap(err, "Set restore password")
	}
	return nil
}

func (m *OceanbaseOperationManager) StartRestoreWithLimit(ctx context.Context, tenantName, uri, restoreOption string, limitKey, limitValue any) error {
	sqlStatement := fmt.Sprintf(sql.StartRestoreWithLimit, tenantName, limitKey)
	err := m.ExecWithTimeout(ctx, config.TenantRestoreTimeOut, sqlStatement, uri, limitValue, restoreOption)
	if err != nil {
		m.Logger.Error(err, "Got exception when start restore with limit")
		return errors.Wrap(err, "Start restore with limit")
	}
	return nil
}

func (m *OceanbaseOperationManager) StartRestoreUnlimited(ctx context.Context, tenantName, uri, restoreOption string) error {
	err := m.ExecWithTimeout(ctx, config.TenantRestoreTimeOut, fmt.Sprintf(sql.StartRestoreUnlimited, tenantName), uri, restoreOption)
	if err != nil {
		m.Logger.Error(err, "Got exception when start restore unlimited")
		return errors.Wrap(err, "Start restore unlimited")
	}
	return nil
}

func (m *OceanbaseOperationManager) CancelRestoreOfTenant(ctx context.Context, tenantName string) error {
	err := m.ExecWithDefaultTimeout(ctx, fmt.Sprintf(sql.CancelRestore, tenantName))
	if err != nil {
		m.Logger.Error(err, "Got exception when cancel restore of tenant")
		return errors.Wrap(err, "Cancel restore of tenant")
	}
	return nil
}

func (m *OceanbaseOperationManager) ReplayStandbyLog(ctx context.Context, tenantName, untilLimit string) error {
	sqlStatement := fmt.Sprintf(sql.ReplayStandbyLog, untilLimit)
	err := m.ExecWithDefaultTimeout(ctx, sqlStatement, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when replay standby log")
		return errors.Wrap(err, "Replay standby log")
	}
	return nil
}

func (m *OceanbaseOperationManager) ActivateStandby(ctx context.Context, tenantName string) error {
	err := m.ExecWithDefaultTimeout(ctx, sql.ActivateStandby, tenantName)
	if err != nil {
		m.Logger.Error(err, "Got exception when activate standby")
		return errors.Wrap(err, "Activate standby")
	}
	return nil
}

func (m *OceanbaseOperationManager) ListRestoreProgress(ctx context.Context) ([]*model.RestoreProgress, error) {
	progressInfos := make([]*model.RestoreProgress, 0)
	err := m.QueryList(ctx, &progressInfos, sql.QueryRestoreProgress)
	if err != nil {
		m.Logger.Error(err, "Got exception when query restore progress")
		return nil, errors.Wrap(err, "List restore progress")
	}
	return progressInfos, nil
}

func (m *OceanbaseOperationManager) ListRestoreHistory(ctx context.Context) ([]*model.RestoreHistory, error) {
	restoreHistory := make([]*model.RestoreHistory, 0)
	err := m.QueryList(ctx, &restoreHistory, sql.QueryRestoreHistory)
	if err != nil {
		m.Logger.Error(err, "Got exception when query restore history")
		return nil, errors.Wrap(err, "List restore history")
	}
	return restoreHistory, nil
}

func (m *OceanbaseOperationManager) GetLatestRestoreProgressOfTenant(ctx context.Context, tenant string) (*model.RestoreProgress, error) {
	latest := make([]*model.RestoreProgress, 0)
	err := m.QueryList(ctx, &latest, sql.GetLatestRestoreProgress, tenant)
	if err != nil {
		m.Logger.Error(err, "Got exception when query latest restore progress")
		return nil, errors.Wrap(err, "Get latest restore progress")
	}
	if len(latest) == 0 {
		return nil, nil
	}
	return latest[0], nil
}

func (m *OceanbaseOperationManager) GetLatestRestoreHistoryOfTenant(ctx context.Context, tenant string) (*model.RestoreHistory, error) {
	latest := make([]*model.RestoreHistory, 0)
	err := m.QueryList(ctx, &latest, sql.GetLatestRestoreHistory, tenant)
	if err != nil {
		m.Logger.Error(err, "Got exception when query latest restore history")
		return nil, errors.Wrap(err, "Get latest restore history")
	}
	if len(latest) == 0 {
		return nil, nil
	}
	return latest[0], nil
}
