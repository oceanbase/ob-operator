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

package constant

import "testing"

func TestServiceAccountNameMatchesDeploymentEnvironment(t *testing.T) {
	t.Setenv("SERVICE_ACCOUNT", "dashboard-qa-sa")
	if got := ServiceAccountName(); got != "dashboard-qa-sa" {
		t.Fatalf("ServiceAccountName() = %q, want deployment value %q", got, "dashboard-qa-sa")
	}
}

func TestServiceAccountNameFallback(t *testing.T) {
	if got := serviceAccountName("  "); got != defaultServiceAccountName {
		t.Fatalf("serviceAccountName() = %q, want fallback %q", got, defaultServiceAccountName)
	}
	if got := serviceAccountName("  dashboard-qa-sa  "); got != "dashboard-qa-sa" {
		t.Fatalf("serviceAccountName() = %q, want trimmed deployment value", got)
	}
}
