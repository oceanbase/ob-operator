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
	"regexp"
	"strings"

	apitypes "github.com/oceanbase/ob-operator/api/types"
	"github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	"github.com/oceanbase/ob-operator/internal/const/status/tenantstatus"
)

const (
	characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789~#%^&*_-+|(){}[]:,.?/"
	factor     = 4294901759
)

// CheckResourceName checks resource name in k8s
func CheckResourceName(name string) bool {
	regex := `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	re := regexp.MustCompile(regex)
	return re.MatchString(name)
}

// CheckPassword checks password when creating cluster
func CheckPassword(password string) bool {
	var (
		countUppercase   int
		countLowercase   int
		countNumber      int
		countSpecialChar int
	)

	for _, char := range password {
		if strings.ContainsRune(characters, char) {
			switch {
			case strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ", char):
				countUppercase++
			case strings.ContainsRune("abcdefghijklmnopqrstuvwxyz", char):
				countLowercase++
			case strings.ContainsRune("0123456789", char):
				countNumber++
			default:
				countSpecialChar++
			}
		} else {
			return false
		}
		// if satisfied
		if countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2 {
			return true
		}
	}
	return countUppercase >= 2 && countLowercase >= 2 && countNumber >= 2 && countSpecialChar >= 2
}

// CheckTenantName check Tenant name when creating tenant
func CheckTenantName(name string) bool {
	regex := `^[_a-zA-Z][^-]*$`
	re := regexp.MustCompile(regex)
	return re.MatchString(name)
}

// CheckTenantStatus checks running status of obtenant
func CheckTenantStatus(tenant *v1alpha1.OBTenant) error {
	if tenant.Status.Status != tenantstatus.Running {
		return fmt.Errorf("OBTenant status invalid, Status:%s", tenant.Status.Status)
	}
	return nil
}

// CheckClusterStatus checks running status of obcluster
func CheckClusterStatus(cluster *v1alpha1.OBCluster) error {
	if cluster.Status.Status != clusterstatus.Running {
		return fmt.Errorf("OBCluster status invalid, Status:%s", cluster.Status.Status)
	}
	return nil
}

// CheckPrimaryTenant checks primary tenant for a standbytenant
func CheckPrimaryTenant(standbytenant *v1alpha1.OBTenant) error {
	if standbytenant.Spec.Source == nil || standbytenant.Spec.Source.Tenant == nil {
		return fmt.Errorf("OBTenant %s has no primary tenant", standbytenant.Name)
	}
	return nil
}

// CheckTenantRole checks tenant role
func CheckTenantRole(tenant *v1alpha1.OBTenant, role apitypes.TenantRole) error {
	if tenant.Status.TenantRole != role {
		return fmt.Errorf("Tenant is not %s tenant", string(role))
	}
	return nil
}
