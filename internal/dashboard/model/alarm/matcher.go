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

package alarm

import (
	"strings"

	alarmconstant "github.com/oceanbase/ob-operator/internal/dashboard/business/alarm/constant"

	amlabels "github.com/prometheus/alertmanager/pkg/labels"
)

type Matcher struct {
	IsRegex bool   `json:"isRegex"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

func (m *Matcher) ToAmMatcher() (*amlabels.Matcher, error) {
	matchType := amlabels.MatchEqual
	if m.IsRegex {
		matchType = amlabels.MatchRegexp
	}
	return amlabels.NewMatcher(matchType, m.Name, m.Value)
}

func (m *Matcher) ExtractMatchedValues() []string {
	matchedValues := []string{m.Value}
	if m.IsRegex {
		matchedValues = strings.Split(m.Value, alarmconstant.RegexOR)
	}
	return matchedValues
}
