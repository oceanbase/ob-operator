package operation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

type activationRecorder struct {
	t        *testing.T
	state    model.OBTenant
	readErr  error
	writeErr error
	writes   int
}

func (r *activationRecorder) QueryRow(_ context.Context, dest any, query string, args ...any) error {
	r.t.Helper()
	if query != sql.GetActivationState || len(args) != 1 || args[0] != "db_tenant" {
		r.t.Fatalf("unexpected role query %q %v", query, args)
	}
	*dest.(*model.OBTenant) = r.state
	return r.readErr
}

func (r *activationRecorder) ExecWithTimeout(_ context.Context, timeout time.Duration, query string, args ...any) error {
	r.t.Helper()
	if timeout != managerConfig.TenantSqlTimeout || query != sql.ActivateStandby || len(args) != 1 || args[0] != "db_tenant" {
		r.t.Fatalf("unexpected activation %s %q %v", timeout, query, args)
	}
	r.writes++
	return r.writeErr
}

func TestActivateStandby(t *testing.T) {
	dbErr := errors.New("database rejected the statement")
	for _, tc := range []struct {
		name, role, status, switchover string
		readErr, writeErr              error
		cancel, wantError              bool
		writes                         int
	}{
		{name: "already activated", role: "PRIMARY", status: "NORMAL", switchover: "NORMAL"},
		{name: "primary still switching", role: "PRIMARY", status: "NORMAL", switchover: "SWITCHING TO PRIMARY", wantError: true},
		{name: "unavailable primary", role: "PRIMARY", status: "DROPPING", switchover: "NORMAL", wantError: true},
		{name: "restoring tenant", role: "RESTORE", status: "RESTORE", wantError: true},
		{name: "activate standby", role: "STANDBY", status: "NORMAL", switchover: "NORMAL", writes: 1},
		{name: "cannot read role", readErr: dbErr, wantError: true},
		{name: "database rejected DDL", role: "STANDBY", status: "NORMAL", writeErr: dbErr, writes: 1, wantError: true},
		{name: "canceled caller", role: "PRIMARY", status: "NORMAL", switchover: "NORMAL", cancel: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				cancel()
			}
			op := &activationRecorder{t: t, state: model.OBTenant{TenantRole: tc.role, Status: tc.status, SwitchoverStatus: tc.switchover}, readErr: tc.readErr, writeErr: tc.writeErr}
			err := activateStandby(ctx, op, "db_tenant")
			if (err != nil) != tc.wantError || op.writes != tc.writes {
				t.Fatalf("err=%v writes=%d", err, op.writes)
			}
			if (tc.readErr != nil || tc.writeErr != nil) && !errors.Is(err, dbErr) {
				t.Fatalf("lost database error: %v", err)
			}
			if tc.cancel && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
		})
	}
}
