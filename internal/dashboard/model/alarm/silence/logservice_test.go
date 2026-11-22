/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package silence

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/alarm"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/oceanbase"
	"testing"
)

func TestLogServiceSilenceRoundtrip(t *testing.T) {
	result := extractInstances(map[string]alarm.Matcher{"namespace": {Name: "namespace", Value: "test"}, "logservice": {Name: "logservice", Value: "ls\\.one|ls-two", IsRegex: true}})
	if len(result) != 2 || result[0].Type != oceanbase.TypeLogService || result[0].LogService != "test/ls.one" || result[1].LogService != "test/ls-two" {
		t.Fatalf("unexpected instances: %+v", result)
	}
}
