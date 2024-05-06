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
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	zonestatus "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/status/zone"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

func (m *OceanbaseOperationManager) AddZone(ctx context.Context, zoneName string) error {
	_, err := m.GetZone(ctx, zoneName)
	// TODO verify it's a not found error
	if err == nil {
		m.Logger.Info("OBZone already exists in observer, skip add", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(ctx, sql.AddZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when add zone")
		return errors.Wrap(err, "Add zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteZone(ctx context.Context, zoneName string) error {
	obzone, err := m.GetZone(ctx, zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status != zonestatus.Inactive {
		m.Logger.Info("OBZone is not inactive, stop it before delete", "zone", zoneName)
		return errors.Errorf("OBZone %s is not inactive, stop it before delete", zoneName)
	}
	err = m.ExecWithDefaultTimeout(ctx, sql.DeleteZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete zone")
		return errors.Wrap(err, "Delete zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) ListZones(ctx context.Context) ([]model.OBZone, error) {
	zoneList := make([]model.OBZone, 0)
	err := m.QueryList(ctx, &zoneList, sql.ListZones)
	if err != nil {
		m.Logger.Error(err, "Got exception when list all zone")
		return nil, errors.Wrap(err, "list all zone")
	}
	return zoneList, nil
}

func (m *OceanbaseOperationManager) GetZone(ctx context.Context, zoneName string) (*model.OBZone, error) {
	zone := &model.OBZone{}
	err := m.QueryRow(ctx, zone, sql.GetZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when query zone")
		return nil, errors.Wrap(err, "query zone info")
	}
	return zone, nil
}

func (m *OceanbaseOperationManager) StartZone(ctx context.Context, zoneName string) error {
	obzone, err := m.GetZone(ctx, zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status == zonestatus.Active {
		m.Logger.Info("OBZone already active", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(ctx, sql.StartZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when start zone")
		return errors.Wrap(err, "Start zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) StopZone(ctx context.Context, zoneName string) error {
	obzone, err := m.GetZone(ctx, zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status == zonestatus.Inactive {
		m.Logger.Info("OBZone already inactive", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(ctx, sql.StopZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when stop zone")
		return errors.Wrap(err, "Start zone")
	}
	return nil
}
