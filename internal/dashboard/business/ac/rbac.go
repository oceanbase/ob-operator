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
	"os"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fa "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/casbin/casbin/v2/util"
)

const modelDefinition = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, desc

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && g(r.obj, p.obj) && g(r.act, p.act)
`

const (
	envRBACPolicyPath      = "RBAC_POLICY_PATH"
	envRBACPolicyConfigMap = "RBAC_POLICY_CONFIG_MAP"
)

type enf struct {
	*casbin.Enforcer
	policyMu      sync.RWMutex
	accMu         sync.RWMutex
	policyPath    string
	configMapPath string
}

var enforcer *enf

func init() {
	var err error
	var rbacPolicyPath = os.Getenv(envRBACPolicyPath)
	var rbacPolicyCm = os.Getenv(envRBACPolicyConfigMap)
	if rbacPolicyCm == "" {
		panic("RBAC_POLICY_CONFIG_MAP is required")
	}
	if rbacPolicyPath == "" {
		rbacPolicyPath = "/etc/rbac/rbac_policy.csv"
	}

	enforcer, err = initEnforcer(rbacPolicyPath, rbacPolicyCm)
	if err != nil {
		panic(err)
	}
}

// Use sub, obj, act to emulate RBAC model with domains
// sub: user
// obj: resource in the format of domain/object
// act: operation on the resource
// e.g., alice, domain1/data1, read
// e.g., bob, domain2/*, write
func Enforcer() *casbin.Enforcer {
	return enforcer.Enforcer
}

func initEnforcer(rbacPolicyPath, rbacPolicyCm string) (*enf, error) {
	var err error
	model, err := model.NewModelFromString(modelDefinition)
	if err != nil {
		return nil, err
	}
	adapter := fa.NewAdapter(rbacPolicyPath)

	e, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, err
	}
	internal := &enf{
		Enforcer:      e,
		policyPath:    rbacPolicyPath,
		configMapPath: rbacPolicyCm,
	}
	internal.AddNamedMatchingFunc("g", "KeyMatch2", util.KeyMatch2)

	return internal, nil
}
