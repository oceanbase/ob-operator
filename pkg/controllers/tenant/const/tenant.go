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

package tenantconst

const (
	TenantRunning   = "Running"
	TenantCreating  = "Creating"
	TenantModifying = "Modifying"

	UnitActive = "ACTIVE"
)

const (
	Charset = "utf8mb4"
)

const (
	TypeFull     = "FULL"
	TypeReadonly = "READONLY"
	TypeLog      = "LOGONLY"
	TypeF        = "F"
	TypeR        = "R"
	TypeL        = "L"
)

const (
	MaxDiskSize   = "512Mi"
	MaxIops       = 128
	MinIops       = 128
	MaxSessionNum = 64
)
