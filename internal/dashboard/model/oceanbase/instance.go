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

type OBInstance struct {
	Type      OBInstanceType `json:"type" binding:"required"`
	OBCluster string         `json:"obcluster,omitempty"`
	OBZone    string         `json:"obzone,omitempty"` // obzone may exist in labels
	OBServer  string         `json:"observer,omitempty"`
	OBTenant  string         `json:"obtenant,omitempty"`
}

func (o *OBInstance) Equals(other *OBInstance) bool {
	if o.Type != other.Type {
		return false
	}
	switch o.Type {
	case TypeOBCluster:
		return o.OBCluster == other.OBCluster
	case TypeOBServer:
		return o.OBServer == other.OBServer
	case TypeOBTenant:
		return (o.OBCluster == other.OBCluster) && (o.OBTenant == other.OBTenant)
	default:
		return false
	}
}
