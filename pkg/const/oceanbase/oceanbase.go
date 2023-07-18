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

const (
	BootstrapTimeoutSeconds    = 300
	DefaultStateWaitTimeout    = 300
	ServerDeleteTimeoutSeconds = 86400
)

const (
	SqlPort = 2881
	RpcPort = 2882
)

const (
	SqlPortName = "sql"
	RpcPortName = "rpc"
)

const (
	ProbeCheckPeriodSeconds = 2
	ProbeCheckDelaySeconds  = 5
)

const (
	ContainerName      = "observer"
	InstallPath        = "/home/admin/oceanbase"
	DataPath           = "/home/admin/data-file"
	ClogPath           = "/home/admin/data-log"
	LogPath            = "/home/admin/log"
	BackupPath         = "/ob-backup"
	DataVolumeSuffix   = "data-file"
	ClogVolumeSuffix   = "data-log"
	LogVolumeSuffix    = "log"
	BackupVolumeSuffix = "backup"
)

const (
	RootUser     = "root"
	ProxyUser    = "proxyro"
	OperatorUser = "operator"
)

const (
	SysTenant       = "sys"
	DefaultDatabase = "oceanbase"
	DefaultRegion   = "default"
)

const (
	LabelRefOBCluster = "ref-obcluster"
	LabelRefOBZone    = "ref-obzone"
	LabelRefOBServer  = "ref-observer"
	LabelRefUID       = "ref-uid"
)

const (
	AllPrivilege    = "all"
	SelectPrivilege = "select"
)
