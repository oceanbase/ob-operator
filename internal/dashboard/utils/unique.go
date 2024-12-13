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

package utils

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

// GenerateName generates unique name based on the base name
// Usually used for generating unique name for one-time resources
func GenerateName(base string) string {
	current := time.Now().Unix()
	if strings.HasSuffix(base, "-") {
		return fmt.Sprintf("%s%d-%s", base, current, rand.String(5))
	} else {
		return fmt.Sprintf("%s-%d-%s", base, current, rand.String(5))
	}
}
