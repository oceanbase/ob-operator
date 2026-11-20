/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
*/

package errors

import (
	"net/http"
	"testing"
)

func TestConflictStatus(t *testing.T) {
	err := New(ErrConflict, "resource already exists")
	if got := err.Status(); got != http.StatusConflict {
		t.Fatalf("conflict status = %d, want %d", got, http.StatusConflict)
	}
}
