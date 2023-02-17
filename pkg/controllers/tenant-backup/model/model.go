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

type SysParameterStat struct {
	Zone      string
	SvrType   string
	SvrIP     string
	SvrPort   int64
	Name      string
	DataType  string
	Value     string
	Info      string
	Section   string
	Scope     string
	Source    string
	EditLevel string
}

type TenantArchiveDest struct {
	DestNo int64
	Name   string
	Value  string
}

type TenantBackupDest struct {
	Name  string
	Value string
}

type TenantArchiveLog struct {
	DestNo        int64
	Status        string
	StartScn      int64
	CheckpointScn int64
	BasePieceId   int64
	UsedPieceId   int64
}

type BackupJob struct {
	BackupSetId int64
	BackupType  string
	Status      string
}

type AllBackupSet struct {
	BackupSetId int64
	BackupType  string
	Status      string
}

type DeletePolicy struct {
	PolicyName     string
	RecoveryWindow string
}
