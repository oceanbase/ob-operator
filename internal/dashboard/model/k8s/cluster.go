/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package k8s

type K8sClusterInfo struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
}

type UpdateK8sClusterParam struct {
	Description string `json:"description,omitempty"`
	KubeConfig  string `json:"kubeConfig,omitempty"`
}

type CreateK8sClusterParam struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	KubeConfig  string `json:"kubeConfig" binding:"required"`
}
