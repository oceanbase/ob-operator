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

package param

import (
	"github.com/oceanbase/ob-operator/internal/dashboard/model/common"
	"github.com/oceanbase/ob-operator/internal/dashboard/model/k8s"
)

type NodeOperation string

const (
	OperationOverwrite NodeOperation = "overwrite"
	OperationDelete    NodeOperation = "delete"
)

type CreateNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type QueryEventParam struct {
	ObjectType string `json:"objectType" query:"objectType" binding:"omitempty"`
	Type       string `json:"type" query:"type" binding:"omitempty"`
	Name       string `json:"name" query:"name" binding:"omitempty"`
	Namespace  string `json:"namespace" query:"namespace" binding:"omitempty"`
}

type NodeLabels struct {
	Labels []common.KVPair `json:"labels"`
}

type NodeTaints struct {
	Taints []k8s.Taint `json:"taints"`
}

type labelOperation struct {
	Operation NodeOperation `json:"operation" binding:"required"`
	Key       string        `json:"key" binding:"required"`
	Value     string        `json:"value" binding:"omitempty"`
}

type taintOperation struct {
	Operation NodeOperation `json:"operation" binding:"required"`
	Key       string        `json:"key" binding:"required"`
	Value     string        `json:"value" binding:"omitempty"`
	Effect    string        `json:"effect" binding:"required"`
}

type BatchUpdateNodesParam struct {
	Nodes           []string         `json:"nodes" binding:"required"`
	LabelOperations []labelOperation `json:"labelOperations"`
	TaintOperations []taintOperation `json:"taintOperations"`
}
