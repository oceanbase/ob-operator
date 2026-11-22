/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package alarm

import (
	"github.com/oceanbase/ob-operator/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

func logServiceSilenceIdentity(value string) (string, string, error) {
	ns, name, ok := strings.Cut(value, "/")
	if !ok || len(validation.IsDNS1123Label(ns)) > 0 || len(validation.IsDNS1123Subdomain(name)) > 0 {
		return "", "", errors.NewBadRequest("LogService identity must be namespace/name")
	}
	return ns, name, nil
}
