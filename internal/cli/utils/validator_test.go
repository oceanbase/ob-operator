/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package utils_test

import (
	"testing"

	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

func TestCheckResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// test valid names
		{"validAlphanumeric", "example", true},
		{"validWithDash", "example-test", true},
		{"validWithDot", "example.com", true},
		{"validWithMultipleLabels", "sub.example.com", true},

		// test invalid names
		{"invalidWithUppercase", "Example", false},
		{"invalidWithUnderscore", "example_test", false},
		{"invalidWithSpace", "example test", false},
		{"invalidWithSpecialChar", "example@domain", false},
		{"invalidWithLeadingDash", "-example", false},
		{"invalidWithTrailingDash", "example-", false},
		{"invalidWithConsecutiveDashes", "example--com", true},
		{"invalidWithOnlyDot", ".", false},
		{"invalidEmptyString", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := utils.CheckResourceName(tt.input); result != tt.expected {
				t.Errorf("CheckResourceName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckTenantName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"validTenantName", "example", true},
		{"tenantNameWithUnderscore", "example_name", true},
		{"tenantNameWithHyphen", "example-name", false}, // it is not allowed to contain a '-' character

		{"emptyTenantName", "", false},
		{"tenantNameStartingWithHyphen", "-exmple", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := utils.CheckTenantName(tt.input); result != tt.expected {
				t.Errorf("CheckTenantName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"validPassword", "Aa1?Aa1?", true}, // valid password

		{"noEnoughlength", "Aa1?", false}, // at least two uppercase, lowercase, number and special char
		{"noUppercase", "aa11??", false},
		{"noLowercase", "AA11??", false},
		{"noNumber", "AAaa!!", false},
		{"noSpecialChar", "AAaa11", false},
		{"emptyPassword", "", false},
		{"onlySpecialChar", "????", false},
		{"onlyNumber", "1111", false},
		{"onlyLowercase", "aaaa", false},
		{"onlyUppercase", "AAAA", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := utils.CheckPassword(tt.input); result != tt.expected {
				t.Errorf("CheckPassword(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
