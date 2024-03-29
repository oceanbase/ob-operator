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

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

// OceanBaseDataSource implements the database.DataSource interface for OceanBase.
type OceanBaseDataSource struct {
	Address  string
	Port     int64
	User     string
	Tenant   string
	Password string
	Database string
}

func NewOceanBaseDataSource(address string, port int64, user, tenant, password, database string) *OceanBaseDataSource {
	return &OceanBaseDataSource{
		Address:  address,
		Port:     port,
		User:     user,
		Tenant:   tenant,
		Password: password,
		Database: database,
	}
}

func (*OceanBaseDataSource) DriverName() string {
	return "mysql"
}

func (ds *OceanBaseDataSource) GetAddress() string {
	return ds.Address
}

func (ds *OceanBaseDataSource) GetPort() int64 {
	return ds.Port
}

func (ds *OceanBaseDataSource) GetUser() string {
	return fmt.Sprintf("%s@%s", ds.User, ds.Tenant)
}

func (ds *OceanBaseDataSource) GetPassword() string {
	return ds.Password
}

func (ds *OceanBaseDataSource) GetDatabase() string {
	return ds.Database
}

func (ds *OceanBaseDataSource) DataSourceName() string {
	passwordPart := ""
	tenantPart := ""
	if ds.Password != "" {
		passwordPart = fmt.Sprintf(":%s", ds.Password)
	}
	if !(ds.Tenant == "" || ds.Tenant == oceanbaseconst.SysTenant) {
		// fix: bootstrap stage will fail if concat this part after v4.2.0
		tenantPart = fmt.Sprintf("@%s", ds.Tenant)
	}
	if ds.Database != "" {
		return fmt.Sprintf("%s%s%s@tcp(%s:%d)/%s?multiStatements=true&interpolateParams=true", ds.User, tenantPart, passwordPart, ds.Address, ds.Port, ds.Database)
	}
	return fmt.Sprintf("%s%s%s@tcp(%s:%d)/", ds.User, tenantPart, passwordPart, ds.Address, ds.Port)
}

func (ds *OceanBaseDataSource) ID() string {
	h := md5.New()
	key := fmt.Sprintf("%s@%s@%s:%d/%s", ds.User, ds.Tenant, ds.Address, ds.Port, ds.Database)
	_, err := h.Write([]byte(key))
	if err != nil {
		return key
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (ds *OceanBaseDataSource) String() string {
	return fmt.Sprintf("address: %s, port: %d, user: %s, tenant: %s, database: %s", ds.Address, ds.Port, ds.User, ds.Tenant, ds.Database)
}
