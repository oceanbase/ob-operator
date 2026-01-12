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
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type UpdateDeleteWithoutWhereRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
}

func NewUpdateDeleteWithoutWhereRule() *UpdateDeleteWithoutWhereRule {
	return &UpdateDeleteWithoutWhereRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *UpdateDeleteWithoutWhereRule) Name() string {
	return "update_delete_without_where_or_true_condition_rule"
}

func (r *UpdateDeleteWithoutWhereRule) Description() string {
	return "UPDATE or DELETE statements should not be executed without a WHERE clause or with an always-true WHERE condition."
}

func (r *UpdateDeleteWithoutWhereRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *UpdateDeleteWithoutWhereRule) EnterDelete_stmt(ctx *obmysql.Delete_stmtContext) {
	// delete_stmt: DELETE ... (WHERE expr)? ...
	// If WHERE is missing, expr will be nil.

	if ctx.WHERE() == nil {
		r.addResult()
	} else if ctx.Expr() != nil {
		if r.isAlwaysTrue(ctx.Expr()) {
			r.addResult()
		}
	}
}

func (r *UpdateDeleteWithoutWhereRule) EnterUpdate_stmt(ctx *obmysql.Update_stmtContext) {
	// update_stmt: UPDATE ... (WHERE expr)? ...

	if ctx.WHERE() == nil {
		r.addResult()
	} else if ctx.Expr() != nil {
		if r.isAlwaysTrue(ctx.Expr()) {
			r.addResult()
		}
	}
}

func (r *UpdateDeleteWithoutWhereRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "CRITICAL",
		Suggestion: "Ensure a proper and specific WHERE condition is used.",
		Reason:     "Executing UPDATE or DELETE statements without a WHERE clause or with an always-true WHERE condition can affect all rows.",
	})
}

// Simple heuristic for always true: 1=1, a=a
func (r *UpdateDeleteWithoutWhereRule) isAlwaysTrue(ctx obmysql.IExprContext) bool {
	// expr -> bool_pri -> bit_expr -> simple_expr -> expr_const -> literal -> INTNUM
	// or bool_pri COMP_EQ bool_pri

	// This is complex to implement fully without an evaluator.
	// We implement a basic check for literal equality (e.g., 1=1).

	e, ok := ctx.(*obmysql.ExprContext)
	if !ok {
		return false
	}

	if e.Bool_pri() != nil {
		b, ok := e.Bool_pri().(*obmysql.Bool_priContext)
		if ok {
			// Check for equality comparison: bool_pri = bool_pri
			if b.COMP_EQ() != nil {
				// We need left and right operands.
				// bool_pri rule is recursive.
				// If we have "1 = 1", it might be:
				// bool_pri (1) COMP_EQ predicate(bit_expr(1))
				// or bool_pri(1) COMP_EQ bool_pri(1) ?

				// Simplified check: Get text of left and right
				// This is "cheating" but effective for "1=1"
				if b.Bool_pri() != nil && b.Predicate() != nil {
					leftText := b.Bool_pri().GetText()
					rightText := b.Predicate().GetText()
					if leftText == rightText {
						return true
					}
				}
			}
		}
	}

	return false
}
