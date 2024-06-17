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

type OBConnection struct {
	Namespace  string `json:"namespace"`
	Cluster    string `json:"cluster,omitempty"`
	Tenant     string `json:"tenant,omitempty"`
	Pod        string `json:"pod"`
	ClientIP   string `json:"clientIp"`
	TerminalID string `json:"terminalId,omitempty"`
	User       string `json:"user"`
	Password   string `json:"-"`
	Host       string `json:"-"`

	OdcVisitURL string `json:"odcVisitURL,omitempty"`
}
