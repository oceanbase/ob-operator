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
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type EnforceFunc func(c *gin.Context) (bool, error)

// groups of domain, resource and action
// if resource is prefixed with ":", it means a parameter in the path
// if resource contains "+", it means a list of resources, like :namespace+:name
func PathGuard(domain, resource, action string) EnforceFunc {
	return func(c *gin.Context) (bool, error) {
		sess := sessions.Default(c)
		usernameIf := sess.Get("username")
		if usernameIf == nil {
			return false, nil
		}
		username := usernameIf.(string)
		finalResource := resource
		if strings.Contains(resource, "+") {
			// resource is a list
			resources := strings.Split(resource, "+")
			parts := make([]string, 0, len(resources))
			for _, res := range resources {
				if strings.HasPrefix(res, ":") {
					parts = append(parts, c.Param(res[1:]))
				} else {
					parts = append(parts, res)
				}
			}
			finalResource = strings.Join(parts, "+")
		} else if strings.HasPrefix(resource, ":") {
			finalResource = c.Param(resource[1:])
		}
		ok, err := enforcer.Enforce(username, domain+"/"+finalResource, action)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		return true, nil
	}
}

type unionLogic string

const (
	unionOr  unionLogic = "OR"
	unionAnd unionLogic = "AND"
)

func unionHelper(logic unionLogic, enforces ...EnforceFunc) EnforceFunc {
	allPass := logic == unionAnd
	return func(c *gin.Context) (bool, error) {
		for _, enforce := range enforces {
			ok, err := enforce(c)
			if err != nil {
				return false, err
			}
			if ok == !allPass {
				return !allPass, nil
			}
		}
		return allPass, nil
	}
}

func OR(enforces ...EnforceFunc) EnforceFunc {
	return unionHelper(unionOr, enforces...)
}

func AND(enforces ...EnforceFunc) EnforceFunc {
	return unionHelper(unionAnd, enforces...)
}
