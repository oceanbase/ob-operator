package operation

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

type dropTestOperations struct {
	zoneStatus string
	servers    int
	rows       [][]zoneStorageDestination
	calls      int
	executed   [][]any
	execErr    error
}

func (f *dropTestOperations) GetZone(_ context.Context, _ string) (*model.OBZone, error) {
	return &model.OBZone{Status: f.zoneStatus}, nil
}
func (f *dropTestOperations) QueryCount(_ context.Context, n *int, query string, args ...any) error {
	if query != sql.CountZoneServers || !reflect.DeepEqual(args, []any{"zone4"}) {
		return errors.New("unscoped count")
	}
	*n = f.servers
	return nil
}
func (f *dropTestOperations) QueryList(_ context.Context, dest any, query string, args ...any) error {
	if query != sql.ListZoneStorageDest || !reflect.DeepEqual(args, []any{"zone4"}) {
		return errors.New("unscoped list")
	}
	i := f.calls
	if i >= len(f.rows) {
		i = len(f.rows) - 1
	}
	*dest.(*[]zoneStorageDestination) = f.rows[i]
	f.calls++
	return nil
}
func (f *dropTestOperations) ExecWithDefaultTimeout(_ context.Context, query string, args ...any) error {
	if query != sql.DropSharedStorageDest {
		return errors.New("unexpected DDL")
	}
	f.executed = append(f.executed, args)
	return f.execErr
}

func TestDropZoneSharedStorage(t *testing.T) {
	ready := zoneStorageDestination{Path: "s3://bucket", Endpoint: "host=http://minio:9000", UsedFor: "ALL", State: "ADDED"}
	pending := ready
	pending.State = "DROPPING"
	for _, tc := range []struct {
		name       string
		ops        dropTestOperations
		wantErr    bool
		executions int
	}{
		{"canonical path and wait", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{ready}, {pending}, nil}}, false, 1},
		{"already missing", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{nil}}, false, 0},
		{"resume dropping", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{pending}, nil}}, false, 0},
		{"active zone", dropTestOperations{zoneStatus: "active"}, true, 0},
		{"nonempty zone", dropTestOperations{zoneStatus: "inactive", servers: 1}, true, 0},
		{"DDL failure", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{ready}}, execErr: errors.New("DDL failed")}, true, 1},
		{"timeout", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{pending}}}, true, 0},
		{"invalid usage", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{{Path: "s3://bucket", UsedFor: "DATA", State: "ADDED"}}}}, true, 0},
		{"add in progress", dropTestOperations{zoneStatus: "inactive", rows: [][]zoneStorageDestination{{{Path: "s3://bucket", UsedFor: "ALL", State: "ADDING"}}}}, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := dropZoneSharedStorage(context.Background(), &tc.ops, "zone4", time.Millisecond, 20*time.Millisecond)
			if (err != nil) != tc.wantErr || len(tc.ops.executed) != tc.executions {
				t.Fatalf("err=%v executions=%d", err, len(tc.ops.executed))
			}
			for _, args := range tc.ops.executed {
				if !reflect.DeepEqual(args, []any{"s3://bucket?host=http://minio:9000", "zone4"}) {
					t.Fatalf("unexpected binding %v", args)
				}
			}
		})
	}
}
