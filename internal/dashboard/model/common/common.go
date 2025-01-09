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
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type ResourceSpec struct {
	Cpu      int64 `json:"cpu" binding:"required"`
	MemoryGB int64 `json:"memory" binding:"required"`
}

type StorageSpec struct {
	StorageClass string `json:"storageClass"`
	SizeGB       int64  `json:"size" binding:"required"`
}

type SelectorExpression struct {
	Key      string   `json:"key" binding:"required"`
	Operator string   `json:"operator,omitempty" binding:"required"`
	Values   []string `json:"values,omitempty"`
}

type AffinityType string

type AffinitySpec struct {
	SelectorExpression `json:",inline"`
	// Enum: NODE, POD, POD_ANTI
	Type      AffinityType `json:"type" binding:"required"`
	Weight    int32        `json:"weight,omitempty"`
	Preferred bool         `json:"preferred,omitempty"`
}

type TolerationSpec struct {
	KVPair            `json:",inline"`
	Operator          string `json:"operator" binding:"required"`
	Effect            string `json:"effect" binding:"required"`
	TolerationSeconds *int64 `json:"tolerationSeconds,omitempty"`
}

type ClusterMode string
