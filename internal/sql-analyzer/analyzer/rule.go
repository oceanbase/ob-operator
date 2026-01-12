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
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

type Rule interface {
	Name() string
	Description() string
	Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo
}
