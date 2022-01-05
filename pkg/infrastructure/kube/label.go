/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package kube

import (
	"k8s.io/apimachinery/pkg/labels"
)

func GetLabelSelector(key string, value string) labels.Selector {
	// Key is empty
	if len(key) == 0 {
		return labels.SelectorFromSet(map[string]string{})
	}
	return labels.SelectorFromSet(labels.Set{key: value})
}

func AddLabel(labels map[string]string, key string, value string) map[string]string {
	if key == "" {
		return labels
	}
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[key] = value
	return labels
}
