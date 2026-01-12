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
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type IndexColumnFuzzyMatchRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
	indexes         []model.IndexInfo
}

func NewIndexColumnFuzzyMatchRule() *IndexColumnFuzzyMatchRule {
	return &IndexColumnFuzzyMatchRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *IndexColumnFuzzyMatchRule) Name() string {
	return "index_column_fuzzy_match_rule"
}

func (r *IndexColumnFuzzyMatchRule) Description() string {
	return "Avoid using fuzzy or left fuzzy matches on indexed columns in query conditions as it may lead to performance degradation."
}

func (r *IndexColumnFuzzyMatchRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.indexes = indexes
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *IndexColumnFuzzyMatchRule) EnterPredicate(ctx *obmysql.PredicateContext) {
	// predicate : bit_expr (NOT? LIKE simple_expr ...)
	if ctx.LIKE() != nil {
		// Use index 0 for Bit_expr and Simple_expr
		bitExpr := ctx.Bit_expr(0)
		simpleExpr := ctx.Simple_expr(0)

		if bitExpr != nil && simpleExpr != nil {
			// Check if bitExpr is a column reference
			colName, tableName := r.extractColumnInfo(bitExpr)
			if colName != "" {
				// Check if column is indexed
				if r.isColumnIndexed(tableName, colName) {
					// Check for fuzzy match pattern
					if r.isFuzzyMatch(simpleExpr) {
						r.addResult(colName)
					}
				}
			}
		}
	}
}

func (r *IndexColumnFuzzyMatchRule) extractColumnInfo(bitExpr obmysql.IBit_exprContext) (string, string) {
	// bit_expr -> simple_expr -> column_ref
	if be, ok := bitExpr.(*obmysql.Bit_exprContext); ok {
		if se := be.Simple_expr(); se != nil {
			if simpleExpr, ok := se.(*obmysql.Simple_exprContext); ok {
				if colRef := simpleExpr.Column_ref(); colRef != nil {
					if cr, ok := colRef.(*obmysql.Column_refContext); ok {
						colName := cr.Column_name().GetText()
						// Strip backticks if present
						colName = strings.Trim(colName, "`")

						tableName := ""
						if len(cr.AllRelation_name()) > 0 {
							tableName = cr.Relation_name(0).GetText()
							tableName = strings.Trim(tableName, "`")
						}
						return colName, tableName
					}
				}
			}
		}
	}
	return "", ""
}

func (r *IndexColumnFuzzyMatchRule) isColumnIndexed(tableName, colName string) bool {
	for _, idx := range r.indexes {
		if tableName != "" && !strings.EqualFold(idx.TableName, tableName) {
			continue
		}

		for _, idxCol := range idx.Columns {
			if strings.EqualFold(idxCol, colName) {
				return true
			}
		}
	}
	return false
}

func (r *IndexColumnFuzzyMatchRule) isFuzzyMatch(ctx obmysql.ISimple_exprContext) bool {
	if se, ok := ctx.(*obmysql.Simple_exprContext); ok {
		if ec := se.Expr_const(); ec != nil {
			if l := ec.Literal(); l != nil {
				val := ""
				if cs := l.Complex_string_literal(); cs != nil {
					val = cs.GetText()
				}
				// Removed STRING_VALUE check

				if val != "" {
					val = strings.Trim(val, "'\"")
					if strings.HasPrefix(val, "%") {
						return true
					}
					if strings.Contains(val, "%") && !strings.HasSuffix(val, "%") {
						return true
					}
				}
			}
		}
	}
	return false
}

func (r *IndexColumnFuzzyMatchRule) addResult(colName string) {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Avoid using fuzzy or left fuzzy matches on indexed columns: " + colName,
		Reason:     r.Description(),
	})
}
