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
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestDataSourceOptionsWithOptionalDatabase(t *testing.T) {
	for _, name := range []string{"", "oceanbase"} {
		ds := NewOceanBaseDataSource("127.0.0.1", 2881, "root", "tenant1", "", name)
		cfg, err := mysql.ParseDSN(ds.DataSourceName())
		if err != nil {
			t.Fatal(err)
		}
		if cfg.DBName != name || !cfg.InterpolateParams || cfg.MultiStatements != (name != "") {
			t.Errorf("database %q: got db=%q interpolate=%t multi=%t", name, cfg.DBName, cfg.InterpolateParams, cfg.MultiStatements)
		}
	}
}
