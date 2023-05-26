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

package model

import "k8s.io/apimachinery/pkg/api/resource"

type ResourceUnitV3 struct {
	MaxCPU     resource.Quantity
	MinCPU     resource.Quantity
	MemorySize resource.Quantity

	MaxIops       int
	MinIops       int
	MaxDiskSize   resource.Quantity
	MaxSessionNum int
}

type ResourceUnitV4 struct {
	MaxCPU     resource.Quantity
	MemorySize resource.Quantity

	MinCPU      resource.Quantity
	MaxIops     int
	MinIops     int
	IopsWeight  int
	LogDiskSize resource.Quantity
}
