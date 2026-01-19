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

package analyzer

import (
	"sync"
	"time"

	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/analyzer/rules"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"

	// Import the generated parser package.
	// Note: This package path must match where 'make generate-parser' outputs the code.
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
	logger "github.com/sirupsen/logrus"
)

type Manager struct {
	rules []Rule
}

func NewManager() *Manager {
	m := &Manager{
		rules: []Rule{},
	}
	m.RegisterRules()
	return m
}

func (m *Manager) RegisterRules() {
	// Register all available rules here
	m.rules = append(m.rules, rules.NewSelectAllRule())
	m.rules = append(m.rules, rules.NewArithmeticRule())
	m.rules = append(m.rules, rules.NewIsNullRule())
	m.rules = append(m.rules, rules.NewLargeInClauseRule())
	m.rules = append(m.rules, rules.NewMultiTableJoinRule())
	m.rules = append(m.rules, rules.NewUpdateDeleteWithoutWhereRule())
	m.rules = append(m.rules, rules.NewUpdateDeleteMultiTableRule())
	m.rules = append(m.rules, rules.NewFullScanRule())
	m.rules = append(m.rules, rules.NewIndexColumnFuzzyMatchRule())
	m.rules = append(m.rules, rules.NewFunctionOnIndexedColumnRule())
}

func (m *Manager) Analyze(sql string, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	var diagnostics []model.SqlDiagnoseInfo

	// Setup ANTLR input stream
	inputStream := antlr.NewInputStream(sql)

	// Create Lexer
	lexer := obmysql.NewOBLexer(inputStream)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create Parser
	p := obmysql.NewOBParser(stream)
	// Add error listener to avoid printing to stdout
	p.RemoveErrorListeners()

	// Use SLL prediction mode for better performance, fallback to LL if it fails
	p.GetInterpreter().SetPredictionMode(antlr.PredictionModeSLL)
	p.SetErrorHandler(antlr.NewBailErrorStrategy())

	// Parse the SQL (assuming 'Sql_stmt' is the entry point rule)
	parseStart := time.Now()
	var tree antlr.ParseTree
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback to LL prediction mode
				stream.Seek(0)
				p.SetTokenStream(stream)
				p.GetInterpreter().SetPredictionMode(antlr.PredictionModeLL)
				p.SetErrorHandler(antlr.NewDefaultErrorStrategy())
				tree = p.Sql_stmt()
			}
		}()
		tree = p.Sql_stmt()
	}()
	logger.Debugf("[Analyzer] Parse took %v", time.Since(parseStart))

	// Run all registered rules
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, rule := range m.rules {
		wg.Add(1)
		go func(r Rule) {
			defer wg.Done()
			results := r.Analyze(tree, indexes)
			logger.Debugf("[Manager] Rule %s returned %d results", r.Name(), len(results))
			if len(results) > 0 {
				mu.Lock()
				// Only keep the first result for each rule, currently each rule returns the same record, can be enriched to include the detailed sql statement part.
				diagnostics = append(diagnostics, results[0])
				mu.Unlock()
			}
		}(rule)
	}
	wg.Wait()

	logger.Debugf("[Manager] Total diagnostics found: %d", len(diagnostics))

	return diagnostics
}
