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

package ac

import (
	"context"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	acmodel "github.com/oceanbase/ob-operator/internal/dashboard/model/ac"
	httpErr "github.com/oceanbase/ob-operator/pkg/errors"
	"github.com/oceanbase/ob-operator/pkg/k8s/client"
)

func GetRole(name string) (*acmodel.Role, error) {
	policies, err := enforcer.GetFilteredPolicy(0, name)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, httpErr.NewNotFound("role not found")
	}
	role := &acmodel.Role{
		Name: name,
	}
	for _, p := range policies {
		role.Description = p[3]
		if p[1] == "*" {
			role.Policies = append(role.Policies, acmodel.NewPolicy("*", "*", p[2]))
		} else {
			parts := strings.Split(p[1], "/")
			if len(parts) != 2 {
				return nil, httpErr.NewInternal("corrupted policy" + strings.Join(p, " "))
			}
			role.Policies = append(role.Policies, acmodel.NewPolicy(parts[0], parts[1], p[2]))
		}
	}
	return role, nil
}

// NOTES: The returned type must be []*acmodel.Role, not []acmodel.Role
func ListRoles(_ context.Context) ([]*acmodel.Role, error) {
	polices, err := enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}
	// "policies"=[["admin" "*" "*" "Super admin"] ["admin2" "book/*" "read" "Book reader"]]
	roles := make([]*acmodel.Role, 0, len(polices))
	roleMapping := make(map[string]*acmodel.Role, 0)
	for _, p := range polices {
		roleName := p[0]
		role, ok := roleMapping[roleName]
		if !ok {
			role = &acmodel.Role{
				Name:        roleName,
				Description: p[3],
			}
			roleMapping[roleName] = role
			roles = append(roles, role)
		}
		if p[1] == "*" {
			role.Policies = append(role.Policies, acmodel.NewPolicy("*", "*", p[2]))
		} else {
			parts := strings.Split(p[1], "/")
			if len(parts) != 2 {
				return nil, httpErr.NewInternal("corrupted policy" + strings.Join(p, " "))
			}
			role.Policies = append(role.Policies, acmodel.NewPolicy(parts[0], parts[1], p[2]))
		}
	}
	return roles, nil
}

// Patch role will delete the role first and then create a new one
func PatchRole(ctx context.Context, roleName string, param *acmodel.PatchRoleParam) (*acmodel.Role, error) {
	_, err := DeleteRole(ctx, roleName, "no-persist")
	if err != nil {
		return nil, err
	}
	createParam := &acmodel.CreateRoleParam{
		Name:        roleName,
		Description: param.Description,
		Permissions: param.Permissions,
	}
	return CreateRole(ctx, createParam)
}

// If the extra parameter is provided, the role will NOT be persisted to the file
func DeleteRole(ctx context.Context, roleName string, extra ...string) (*acmodel.Role, error) {
	policies, err := enforcer.GetFilteredPolicy(0, roleName)
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, httpErr.NewNotFound("role not found")
	}
	role := &acmodel.Role{
		Name: roleName,
	}
	for _, p := range policies {
		role.Description = p[3]
		role.Policies = append(role.Policies, acmodel.NewPolicy(p[1], p[1], p[2]))
		ok, err := enforcer.RemovePolicy(p)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to remove policy")
		}
	}
	if len(extra) == 0 {
		err = persistPolicies(ctx, enforcer.policyPath, enforcer.configMapPath)
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

// CreateRole creates a role with the given name and permissions
// If the extra parameter is provided, the role will NOT be persisted to the file
func CreateRole(ctx context.Context, param *acmodel.CreateRoleParam, extra ...string) (*acmodel.Role, error) {
	var err error
	role := &acmodel.Role{
		Name:        param.Name,
		Description: param.Description,
	}
	for _, p := range param.Permissions {
		ok, err := enforcer.AddPolicy(param.Name, p.ComposeDomainObject(), string(p.Action), param.Description)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, httpErr.NewInternal("failed to add policy")
		}
		role.Policies = append(role.Policies, p)
	}
	if len(extra) == 0 {
		err = persistPolicies(ctx, enforcer.policyPath, enforcer.configMapPath)
		if err != nil {
			return nil, err
		}
	}
	return role, nil
}

// Transform the policies to a CSV string
func policiesToCsv() (string, error) {
	// "policies"=[["admin" "*" "*" "Super admin"] ["admin2" "book/*" "read" "Book reader"]]
	// admin, *, *, "Super admin"
	// admin2, book/*, read, "Book reader"
	csv := ""
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return csv, err
	}
	for _, p := range policies {
		csv += strings.Join([]string{"p", p[0], p[1], p[2]}, ", ") + ", \"" + p[3] + "\"\n"
	}
	groupings, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return csv, err
	}
	for _, g := range groupings {
		csv += strings.Join([]string{"g", g[0], g[1]}, ", ") + "\n"
	}
	return csv, nil
}

// Persist the policies to the file or config map
// Extra[0] is the config map name
func persistPolicies(ctx context.Context, targetFile string, extra ...string) error {
	csv, err := policiesToCsv()
	if err != nil {
		return err
	}
	if len(extra) > 0 {
		clt := client.GetClient()
		var cm *corev1.ConfigMap
		cm, err = clt.ClientSet.CoreV1().ConfigMaps("default").Get(ctx, extra[0], metav1.GetOptions{})
		if err != nil {
			return err
		}
		cm.Data[targetFile] = csv
		_, err = clt.ClientSet.CoreV1().ConfigMaps("default").Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(csv)
	if err != nil {
		return err
	}
	return nil
}
