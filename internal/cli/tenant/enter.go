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

const (
	// Flagsets for tenant
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
	FLAG_RESTORE_TYPE        = "type"
	FLAG_ARCHIVE_SOURCE      = "archive-source"
	FLAG_BAK_ENCRYPTION_PASS = "bak-encryption-password"
	FLAG_BAK_DATA_SOURCE     = "bak-data-source"
	FLAG_OSS_ACCESS_ID       = "oss-access-id"
	FLAG_OSS_ACCESS_KEY      = "oss-access-key"
	FLAG_UNLIMITED           = "unlimited"
	FLAG_UNTIL_TIMESTAMP     = "until-timestamp"
	FLAG_PASSWD              = "password"
)
