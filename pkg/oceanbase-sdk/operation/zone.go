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
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	zonestatus "github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/status/zone"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

const (
	sharedStorageDestPollInterval = time.Second
	sharedStorageDestWaitTimeout  = 30 * time.Second
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

func (m *OceanbaseOperationManager) AddSharedStorageDest(ctx context.Context, path, accessInfo, attribute, zoneName string) error {
	return m.addSharedStorageDest(
		ctx,
		path,
		accessInfo,
		attribute,
		zoneName,
		sharedStorageDestPollInterval,
		sharedStorageDestWaitTimeout,
	)
}

func (m *OceanbaseOperationManager) addSharedStorageDest(
	ctx context.Context,
	path, accessInfo, attribute, zoneName string,
	pollInterval, waitTimeout time.Duration,
) error {
	status, states, err := m.getSharedStorageDestStatus(ctx, zoneName)
	if err != nil {
		return errors.Wrap(err, "Check shared storage destination")
	}
	switch status {
	case sharedStorageDestReady:
		m.Logger.Info("Shared storage destination already exists, skip add", "zone", zoneName)
		return nil
	case sharedStorageDestPending:
		m.Logger.Info("Shared storage destination is being prepared, wait until ready", "zone", zoneName)
		return m.waitSharedStorageDestReady(ctx, zoneName, pollInterval, waitTimeout)
	case sharedStorageDestInvalid:
		return sharedStorageDestNotReadyError(zoneName, states)
	}

	execErr := m.ExecWithDefaultTimeout(ctx, sql.AddSharedStorageDest, path, accessInfo, attribute, zoneName)
	// The DDL may have committed even when the client observed a timeout or
	// connection error. Wait through the normal ADDING/ROTATING transition so
	// the flow cannot start the zone early or report a transient Task failed
	// warning for a successful expansion.
	waitErr := m.waitSharedStorageDestReady(ctx, zoneName, pollInterval, waitTimeout)
	if waitErr == nil {
		if execErr != nil {
			m.Logger.Info("Shared storage destination became ready after execution error, treat as successful", "zone", zoneName)
		}
		return nil
	}
	if execErr != nil {
		m.Logger.Error(waitErr, "Shared storage destination did not become ready after execution error", "zone", zoneName)
		return errors.Wrap(execErr, "Add shared storage destination")
	}
	return errors.Wrap(waitErr, "Wait shared storage destination after add")
}

type sharedStorageDestRecord struct {
	State string `db:"state"`
}

type sharedStorageDestStatus int

const (
	sharedStorageDestMissing sharedStorageDestStatus = iota
	sharedStorageDestPending
	sharedStorageDestReady
	sharedStorageDestInvalid
)

func (m *OceanbaseOperationManager) getSharedStorageDestStatus(ctx context.Context, zoneName string) (sharedStorageDestStatus, []string, error) {
	records := make([]sharedStorageDestRecord, 0, 1)
	err := m.QueryList(ctx, &records, sql.ListZoneStorageState, zoneName)
	if err != nil {
		m.Logger.Error(err, "Got exception when query shared storage destination", "zone", zoneName)
		return sharedStorageDestMissing, nil, errors.Wrap(err, "Query shared storage destination")
	}
	if len(records) == 0 {
		return sharedStorageDestMissing, nil, nil
	}

	status := sharedStorageDestReady
	states := make([]string, 0, len(records))
	for _, record := range records {
		state := strings.ToUpper(strings.TrimSpace(record.State))
		states = append(states, state)
		switch state {
		case "ADDED", "ROTATED":
		case "ADDING", "ROTATING":
			if status != sharedStorageDestInvalid {
				status = sharedStorageDestPending
			}
		default:
			status = sharedStorageDestInvalid
		}
	}
	return status, states, nil
}

func sharedStorageDestNotReadyError(zoneName string, states []string) error {
	if len(states) == 0 {
		return errors.Errorf("Shared storage destination metadata for zone %s not found", zoneName)
	}
	return errors.Errorf(
		"Shared storage destination for zone %s is not ready (states: %s)",
		zoneName,
		strings.Join(states, ","),
	)
}

func (m *OceanbaseOperationManager) waitSharedStorageDestReady(
	ctx context.Context,
	zoneName string,
	pollInterval, waitTimeout time.Duration,
) error {
	if pollInterval <= 0 || waitTimeout <= 0 {
		return errors.New("Shared storage destination wait interval and timeout must be positive")
	}

	waitCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		status, states, err := m.getSharedStorageDestStatus(waitCtx, zoneName)
		if err != nil {
			lastErr = err
		} else {
			switch status {
			case sharedStorageDestReady:
				return nil
			case sharedStorageDestInvalid:
				return sharedStorageDestNotReadyError(zoneName, states)
			default:
				lastErr = sharedStorageDestNotReadyError(zoneName, states)
			}
		}

		select {
		case <-waitCtx.Done():
			if ctx.Err() != nil {
				return errors.Wrap(ctx.Err(), "Wait shared storage destination canceled")
			}
			return errors.Wrapf(lastErr, "Wait shared storage destination timed out after %s", waitTimeout)
		case <-ticker.C:
		}
	}
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
