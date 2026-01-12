package rules

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type MultiTableJoinRule struct {
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
	joinCount       int
}

func NewMultiTableJoinRule() *MultiTableJoinRule {
	return &MultiTableJoinRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

func (r *MultiTableJoinRule) Name() string {
	return "multi_table_join_rule"
}

func (r *MultiTableJoinRule) Description() string {
	return "The number of association tables is not recommended to exceed 5"
}

func (r *MultiTableJoinRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.joinCount = 0
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)

	if r.joinCount > 5 {
		r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
			RuleName:   r.Name(),
			Level:      "WARN",
			Suggestion: "Consider breaking the query into smaller, simpler queries or reviewing schema design.",
			Reason:     fmt.Sprintf("The query involves %d tables in JOIN operations, exceeding the recommended limit of 5.", r.joinCount),
		})
	}

	return r.diagnoseResults
}

func (r *MultiTableJoinRule) EnterJoined_table(ctx *obmysql.Joined_tableContext) {
	// joined_table : table_factor inner_join_type table_factor ...
	// Every time we visit a joined_table node, it implies a join operation.
	r.joinCount++
}
