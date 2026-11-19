package obtenantoperation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
)

type upgradeProbe struct {
	active    [][]int64
	tenantID  int64
	reads     int
	writes    int
	params    [][]*model.Parameter
	submitErr error
}

func (p *upgradeProbe) ListTenantWithName(context.Context, string) ([]*model.OBTenant, error) {
	return []*model.OBTenant{{TenantID: p.tenantID, TenantRole: "PRIMARY", Status: "NORMAL"}}, nil
}
func (p *upgradeProbe) SelectCompatibleOfTenants(context.Context) ([]*model.Parameter, error) {
	i := min(p.reads, len(p.params)-1)
	p.reads++
	return p.params[i], nil
}
func (p *upgradeProbe) ListActiveTenantUpgradeIDs(context.Context) ([]int64, error) {
	if len(p.active) == 0 {
		return nil, nil
	}
	return p.active[min(p.reads-1, len(p.active)-1)], nil
}
func (p *upgradeProbe) UpgradeTenantWithName(context.Context, string) error {
	p.writes++
	return p.submitErr
}

func TestTenantUpgradeConfirmation(t *testing.T) {
	sys := &model.Parameter{TenantID: 1, Value: "4.10.0.0"}
	old := &model.Parameter{TenantID: 1002, Value: "4.2.1.0"}
	current := &model.Parameter{TenantID: 1002, Value: "4.10.0.0"}
	for _, tc := range []struct {
		name       string
		expectedID int64
		rows       [][]*model.Parameter
		submitErr  error
		success    bool
		writes     int
	}{
		{"restore snapshot had no ID", 0, [][]*model.Parameter{{sys, old}, {sys, current, old}, {sys, current, current}}, nil, true, 1},
		{"already upgraded", 1002, [][]*model.Parameter{{sys, current, current}}, nil, true, 0},
		{"in progress on another observer", 1002, [][]*model.Parameter{{sys, old, current}, {sys, current, current}}, nil, true, 0},
		{"different physical tenant", 1003, [][]*model.Parameter{{sys, current}}, nil, false, 0},
		{"missing target rows", 1002, [][]*model.Parameter{{sys}}, nil, false, 0},
		{"never converges", 0, [][]*model.Parameter{{sys, old}}, nil, false, 1},
		{"ambiguous reply after success", 0, [][]*model.Parameter{{sys, old}, {sys, current}}, errors.New("reply lost"), true, 1},
		{"permission rejected", 0, [][]*model.Parameter{{sys, old}, {sys, current}}, &mysql.MySQLError{Number: 1227}, false, 1},
		{"inconsistent sys rows", 0, [][]*model.Parameter{{sys, {TenantID: 1, Value: "4.2.1.1"}, current}}, nil, false, 0},
		{"unsupported system version", 0, [][]*model.Parameter{{{TenantID: 1, Value: "4.0.0.0"}, old}}, nil, false, 0},
		{"future tenant version", 0, [][]*model.Parameter{{sys, {TenantID: 1002, Value: "4.11.0.0"}}}, nil, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			p := &upgradeProbe{tenantID: 1002, params: tc.rows, submitErr: tc.submitErr}
			err := upgradeAndWait(ctx, p, "target", tc.expectedID, time.Millisecond)
			if (err == nil) != tc.success || p.writes != tc.writes {
				t.Fatalf("err=%v writes=%d; want success=%t writes=%d", err, p.writes, tc.success, tc.writes)
			}
		})
	}
}

func TestTenantUpgradeAdoptsRunningJob(t *testing.T) {
	sys := &model.Parameter{TenantID: 1, Value: "4.2.1.1"}
	old := &model.Parameter{TenantID: 1002, Value: "4.2.0.0"}
	newer := &model.Parameter{TenantID: 1002, Value: sys.Value}
	for _, tc := range []struct {
		name     string
		active   [][]int64
		writeErr error
		writes   int
	}{
		{"previous leader", [][]int64{{1002}, {1002}, {}}, nil, 0},
		{"another tenant", [][]int64{{1003}, {}, {}}, nil, 1},
		{"executor acquired concurrently", nil, &mysql.MySQLError{Number: 4179}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			p := &upgradeProbe{tenantID: 1002, active: tc.active, params: [][]*model.Parameter{{sys, old}, {sys, old}, {sys, newer}}, submitErr: tc.writeErr}
			if err := upgradeAndWait(ctx, p, "target", 0, time.Millisecond); err != nil || p.writes != tc.writes {
				t.Fatalf("err=%v writes=%d want=%d", err, p.writes, tc.writes)
			}
		})
	}
}
