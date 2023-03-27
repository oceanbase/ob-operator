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
package tenant

import (
	"time"
)

const (
	ClusterReady = "Ready"

	TryInterval                   = 1 * time.Second
	ApplyWaitTime                 = 5 * time.Second
	TenantCreateTimeout           = 300 * time.Second
	TenantModifyTimeout           = 300 * time.Second
	OBClusterReadyimeout          = 1000 * time.Second
	OBClusterUpdateTReadyimeout   = 300 * time.Second
	StatefulappUpdateReadyTimeout = 60 * time.Second
)
