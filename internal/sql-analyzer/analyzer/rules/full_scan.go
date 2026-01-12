package rules

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type FullScanRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
	hasSargablePred bool
	isDML           bool
}

func NewFullScanRule() *FullScanRule {
	return &FullScanRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *FullScanRule) Name() string {
	return "full_scan_rule"
}

func (r *FullScanRule) Description() string {
	return "Online query full table scan is not recommended. Exceptions are: very small table, very low frequency, or small result set."
}

func (r *FullScanRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.hasSargablePred = false
	r.isDML = false

	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)

	if r.isDML && !r.hasSargablePred {
		r.addResult()
	}

	return r.diagnoseResults
}

func (r *FullScanRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Detected a potential full table scan which may impact performance. Consider adding indexes, refining WHERE clauses, or restructuring the query to utilize existing indexes.",
		Reason:     r.Description(),
	})
}

func (r *FullScanRule) EnterSelect_stmt(ctx *obmysql.Select_stmtContext) {
	r.isDML = true
}

func (r *FullScanRule) EnterUpdate_stmt(ctx *obmysql.Update_stmtContext) {
	r.isDML = true
}

func (r *FullScanRule) EnterDelete_stmt(ctx *obmysql.Delete_stmtContext) {
	r.isDML = true
}

func (r *FullScanRule) EnterBool_pri(ctx *obmysql.Bool_priContext) {
	if ctx.COMP_EQ() != nil || ctx.COMP_GE() != nil || ctx.COMP_GT() != nil || ctx.COMP_LE() != nil || ctx.COMP_LT() != nil {
		r.hasSargablePred = true
	}
}

func (r *FullScanRule) EnterPredicate(ctx *obmysql.PredicateContext) {
	if ctx.IN() != nil {
		if ctx.Not() == nil {
			r.hasSargablePred = true
		}
	} else if ctx.BETWEEN() != nil {
		if ctx.Not() == nil {
			r.hasSargablePred = true
		}
	} else if ctx.LIKE() != nil {
		if ctx.Not() == nil {
			// Use index 0 for Simple_expr as it's a list in generated code
			patternCtx := ctx.Simple_expr(0)
			if patternCtx != nil {
				if r.isLeftFuzzy(patternCtx) {
					// Left fuzzy ('%abc') is NOT SARGable.
				} else {
					r.hasSargablePred = true
				}
			}
		}
	}
}

func (r *FullScanRule) isLeftFuzzy(ctx obmysql.ISimple_exprContext) bool {
	if se, ok := ctx.(*obmysql.Simple_exprContext); ok {
		if ec := se.Expr_const(); ec != nil {
			if l := ec.Literal(); l != nil {
				if cs := l.Complex_string_literal(); cs != nil {
					val := cs.GetText()
					val = strings.Trim(val, "'\"")
					if strings.HasPrefix(val, "%") {
						return true
					}
				}
			}
		}
	}
	return false
}

func (r *FullScanRule) EnterSimple_expr(ctx *obmysql.Simple_exprContext) {
	if ctx.EXISTS() != nil {
		r.hasSargablePred = true
	}
}
