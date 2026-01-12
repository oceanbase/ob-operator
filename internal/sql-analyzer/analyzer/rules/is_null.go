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

type IsNullRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
}

func NewIsNullRule() *IsNullRule {
	return &IsNullRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *IsNullRule) Name() string {
	return "is_null_rule"
}

func (r *IsNullRule) Description() string {
	return "Use IS NULL to determine whether it is a NULL value. Direct comparison of NULL to any value is NULL."
}

func (r *IsNullRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *IsNullRule) EnterBool_pri(ctx *obmysql.Bool_priContext) {
	// The rule is to warn against using `=`, `!=`, `<=>` with NULL.
	// Valid forms: `col IS NULL`, `col IS NOT NULL`.
	// Invalid forms: `col = NULL`, `col != NULL`, `NULL = col`, `NULL != col`.

	// Check for `bit_expr IS NULLX` or `bit_expr IS NOT NULLX`.
	// If it contains `IS` and `NULLX`, it's a correct usage, so we skip it for this rule.
	if ctx.IS() != nil && ctx.NULLX() != nil {
		// This is a correct usage (`IS NULL`), so we skip it for this rule.
		return
	}
	if ctx.IS() != nil && ctx.Not() != nil && ctx.NULLX() != nil {
		// This is also a correct usage (`IS NOT NULL`), skip.
		return
	}

	// Now check for comparisons that use operators like `=`, `!=`, `<=>`.
	isComparisonOp := ctx.COMP_EQ() != nil || ctx.COMP_NE() != nil || ctx.COMP_NSEQ() != nil

	if isComparisonOp {
		// A Bool_priContext representing a comparison usually has a left-hand side (recursive Bool_pri)
		// and a right-hand side (Predicate).
		// We need to check if either operand directly resolves to a NULL literal.

		leftOperandIsLiteralNull := false
		if ctx.Bool_pri() != nil {
			leftOperandIsLiteralNull = r.containsNullLiteral(ctx.Bool_pri())
		}

		rightOperandIsLiteralNull := false
		if ctx.Predicate() != nil {
			rightOperandIsLiteralNull = r.containsNullLiteral(ctx.Predicate())
		}

		if leftOperandIsLiteralNull || rightOperandIsLiteralNull {
			r.addResult()
		}
	}
}

func (r *IsNullRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Detected comparison with NULL using =, !=, or <=>. Use 'IS NULL' or 'IS NOT NULL' for correct NULL checks.",
		Reason:     r.Description(),
	})
}

// containsNullLiteral recursively checks if any descendant of the given ParseTree is a NULLX literal.
func (r *IsNullRule) containsNullLiteral(node antlr.ParseTree) bool {
	if node == nil {
		return false
	}

	if literalCtx, ok := node.(obmysql.ILiteralContext); ok && literalCtx.NULLX() != nil {
		return true
	}

	// Check if the node itself is a NULLX terminal node
	if terminalNode, ok := node.(antlr.TerminalNode); ok && terminalNode.GetSymbol().GetTokenType() == obmysql.OBParserNULLX {
		return true
	}

	// Recursively check children
	for _, child := range node.GetChildren() {
		if parseTreeChild, ok := child.(antlr.ParseTree); ok {
			if r.containsNullLiteral(parseTreeChild) {
				return true
			}
		}
	}
	return false
}
