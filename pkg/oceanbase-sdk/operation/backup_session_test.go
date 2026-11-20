package operation

import (
	"strings"
	"testing"

	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/const/sql"
)

func TestBackupSessionPasswordsAreRedacted(t *testing.T) {
	const sample = "sample-test-value"
	for _, statement := range []string{sql.SetBackupPassword, sql.SetRestorePassword} {
		_, sensitive := redactSQLForLog(statement)
		params, masked := redactParamsForLogWithPolicy([]any{sample}, sensitive)
		if !sensitive || !masked || strings.Contains(params, sample) {
			t.Fatal("backup or restore password was not masked")
		}
	}
}
