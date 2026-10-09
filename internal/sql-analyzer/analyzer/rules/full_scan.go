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

package rules

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

type FullScanRule struct {
	diagnoseResults []model.SqlDiagnoseInfo
	planOperators   []string
}

func NewFullScanRule(planOperators []string) *FullScanRule {
	return &FullScanRule{
		planOperators: planOperators,
	}
}

func (r *FullScanRule) Name() string {
	return "full_scan_rule"
}

func (r *FullScanRule) Description() string {
	return "The execution plan contains a TABLE FULL SCAN operator. Online full table scans are not recommended except for very small tables, very low-frequency queries, or small result sets."
}

func (r *FullScanRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}

	for _, operator := range r.planOperators {
		if normalizePlanOperator(operator) == "TABLE FULL SCAN" {
			r.addResult()
			break
		}
	}

	return r.diagnoseResults
}

func (r *FullScanRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "The execution plan performs a full table scan which may impact performance. Consider adding indexes, refining WHERE clauses, or restructuring the query to utilize existing indexes.",
		Reason:     r.Description(),
	})
}

func normalizePlanOperator(operator string) string {
	return strings.Join(strings.Fields(strings.ToUpper(operator)), " ")
}
