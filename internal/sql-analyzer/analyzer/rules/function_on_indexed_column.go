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

type FunctionOnIndexedColumnRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
	indexes         []model.IndexInfo
	predicateDepth  int
	functionDepth   int
}

func NewFunctionOnIndexedColumnRule() *FunctionOnIndexedColumnRule {
	return &FunctionOnIndexedColumnRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *FunctionOnIndexedColumnRule) Name() string {
	return "function_on_indexed_column_rule"
}

func (r *FunctionOnIndexedColumnRule) Description() string {
	return "Avoid wrapping indexed columns in functions within WHERE/ON clauses, as this prevents index usage (e.g., use 'col >= ...' instead of 'YEAR(col) = ...')."
}

func (r *FunctionOnIndexedColumnRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.indexes = indexes
	r.predicateDepth = 0
	r.functionDepth = 0
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *FunctionOnIndexedColumnRule) EnterPredicate(ctx *obmysql.PredicateContext) {
	r.predicateDepth++
}

func (r *FunctionOnIndexedColumnRule) ExitPredicate(ctx *obmysql.PredicateContext) {
	r.predicateDepth--
}

func (r *FunctionOnIndexedColumnRule) EnterFunc_expr(ctx *obmysql.Func_exprContext) {
	r.functionDepth++
}

func (r *FunctionOnIndexedColumnRule) ExitFunc_expr(ctx *obmysql.Func_exprContext) {
	r.functionDepth--
}

func (r *FunctionOnIndexedColumnRule) EnterColumn_ref(ctx *obmysql.Column_refContext) {
	if r.predicateDepth > 0 && r.functionDepth > 0 {
		colName := ctx.Column_name().GetText()
		colName = strings.Trim(colName, "`")

		tableName := ""
		if len(ctx.AllRelation_name()) > 0 {
			tableName = ctx.Relation_name(0).GetText()
			tableName = strings.Trim(tableName, "`")
		}

		if r.isColumnIndexed(tableName, colName) {
			r.addResult(colName)
		}
	}
}

func (r *FunctionOnIndexedColumnRule) isColumnIndexed(tableName, colName string) bool {
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

func (r *FunctionOnIndexedColumnRule) addResult(colName string) {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Indexed column '" + colName + "' is wrapped in a function in a predicate. This prevents index usage. Consider rewriting the query.",
		Reason:     r.Description(),
	})
}
