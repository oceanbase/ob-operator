/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package store

import (
	"context"
	"testing"
	"time"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
)

func TestParsePlanGeneratedTimeUsesDatabaseLocalTimezone(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	got, err := parsePlanGeneratedTime("2026-09-29 09:41:20", location)
	if err != nil {
		t.Fatalf("parsePlanGeneratedTime() error = %v", err)
	}
	want := time.Date(2026, time.September, 29, 9, 41, 20, 0, location)
	if !got.Equal(want) {
		t.Fatalf("parsePlanGeneratedTime() = %s (%d), want %s (%d)", got, got.Unix(), want, want.Unix())
	}
}

func TestParsePlanGeneratedTimeRejectsInvalidValue(t *testing.T) {
	if _, err := parsePlanGeneratedTime("invalid", time.UTC); err == nil {
		t.Fatal("parsePlanGeneratedTime() error = nil, want parse error")
	}
}

func TestQuerySqlDetailInfoUsesLocalTimezoneForPlanGeneratedTime(t *testing.T) {
	auditStore := newSchemaTestStore(t)
	writeSchemaParquet(t, auditStore, "2026-09-29-09-41-20-current.parquet", false,
		schemaAuditRow{id: "timezone-sql", query: "SELECT 42", executions: 1})

	planStore, err := NewPlanStore(context.Background(), t.TempDir(), 2, 1, auditStore.Logger)
	if err != nil {
		t.Fatalf("NewPlanStore() error = %v", err)
	}
	t.Cleanup(planStore.Close)
	if err := planStore.InitSqlPlanTable(); err != nil {
		t.Fatalf("InitSqlPlanTable() error = %v", err)
	}
	if err := planStore.Store(model.SqlPlan{
		TenantID: 1, SvrIP: "127.0.0.1", SvrPort: 2882, PlanID: 1,
		SqlID: "timezone-sql", PlanHash: 1, ID: 0, GmtCreate: "2026-09-29 09:41:20",
	}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	originalLocation := time.Local
	location := time.FixedZone("UTC+8", 8*60*60)
	time.Local = location
	t.Cleanup(func() { time.Local = originalLocation })

	got, err := auditStore.QuerySqlDetailInfo(planStore, apimodel.SqlDetailRequest{
		StartTime: 99, EndTime: 101, SqlId: "timezone-sql",
	})
	if err != nil {
		t.Fatalf("QuerySqlDetailInfo() error = %v", err)
	}
	if len(got.Plans) != 1 {
		t.Fatalf("QuerySqlDetailInfo() returned %d plans, want 1", len(got.Plans))
	}
	want := time.Date(2026, time.September, 29, 9, 41, 20, 0, location).Unix()
	if got.Plans[0].GeneratedTime != want {
		t.Fatalf("GeneratedTime = %d, want %d", got.Plans[0].GeneratedTime, want)
	}
}

func TestParquetFileNamesRemainUTC(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	localTime := time.Date(2026, time.September, 29, 9, 41, 20, 0, location)
	if got, want := formatTimeForFileName(localTime), "2026-09-29-01-41-20"; got != want {
		t.Fatalf("formatTimeForFileName() = %q, want %q", got, want)
	}

	want := localTime.UTC()
	for _, fileName := range []string{
		"2026-09-29-01-41-20-a1b2c3d4.parquet",
		"compacted-2026-09-29-01-41-20.parquet",
	} {
		t.Run(fileName, func(t *testing.T) {
			got, err := parseTimeFromFileName(fileName)
			if err != nil {
				t.Fatalf("parseTimeFromFileName() error = %v", err)
			}
			if !got.Equal(want) {
				t.Fatalf("parseTimeFromFileName() = %s (%d), want %s (%d)", got, got.Unix(), want, want.Unix())
			}
		})
	}
}
