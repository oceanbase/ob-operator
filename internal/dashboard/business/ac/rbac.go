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
	envRBACPolicyPath = "RBAC_POLICY_PATH"
)

var enforcer *casbin.Enforcer

func init() {
	var err error
	enforcer, err = initEnforcer()
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
	return enforcer
}

func initEnforcer() (*casbin.Enforcer, error) {
	var err error
	var rbacPolicyPath = os.Getenv(envRBACPolicyPath)
	model, err := model.NewModelFromString(modelDefinition)
	if err != nil {
		return nil, err
	}

	if rbacPolicyPath == "" {
		rbacPolicyPath = "./rbac_policy.csv"
	}

	adapter := fa.NewAdapter(rbacPolicyPath)

	enforcer, err = casbin.NewEnforcer(model, adapter)
	if err != nil {
		return nil, err
	}
	enforcer.AddNamedMatchingFunc("g", "KeyMatch2", util.KeyMatch2)

	return enforcer, nil
}
