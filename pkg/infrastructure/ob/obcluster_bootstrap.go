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

package ob

import (
	"fmt"
	"strings"
)

type ClusterType string

const (
	PRIMARY ClusterType = "PRIMARY"
	STANDBY ClusterType = "STANDBY"
)

const (
	bootstrapPrimarySql         = "ALTER SYSTEM BOOTSTRAP %v"
	bootstrapStandbySql         = "ALTER SYSTEM BOOTSTRAP CLUSTER STANDBY %v"
	bootstrapPrimaryInfoSqlPart = " PRIMARY_CLUSTER_ID %v PRIMARY_ROOTSERVICE_LIST '%v'"
	bootstrapRSInfoSqlPart      = "REGION '%v' ZONE '%v' SERVER '%v:%v'"
)

type BootstrapParam struct {
	ClusterType ClusterType  `json:"clusterType"` // cluster type, primary or standby
	PrimaryInfo *PrimaryInfo `json:"primaryInfo"` // provided when bootstrap standby clusters, nil if not needed
	RSInfoList  []RSInfo     `json:"rsInfoList"`  // root server info
}

type PrimaryInfo struct {
	PrimaryClusterId       int    `json:"primaryClusterId"`
	PrimaryRootServiceList string `json:"primaryRootServiceList"`
}

type RSInfo struct {
	Region string     `json:"region"`
	Zone   string     `json:"zone"`
	Server ServerInfo `json:"server"`
}

type ServerInfo struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

func GenerateBootstrapSQL(param BootstrapParam) string {
	var rsInfos []string
	for _, rsInfo := range param.RSInfoList {
		rsInfoString := fmt.Sprintf(bootstrapRSInfoSqlPart, rsInfo.Region, rsInfo.Zone, rsInfo.Server.Ip, rsInfo.Server.Port)
		rsInfos = append(rsInfos, rsInfoString)
	}
	rsInfoList := strings.Join(rsInfos, ", ")
	var sql string
	if param.ClusterType == PRIMARY {
		sql = fmt.Sprintf(bootstrapPrimarySql, rsInfoList)
	} else if param.ClusterType == STANDBY {
		sql = fmt.Sprintf(bootstrapStandbySql, rsInfoList)
		if param.PrimaryInfo != nil {
			sql += fmt.Sprintf(bootstrapPrimaryInfoSqlPart, param.PrimaryInfo.PrimaryClusterId, param.PrimaryInfo.PrimaryRootServiceList)
		}
	} else {
		return ""
	}
	return sql
}
