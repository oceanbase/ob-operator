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

package model

// database entity
type OBZoneInfo struct {
	Name  string `json:"name" db:"name"`
	Value int64  `json:"value" db:"value"`
	Info  string `json:"info" db:"info"`
}

// response data
type OBZone struct {
	Name           string `json:"name"`
	Idc            string `json:"idc"`
	RecoveryStatus string `json:"recovery_status"`
	Region         string `json:"region"`
	Status         string `json:"status"`
	StorageType    string `json:"storage_type"`
	ZoneType       string `json:"zone_type"`
}

func NewOBZone(zoneName string, obzoneInfoList []OBZoneInfo) *OBZone {
	obzone := &OBZone{
		Name: zoneName,
	}
	for _, obzoneInfo := range obzoneInfoList {
		if obzoneInfo.Name == "idc" {
			obzone.Idc = obzoneInfo.Info
		}
		if obzoneInfo.Name == "recovery_status" {
			obzone.RecoveryStatus = obzoneInfo.Info
		}
		if obzoneInfo.Name == "region" {
			obzone.Region = obzoneInfo.Info
		}
		if obzoneInfo.Name == "status" {
			obzone.Status = obzoneInfo.Info
		}
		if obzoneInfo.Name == "storage_type" {
			obzone.StorageType = obzoneInfo.Info
		}
		if obzoneInfo.Name == "zone_type" {
			obzone.ZoneType = obzoneInfo.Info
		}
	}
	return obzone
}
