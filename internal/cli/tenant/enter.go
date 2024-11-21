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
package tenant

// Flags and FlagSets for tenant management
const (
	// FlagSets for tenant
	FLAGSET_UNIT    = "unit"
	FLAGSET_RESTORE = "restore"
	FLAGSET_ZONE    = "zone"

	// Basic Flags
	FLAG_TENANT_NAME  = "tenant-name"
	FLAG_CLUSTER_NAME = "cluster"
	FLAG_NAMESPACE    = "namespace"

	// Other Flags
	FLAG_ROOTPASSWD         = "root-password"
	FLAG_FORCE              = "force"
	FLAG_CHARSET            = "charset"
	FLAG_CONNECT_WHITE_LIST = "connect-white-list"
	FLAG_FROM               = "from"
	FLAG_ZONE_PRIORITY      = "priority"

	// unit-resource-related flags
	FLAG_UNIT_NUMBER   = "unit-number"
	FLAG_MAX_IOPS      = "max-iops"
	FLAG_MIN_IOPS      = "min-iops"
	FLAG_IOPS_WEIGHT   = "iops-weight"
	FLAG_CPU_COUNT     = "cpu-count"
	FLAG_MEMORY_SIZE   = "memory-size"
	FLAG_LOG_DISK_SIZE = "log-disk-size"

	// restore flags
	FLAG_RESTORE             = "restore"
	FLAG_RESTORE_TYPE        = "restore-type"
	FLAG_ARCHIVE_SOURCE      = "archive-source"
	FLAG_BAK_ENCRYPTION_PASS = "bak-encryption-password"
	FLAG_BAK_DATA_SOURCE     = "bak-data-source"
	FLAG_OSS_ACCESS_ID       = "oss-access-id"
	FLAG_OSS_ACCESS_KEY      = "oss-access-key"
	FLAG_UNLIMITED           = "unlimited"
	FLAG_UNTIL_TIMESTAMP     = "until-timestamp"
	FLAG_PASSWD              = "password"
)

// Default values for tenant management
const (
	// Default values for int and string flags
	DEFAULT_NAMESPACE          = "default"
	DEFAULT_CHARSET            = "utf8mb4"
	DEFAULT_CONNECT_WHITE_LIST = "%"
	DEFAULT_UNIT_NUMBER        = 1
	DEFAULT_MAX_IOPS           = 1024
	DEFAULT_MIN_IOPS           = 1024
	DEFAULT_IOPS_WEIGHT        = 1
	DEFAULT_CPU_COUNT          = "1"
	DEFAULT_MEMORY_SIZE        = "2Gi"
	DEFAULT_LOG_DISK_SIZE      = "4Gi"
	DEFAULT_RESTORE_TYPE       = "OSS"

	// Default values for bool flags
	DEFAULT_UNLIMITED_FLAG = true
	DEFAULT_FORCE_FLAG     = false
	DEFAULT_RESTORE_FLAG   = false

	// Default Tenant name for demo cmd
	DEFAULT_TENANT_NAME_IN_K8S = "t1"
	DEFAULT_TENANT_NAME        = "t1"
)

// using in `demo` command, tenant resource name in k8s
const (
	FLAG_TENANT_NAME_IN_K8S = "tenant-name-in-k8s"
)
