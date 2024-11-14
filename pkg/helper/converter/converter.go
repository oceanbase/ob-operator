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

package converter

import (
	"fmt"
	"math"
	"strconv"
)

func ConvertToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ConvertFloat(value any) any {
	switch v := value.(type) {
	case float64:
		return int64(math.Round(v))
	default:
		return v
	}
}

func AutoConvert(value string) any {
	// try parse to int
	if ret, err := strconv.Atoi(value); err == nil {
		return ret
	}
	// try parse to float64
	if ret, err := strconv.ParseFloat(value, 64); err == nil {
		return ret
	}
	// try parse to bool
	if ret, err := strconv.ParseBool(value); err == nil {
		return ret
	}
	// return string value
	return value
}
