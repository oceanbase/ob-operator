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

package helper

import (
	"testing"
)

func TestVersion(t *testing.T) {
	t.Run("test version", func(t *testing.T) {
		if "4.2.1.1-101010012023111012" < "4.2.0" {
			t.Error("version compare failed")
		}

		if "4.2.1.1-101010012023111012" < "4.2.0.0-100001282023042317" {
			t.Error("version compare failed")
		}
	})

	t.Run("test version struct", func(t *testing.T) {
		obv1, err := ParseOceanBaseVersion("4.2.1.1-101010012023111012")
		if err != nil {
			t.Error(err)
		}

		obv2, err := ParseOceanBaseVersion("4.2.0.0-100001282023042317")
		if err != nil {
			t.Error(err)
		}
		if obv1.Cmp(obv2) < 0 {
			t.Error("version compare failed")
		}

		obv3, err := ParseOceanBaseVersion("4.10.0.0-100001282023042317")
		if err != nil {
			t.Error(err)
		}
		if obv1.Cmp(obv3) > 0 {
			t.Error("version compare failed")
		}

		obv4, err := ParseOceanBaseVersion("3.1.2-101010012023111012")
		if err != nil {
			t.Error(err)
		}
		if obv1.Cmp(obv4) < 0 {
			t.Error("version compare failed")
		}

		_, err = ParseOceanBaseVersion("3.2-101010012023111012")
		if err == nil {
			t.Error("version parse failed")
		}
	})
}

func TestParseVersionOutput(t *testing.T) {
	const seekBanner = "/home/admin/oceanbase/bin/observer -V\nobserver (OceanSeek 4.6.3.1)\n\nREVISION: 282026111420-ee8c238b\nLIBPALF_VERSION: 1.3.4 (12026102715.al8-92eb5f77)"
	for _, tc := range []struct{ name, input, want string }{
		{"short", "4.2.1.7\n", "4.2.1.7"},
		{"short with build", " 4.2.1.3-1020231228\r\n", "4.2.1.3-1020231228"},
		{"OceanSeek", seekBanner, "4.6.3.1-282026111420"},
		{"startup messages", "runtime warning 3.1.0\n\n" + seekBanner, "4.6.3.1-282026111420"},
		{"legacy banner", "observer (OceanBase_CE 4.2.1.3)\r\n\tREVISION: 103000032023122818-8fe69c2056b07154bbd1ebd2c26e818ee0d5c56f\r\n", "4.2.1.3-103000032023122818"},
		{"missing revision", "observer (OceanBase 4.2.1.3)\nLIBPALF_VERSION: 1.3.4", ""},
		{"malformed revision", "observer (OceanBase 4.2.1.3)\nREVISION: 123abc-aaaa", ""},
		{"missing product version", "runtime version 1.3.4\nREVISION: 1020231228-aaaa", ""},
		{"blank", "\n\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseVersionOutput(tc.input)
			if got != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}
