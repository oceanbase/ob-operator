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

package connector

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

type OceanbaseConnectProperties struct {
	Address  string
	Port     int64
	User     string
	Tenant   string
	Password string
	Database string
}

func NewOceanbaseConnectProperties(address string, port int64, user, tenant, password, database string) *OceanbaseConnectProperties {
	return &OceanbaseConnectProperties{
		Address:  address,
		Port:     port,
		User:     user,
		Tenant:   tenant,
		Password: password,
		Database: database,
	}
}

func (p *OceanbaseConnectProperties) GetDSN() string {
	if p.Database != "" {
		return fmt.Sprintf("%s@%s:%s@tcp(%s:%d)/%s?multiStatements=true&interpolateParams=true", p.User, p.Tenant, p.Password, p.Address, p.Port, p.Database)
	}
	return fmt.Sprintf("%s@%s@tcp(%s:%d)/", p.User, p.Tenant, p.Address, p.Port)
}

func (p *OceanbaseConnectProperties) HashValue() string {
	hasher := md5.New()
	key := fmt.Sprintf("%s@%s@%s:%d/%s", p.User, p.Tenant, p.Address, p.Port, p.Database)
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
