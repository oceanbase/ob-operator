package rules

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type ArithmeticRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
}

func NewArithmeticRule() *ArithmeticRule {
	return &ArithmeticRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *ArithmeticRule) Name() string {
	return "arithmetic_rule"
}

func (r *ArithmeticRule) Description() string {
	return "Field operations are not recommended. Example: a + 1 > 2 => a > 2 - 1"
}

func (r *ArithmeticRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *ArithmeticRule) EnterBit_expr(ctx *obmysql.Bit_exprContext) {
	// Check for arithmetic operators
	if ctx.Plus() != nil || ctx.Minus() != nil || ctx.Star() != nil || ctx.Div() != nil || ctx.Mod() != nil || ctx.MOD() != nil || ctx.DIV() != nil {
		// Bit_expr has children.
		// Grammar: bit_expr operator bit_expr
		// If it has an operator, it likely has 3 children (left, op, right) or more if chained?
		// ANTLR usually creates a list of Bit_expr children.

		exprs := ctx.AllBit_expr()
		if len(exprs) >= 2 {
			left := exprs[0]
			right := exprs[1]

			if r.isColumn(left) || r.isColumn(right) {
				r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
					RuleName:   r.Name(),
					Level:      "NOTICE",
					Suggestion: "Consider simplifying your expressions by moving constants out of comparisons.",
					Reason:     r.Description(),
				})
			}
		}
	}
}

func (r *ArithmeticRule) isColumn(ctx obmysql.IBit_exprContext) bool {
	// Cast interface to concrete struct to access methods
	c, ok := ctx.(*obmysql.Bit_exprContext)
	if !ok {
		return false
	}

	// bit_expr -> simple_expr
	if c.Simple_expr() != nil {
		s, ok := c.Simple_expr().(*obmysql.Simple_exprContext)
		if !ok {
			return false
		}
		// simple_expr -> column_ref
		if s.Column_ref() != nil {
			return true
		}
	}
	return false
}
