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
package cluster

// Flags and FlagSets for cluster management
const (
	// FlagSets for cluster
	FLAGSET_ZONE          = "zone"
	FLAGSET_OBSERVER      = "observer"
	FLAGSET_MONITOR       = "monitor"
	FLAGSET_BACKUP_VOLUME = "backup-volume"
	FLAGSET_PARAMETERS    = "parameters"

	// Flags for all the commands in cluster management
	FLAG_CLUSTER_NAME  = "cluster-name"
	FLAG_NAMESPACE     = "namespace"
	FLAG_CLUSTER_ID    = "id"
	FLAG_ROOT_PASSWORD = "root-password"
	FLAG_MODE          = "mode"
	FLAG_NAME          = "name"

	// Flags for zone-related options
	FLAG_ZONES = "zones"

	// Flags for observer-related options
	FLAG_OBSERVER_IMAGE         = "image"
	FLAG_OBSERVER_CPU           = "cpu"
	FLAG_OBSERVER_MEMORY        = "memory"
	FLAG_DATA_STORAGE_CLASS     = "data-storage-class"
	FLAG_REDO_LOG_STORAGE_CLASS = "redo-log-storage-class"
	FLAG_LOG_STORAGE_CLASS      = "log-storage-class"
	FLAG_DATA_STORAGE_SIZE      = "data-storage-size"
	FLAG_REDO_LOG_STORAGE_SIZE  = "redo-log-storage-size"
	FLAG_LOG_STORAGE_SIZE       = "log-storage-size"

	// Flags for monitor-related options
	FLAG_MONITOR_IMAGE  = "monitor-image"
	FLAG_MONITOR_CPU    = "monitor-cpu"
	FLAG_MONITOR_MEMORY = "monitor-memory"

	// Flags for backup-volume-related options
	FLAG_BACKUP_ADDRESS = "backup-storage-address"
	FLAG_BACKUP_PATH    = "backup-storage-path"

	// Flags for parameter-related options
	FLAG_PARAMETERS                    = "parameters"
	FLAG_MIN_FULL_RESOURCE_POOL_MEMORY = "min-full-resource-pool-memory"
	FLAG_SYSTEM_MEMORY                 = "system-memory"
)

// Default values for cluster management
const (
	// Default values for int and string flags
	DEFAULT_NAMESPACE              = "default"
	DEFAULT_ID                     = 0
	DEFAULT_OBSERVER_IMAGE         = "quay.io/oceanbase/oceanbase-cloud-native:4.2.1.6-106000012024042515"
	DEFAULT_OBSERVER_CPU           = 2
	DEFAULT_OBSERVER_MEMORY        = 10
	DEFAULT_DATA_STORAGE_CLASS     = "local-path"
	DEFAULT_REDO_LOG_STORAGE_CLASS = "local-path"
	DEFAULT_LOG_STORAGE_CLASS      = "local-path"
	DEFAULT_DATA_STORAGE_SIZE      = 50
	DEFAULT_REDO_LOG_STORAGE_SIZE  = 50
	DEFAULT_LOG_STORAGE_SIZE       = 20
	DEFAULT_MONITOR_IMAGE          = "oceanbase/obagent:4.2.1-100000092023101717"
	DEFAULT_MONITOR_CPU            = 1
	DEFAULT_MONITOR_MEMORY         = 1
	DEFAULT_NAME                   = "test"
	DEFAULT_NAME                   = "test"

	// Default values for Parameter flag
	DEFAULT_MIN_FULL_RESOURCE_POOL_MEMORY = "2147483648"
	DEFAULT_SYSTEM_MEMORY                 = "1G"
)

// Default cluster type for easier cluster creation
const (
	SINGLE_NODE = "single-node"
	THREE_NODE  = "three-node"
)
