package utils

import (
	"github.com/oceanbase/ob-operator/pkg/controllers/observer/sql"
)

func NewDefaultDBConnectProperties(ip string) *sql.DBConnectProperties {
	return &sql.DBConnectProperties{
		IP:       ip,
		Port:     2881,
		User:     "root",
		Password: "",
		Database: "oceanbase",
		Timeout:  10,
	}
}

func NewDefaultSqlOperator(ip string) *sql.SqlOperator {
	return sql.NewSqlOperator(NewDefaultDBConnectProperties(ip))
}
