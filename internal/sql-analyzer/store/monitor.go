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

package store

import (
	"context"
	"database/sql"
	"time"

	logger "github.com/sirupsen/logrus"
)

func StartMemoryMonitoring(ctx context.Context, db *sql.DB, l *logger.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rows, err := db.QueryContext(ctx, "SELECT * FROM duckdb_memory()")
				if err != nil {
					l.Warnf("Failed to query duckdb_memory: %v", err)
					continue
				}

				cols, _ := rows.Columns()
				for rows.Next() {
					columns := make([]any, len(cols))
					columnPointers := make([]any, len(cols))
					for i := range columns {
						columnPointers[i] = &columns[i]
					}

					if err := rows.Scan(columnPointers...); err != nil {
						l.Warnf("Failed to scan duckdb_memory row: %v", err)
						break
					}

					rowMap := make(map[string]any)
					for i, colName := range cols {
						val := columnPointers[i].(*any)
						rowMap[colName] = *val
					}
					l.Infof("DuckDB Memory: %v", rowMap)
				}
				rows.Close()
			case <-ctx.Done():
				return
			}
		}
	}()
}
