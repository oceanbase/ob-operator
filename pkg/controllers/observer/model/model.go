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

type AllServer struct {
	ID               int64
	Zone             string
	SvrIP            string
	SvrPort          int64
	InnerPort        int64
	WithRootService  int64
	WithPartition    int64
	Status           string
	StartServiceTime int64
}

type AllZone struct {
	Zone  string
	Name  string
	Value string
	Info  string
}

type AllVirtualCoreMeta struct {
	Zone    string
	SvrIP   string
	SvrPort int64
	Role    int64
}

type RSJobStatus struct {
	JobStatus string
	Progress  int64
}

type SysParameterStat struct {
	Zone      string
	SvrIP     string
	SvrPort   int64
	Name      string
	Value     string
	Scope     string
	EditLevel string
}

type ZoneUpGradeMode struct {
	Zone    string
	SvrIP   string
	SvrPort int64
	Value   string
}

type ZoneLeaderCount struct {
	Zone        string
	LeaderCount int64
}

type AllUnit struct {
	UnitIp             string
	ResourcePoolIp     string
	GroupId            string
	Zone               string
	SvrIP              string
	SvrPort            int64
	MigrateFromSvrIp   string
	MigrateFromSvrPort int64
	ManualMigrate      int64
	Status             string
	ReplicaType        int64
}

type ClogStat struct {
	SvrIP     string
	SvrPort   int64
	IsOffline int8
	IsInSync  int8
}
