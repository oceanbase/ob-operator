/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package analyzer

import (
	"testing"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

func TestAnalyzeDoesNotReportFullScanWithoutPlanEvidence(t *testing.T) {
	diagnostics := NewManager().Analyze("SELECT SLEEP(0.04) AS qa_diag_live_20261008_sleep", nil)
	assertFullScanDiagnosis(t, diagnostics, false)
}

func TestAnalyzeReportsFullScanFromPlan(t *testing.T) {
	diagnostics := NewManager("HASH GROUP BY", "  table   full scan  ").Analyze(
		"SELECT SUM(amount) FROM qa_diag_live_20261008 WHERE note = 'a'",
		nil,
	)
	assertFullScanDiagnosis(t, diagnostics, true)
}

func TestAnalyzeDoesNotReportFullScanForOtherPlanOperators(t *testing.T) {
	diagnostics := NewManager("EXPRESSION", "TABLE RANGE SCAN").Analyze(
		"SELECT SUM(amount) FROM qa_diag_live_20261008",
		nil,
	)
	assertFullScanDiagnosis(t, diagnostics, false)
}

func assertFullScanDiagnosis(t *testing.T, diagnostics []model.SqlDiagnoseInfo, expected bool) {
	t.Helper()
	found := false
	for _, diagnostic := range diagnostics {
		if diagnostic.RuleName == "full_scan_rule" {
			found = true
		}
	}
	if found != expected {
		t.Fatalf("full scan diagnosis = %v, want %v; diagnostics: %+v", found, expected, diagnostics)
	}
}
