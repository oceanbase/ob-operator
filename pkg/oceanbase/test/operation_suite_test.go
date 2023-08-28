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

package test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	host        string
	port        int64
	user        string
	password    string
	sysUser     string
	sysPassword string
	tenant      string
	database    string
)

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() {
	flag.StringVar(&tenant, "tenant", "", "ob tenant")
	flag.StringVar(&user, "user", "root", "ob database user")
	flag.StringVar(&sysUser, "user", "root", "ob database user")
	flag.StringVar(&password, "password", "root", "password to log in")
	flag.StringVar(&sysPassword, "password", "root", "password to log in")
	flag.StringVar(&database, "database", "oceanbase", "ob database")
	flag.StringVar(&host, "host", "", "observer host")
	flag.Int64Var(&port, "port", 2881, "observer port")
}

func TestOperation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Operation Suite")
}

func printSlice[T any](s []T, extraMsg ...interface{}) {
	for _, msg := range extraMsg {
		GinkgoWriter.Println("[TEST INFO]", msg)
	}
	for i, v := range s {
		GinkgoWriter.Printf("%d# object: %+v\n", i, v)
	}
}
