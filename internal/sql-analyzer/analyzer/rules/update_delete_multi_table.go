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

type UpdateDeleteMultiTableRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
}

func NewUpdateDeleteMultiTableRule() *UpdateDeleteMultiTableRule {
	return &UpdateDeleteMultiTableRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *UpdateDeleteMultiTableRule) Name() string {
	return "update_delete_multi_table_rule"
}

func (r *UpdateDeleteMultiTableRule) Description() string {
	return "UPDATE / DELETE does not recommend using multiple tables"
}

func (r *UpdateDeleteMultiTableRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)
	return r.diagnoseResults
}

func (r *UpdateDeleteMultiTableRule) EnterDelete_stmt(ctx *obmysql.Delete_stmtContext) {
	// Check if multi_delete_table is used
	if ctx.Multi_delete_table() != nil {
		r.addResult()
	}
}

func (r *UpdateDeleteMultiTableRule) EnterUpdate_stmt(ctx *obmysql.Update_stmtContext) {
	// update_stmt: UPDATE ... table_references ...
	// table_references -> table_reference (Comma table_reference)*
	// table_reference -> joined_table

	refs := ctx.Table_references()
	if refs != nil {
		tr, ok := refs.(*obmysql.Table_referencesContext)
		if ok {
			// Check for multiple table_reference (comma separated)
			if len(tr.AllTable_reference()) > 1 {
				r.addResult()
			} else if len(tr.AllTable_reference()) == 1 {
				// Check for JOINs in the single reference
				ref := tr.Table_reference(0)
				// Cast to interface to check type
				if refCtx, ok := ref.(*obmysql.Table_referenceContext); ok {
					if refCtx.Joined_table() != nil {
						r.addResult()
					}
				}
			}
		}
	}
}

func (r *UpdateDeleteMultiTableRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Break down the operation into separate single-table statements or use transactions.",
		Reason:     r.Description(),
	})
}
