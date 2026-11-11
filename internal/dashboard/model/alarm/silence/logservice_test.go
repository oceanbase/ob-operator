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
