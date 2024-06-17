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

package response

type DashboardInfo struct {
	AppName          string `json:"appName" binding:"required"`
	Version          string `json:"version" binding:"required"`
	PublicKey        string `json:"publicKey" binding:"required"`
	ReportStatistics bool   `json:"reportStatistics" binding:"required"`
	ReportHost       string `json:"reportHost" binding:"required"`

	ConfigurableInfo ConfigurableInfo `json:"configurableInfo" binding:"required"`
}

type ConfigurableInfo struct {
	OdcURL string `json:"odcURL" binding:"required"`
}
