/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package rule

type RuleResponse struct {
	State          RuleState  `json:"state" binding:"required"`
	KeepFiringFor  int        `json:"keepFiringFor" binding:"required"`
	Health         RuleHealth `json:"health" binding:"required"`
	LastEvaluation int64      `json:"lastEvaluation" binding:"required"`
	EvaluationTime float64    `json:"evaluationTime" binding:"required"`
	LastError      string     `json:"lastError,omitempty"`
	Rule
}
