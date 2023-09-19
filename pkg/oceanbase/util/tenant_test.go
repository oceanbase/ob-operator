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

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertFromLocalityStr(t *testing.T) {
	locality := "FULL{1}@zone1, FULL{1}@zone2, FULL{1}@zone3"
	replicas := ConvertFromLocalityStr(locality)
	require.Equal(t, 3, len(replicas))
}
