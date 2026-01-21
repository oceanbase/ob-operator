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
	"runtime"
	"time"

	logger "github.com/sirupsen/logrus"
)

func StartMemoryMonitoring(ctx context.Context, name string, db *sql.DB, l *logger.Logger) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Log Go Memory Stats
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				// bToMb helper closure
				bToMb := func(b uint64) uint64 {
					return b / 1024 / 1024
				}
				l.Infof("[%s] Go Memory: Alloc=%v MiB, Sys=%v MiB, HeapSys=%v MiB, StackSys=%v MiB, NumGC=%v",
					name, bToMb(m.Alloc), bToMb(m.Sys), bToMb(m.HeapSys), bToMb(m.StackSys), m.NumGC)

				rows, err := db.QueryContext(ctx, "SELECT * FROM duckdb_memory()")
				if err != nil {
					l.Warnf("[%s] Failed to query duckdb_memory: %v", name, err)
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
						l.Warnf("[%s] Failed to scan duckdb_memory row: %v", name, err)
						break
					}

					rowMap := make(map[string]any)
					for i, colName := range cols {
						val := columnPointers[i].(*any)
						rowMap[colName] = *val
					}
					l.Infof("[%s] DuckDB Memory: %v", name, rowMap)
				}
				rows.Close()
			case <-ctx.Done():
				return
			}
		}
	}()
}
