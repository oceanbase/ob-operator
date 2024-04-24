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

package oceanbase

import "k8s.io/apimachinery/pkg/api/resource"

const (
	DefaultDiskExpandPercent  = 10
	DefaultLogPercent         = 80
	InitialDataDiskUsePercent = 20
	DefaultDiskUsePercent     = 95
	DefaultMemoryLimitPercent = 90
	GigaConverter             = 1 << 30
	MegaConverter             = 1 << 20
)

const (
	DefaultMemoryLimitSize  = "0M"
	DefaultDatafileMaxSize  = "0M"
	DefaultDatafileNextSize = "1G"
)

const (
	MinMemorySizeS      = "8Gi"
	MinDataDiskSizeS    = "30Gi"
	MinRedoLogDiskSizeS = "30Gi"
	MinLogDiskSizeS     = "10Gi"
)

var (
	MinMemorySize      = resource.MustParse(MinMemorySizeS)
	MinDataDiskSize    = resource.MustParse(MinDataDiskSizeS)
	MinRedoLogDiskSize = resource.MustParse(MinRedoLogDiskSizeS)
	MinLogDiskSize     = resource.MustParse(MinLogDiskSizeS)
)
