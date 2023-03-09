/*
Copyright (c) 2021 OceanBase
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

import "fmt"

func TrimSuffixLastest(s string) string {
	s = s[:len(s)-1]
	return s
}

func FormatSize(size int) string {
	units := [...]string{"Bi", "Ki", "Mi", "Gi", "Ti", "Pi"}
	idx := 0
	size1 := float64(size)
	for idx < 5 && size1 >= 1024 {
		size1 /= 1024.0
		idx += 1
	}
	res := fmt.Sprintf("%.1f%s", size1, units[idx])
	return res
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
