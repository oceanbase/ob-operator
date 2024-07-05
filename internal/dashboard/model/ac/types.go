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

type Object string

type Action string

type Domain string

type Policy struct {
	Domain Domain `json:"domain" binding:"required"`
	Object Object `json:"object" binding:"required"`
	Action Action `json:"action" binding:"required"`
}

func (p Policy) String() string {
	return string(p.Domain) + ":" + string(p.Object) + ":" + string(p.Action)
}

// ComposeDomainObject returns the domain and object of the policy in format "domain/object"
func (p Policy) ComposeDomainObject() string {
	return string(p.Domain) + "/" + string(p.Object)
}

func NewPolicy(domain, object, action string) Policy {
	return Policy{
		Domain: Domain(domain),
		Object: Object(object),
		Action: Action(action),
	}
}
