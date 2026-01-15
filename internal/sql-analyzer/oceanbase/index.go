/*
Copyright (c) 2025 OceanBase
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

import (
	"context"
	"strings"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/operation"
)

type IndexRow struct {
	IndexName  string `db:"index_name"`
	IndexType  string `db:"index_type"`
	Uniqueness string `db:"uniqueness"`
	Status     string `db:"status"`
	ColumnName string `db:"column_name"`
}

type PrimaryKeyRow struct {
	ColumnName string `db:"column_name"`
}

func QueryTableIndexes(ctx context.Context, opMgr *operation.OceanbaseOperationManager, tenantID uint64, dbName, tableName string, tableID int64) ([]model.IndexInfo, error) {
	var indexes []model.IndexInfo
	var pkRows []PrimaryKeyRow
	err := opMgr.QueryList(ctx, &pkRows, sqlconst.GetTablePrimaryKey, tenantID, tableID)
	if err != nil {
		return nil, err
	}

	if len(pkRows) > 0 {
		var columns []string
		for _, r := range pkRows {
			columns = append(columns, r.ColumnName)
		}

		indexes = append(indexes, model.IndexInfo{
			TableName:  tableName,
			IndexType:  "PRIMARY KEY",
			Uniqueness: "UNIQUE",
			IndexName:  "PRIMARY KEY",
			Columns:    columns,
			Status:     "VALID",
		})
	}

	var rows []IndexRow
	err = opMgr.QueryList(ctx, &rows, sqlconst.GetTableIndex, tenantID, dbName, tableName)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if strings.HasPrefix(row.IndexName, "t_pk_obpk_") {
			continue
		}

		indexes = append(indexes, model.IndexInfo{
			TableName:  tableName,
			IndexType:  row.IndexType,
			Uniqueness: row.Uniqueness,
			IndexName:  row.IndexName,
			Columns:    strings.Split(row.ColumnName, ","),
			Status:     row.Status,
		})
	}

	return indexes, nil
}
