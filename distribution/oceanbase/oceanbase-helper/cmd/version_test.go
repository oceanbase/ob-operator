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

package cmd

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
