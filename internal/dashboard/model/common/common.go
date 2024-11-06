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

package common

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ResourceSpec struct {
	Cpu      int64 `json:"cpu"`
	MemoryGB int64 `json:"memory"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass"`
	SizeGB       int64  `json:"size"`
}

type SelectorExpression struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator,omitempty"`
	Values   []string `json:"values,omitempty"`
}

type AffinityType string

type AffinitySpec struct {
	SelectorExpression `json:",inline"`
	Type               AffinityType `json:"type"`
	Weight             int32        `json:"weight,omitempty"`
	Preferred          bool         `json:"preferred,omitempty"`
}

type TolerationSpec struct {
	KVPair            `json:",inline"`
	Operator          string `json:"operator"`
	Effect            string `json:"effect"`
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty"`
}

type ClusterMode string
