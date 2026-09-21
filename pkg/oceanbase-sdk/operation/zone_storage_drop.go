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

type zoneStorageDestination struct {
	Path     string `db:"path"`
	Endpoint string `db:"endpoint"`
	UsedFor  string `db:"used_for"`
	State    string `db:"state"`
}

type zoneStorageDropOperations interface {
	GetZone(context.Context, string) (*model.OBZone, error)
	QueryCount(context.Context, *int, string, ...any) error
	QueryList(context.Context, any, string, ...any) error
	ExecWithDefaultTimeout(context.Context, string, ...any) error
}

// DropZoneSharedStorage detaches storage from one empty, stopped SS zone. It
// does not remove a bucket or delete objects shared with other zones.
func (m *OceanbaseOperationManager) DropZoneSharedStorage(ctx context.Context, zone string) error {
	return dropZoneSharedStorage(ctx, m, zone, sharedStorageDestPollInterval, sharedStorageDestWaitTimeout)
}

func dropZoneSharedStorage(ctx context.Context, op zoneStorageDropOperations, zone string, interval, timeout time.Duration) error {
	if strings.TrimSpace(zone) == "" || interval <= 0 || timeout <= 0 {
		return errors.New("zone and positive storage-drop wait durations are required")
	}
	z, err := op.GetZone(ctx, zone)
	if err != nil {
		return err
	}
	if z.Status != zonestatus.Inactive {
		return errors.New("stop zone before dropping shared storage")
	}
	var count int
	if err := op.QueryCount(ctx, &count, sql.CountZoneServers, zone); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("remove all servers before dropping zone shared storage")
	}
	list := func(ctx context.Context) ([]zoneStorageDestination, error) {
		var destinations []zoneStorageDestination
		err := op.QueryList(ctx, &destinations, sql.ListZoneStorageDest, zone)
		return destinations, err
	}
	destinations, err := list(ctx)
	if err != nil || len(destinations) == 0 {
		return err
	}
	// Validate every row before issuing any DDL; do not broaden unknown storage
	// usage types to ALL or interfere with an in-progress add/rotation.
	for _, dest := range destinations {
		if dest.Path == "" || dest.UsedFor != "ALL" {
			return errors.New("unexpected shared storage destination; refusing automatic drop")
		}
		switch dest.State {
		case "ADDED", "ROTATED", "DROPPING":
		default:
			return errors.New("shared storage destination is not ready for removal")
		}
	}
	for _, dest := range destinations {
		if dest.State == "DROPPING" {
			continue
		}
		// DROP matches the complete suffix after '?' against ENDPOINT. Reusing
		// the creation URL (e.g. including s3_region) yields Entry not exist.
		path := dest.Path
		if dest.Endpoint != "" {
			path += "?" + dest.Endpoint
		}
		if err := op.ExecWithDefaultTimeout(ctx, sql.DropSharedStorageDest, path, zone); err != nil {
			return err
		}
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		destinations, err = list(waitCtx)
		if err != nil {
			return err
		}
		if len(destinations) == 0 {
			return nil
		}
		select {
		case <-waitCtx.Done():
			return errors.Wrap(waitCtx.Err(), "Wait for zone shared storage removal")
		case <-ticker.C:
		}
	}
}
