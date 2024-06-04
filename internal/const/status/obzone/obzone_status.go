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

package obzone

const (
	New                  = "new"
	MigrateFromExisting  = "migrate from existing"
	Maintaining          = "maintaining"
	Running              = "running"
	AddOBServer          = "add observer"
	DeleteOBServer       = "delete observer"
	Deleting             = "deleting"
	Upgrade              = "upgrade"
	BootstrapReady       = "bootstrap ready"
	FinalizerFinished    = "finalizer finished"
	ScaleVertically      = "scale vertically"
	ExpandPVC            = "expand pvc"
	ModifyServerTemplate = "modify server template"
	RollingUpdateServers = "rolling update servers"
)
