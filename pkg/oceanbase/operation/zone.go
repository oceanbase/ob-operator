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
	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase/const/sql"
	zonestatus "github.com/oceanbase/ob-operator/pkg/oceanbase/const/status/zone"
	"github.com/oceanbase/ob-operator/pkg/oceanbase/model"
)

func (m *OceanbaseOperationManager) AddZone(zoneName string) error {
	_, err := m.GetZone(zoneName)
	// TODO verify it's a not found error
	if err == nil {
		m.Logger.Info("Obzone already exists in observer, skip add", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(sql.AddZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when add zone")
		return errors.Wrap(err, "Add zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) DeleteZone(zoneName string) error {
	obzone, err := m.GetZone(zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status != zonestatus.Inactive {
		m.Logger.Info("Obzone is not inactive, stop it before delete", "zone", zoneName)
		return errors.Errorf("Obzone %s is not inactive, stop it before delete", zoneName)
	}
	err = m.ExecWithDefaultTimeout(sql.DeleteZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when delete zone")
		return errors.Wrap(err, "Delete zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) GetZone(zoneName string) (*model.OBZone, error) {
	obzoneInfoList := make([]model.OBZoneInfo, 0)
	err := m.QueryList(&obzoneInfoList, sql.GetZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when query zone info")
		return nil, errors.Wrap(err, "Query zone info")
	}
	if len(obzoneInfoList) == 0 {
		return nil, errors.Errorf("Query obzone %s info get empty result", zoneName)
	}
	obzone := model.NewOBZone(zoneName, obzoneInfoList)
	return obzone, nil
}

func (m *OceanbaseOperationManager) StartZone(zoneName string) error {
	obzone, err := m.GetZone(zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status == zonestatus.Active {
		m.Logger.Info("Obzone already active", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(sql.StartZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when start zone")
		return errors.Wrap(err, "Start zone")
	}
	return nil
}

func (m *OceanbaseOperationManager) StopZone(zoneName string) error {
	obzone, err := m.GetZone(zoneName)
	if err != nil {
		m.Logger.Error(err, "Query obzone failed")
		return errors.Wrapf(err, "Query obzone %s failed", zoneName)
	}
	if obzone.Status == zonestatus.Inactive {
		m.Logger.Info("Obzone already inactive", "zone", zoneName)
		return nil
	}
	err = m.ExecWithDefaultTimeout(sql.StopZone, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when start zone")
		return errors.Wrap(err, "Start zone")
	}
	return nil
}
