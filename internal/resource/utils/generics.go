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

package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetRef[T any](val T) *T {
	return &val
}

func IsZero[T comparable](val T) bool {
	return val == *(new(T))
}

func Min[T int | int64 | uint | uint64 | float64 | float32](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func GetAnnotationField[T client.Object](obj T, key string) (string, bool) {
	annos := obj.GetAnnotations()
	if annos == nil {
		return "", false
	}
	if val, ok := annos[key]; ok {
		return val, true
	}
	return "", false
}
