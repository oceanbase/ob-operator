package rules

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
	logger "github.com/sirupsen/logrus"
)

// SelectAllRule checks for the usage of SELECT *
type SelectAllRule struct {
	// Embed the BaseOBParserListener to get default no-op implementations
	*obmysql.BaseOBParserListener
	diagnoseResults []model.SqlDiagnoseInfo
}

// NewSelectAllRule creates a new SelectAllRule
func NewSelectAllRule() *SelectAllRule {
	return &SelectAllRule{
		BaseOBParserListener: &obmysql.BaseOBParserListener{},
	}
}

// Name returns the name of the rule
func (r *SelectAllRule) Name() string {
	return "select_all_rule"
}

// Description returns the description of the rule
func (r *SelectAllRule) Description() string {
	return "Avoid using SELECT * in queries."
}

// Analyze traverses the parse tree to find SELECT * statements
func (r *SelectAllRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{} // Reset results for each analysis

	logger.Infof("[SelectAllRule] Starting analysis")
	// Use ParseTreeWalker to walk the tree with the listener
	walker := antlr.NewParseTreeWalker()
	walker.Walk(r, tree)

	return r.diagnoseResults
}

// EnterProjection is called when the listener enters a 'projection' rule.
// Matches: projection : bit_expr | bit_expr AS? column_label | bit_expr AS? STRING_VALUE | Star ;
func (r *SelectAllRule) EnterProjection(ctx *obmysql.ProjectionContext) {
	logger.Infof("[SelectAllRule] Entering Projection. Text: %s", ctx.GetText())
	if ctx.Star() != nil { // Check if the '*' token exists in this projection context
		logger.Infof("[SelectAllRule] Found Star in Projection")
		r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
			RuleName:   r.Name(),
			Level:      "WARN",
			Suggestion: "Specify specific columns instead of using SELECT * to improve performance and clarity.",
			Reason:     "Using SELECT * retrieves all columns, which can be inefficient and return unnecessary data.",
		})
	} else {
		logger.Infof("[SelectAllRule] No Star in Projection")
	}
}
