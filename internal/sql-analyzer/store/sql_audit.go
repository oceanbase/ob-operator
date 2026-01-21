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
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	duckdb "github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	apimodel "github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/common"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/const/parquet"
	sqlconst "github.com/oceanbase/ob-operator/internal/sql-analyzer/const/sql"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"

	logger "github.com/sirupsen/logrus"
)

type SqlAuditStore struct {
	ctx    context.Context
	db     *sql.DB
	path   string
	Logger *logger.Logger
}

func NewSqlAuditStore(c context.Context, path string, maxOpenConns int, l *logger.Logger) (*SqlAuditStore, error) {
	// Use an in-memory DuckDB database for operations.
	db, err := sql.Open("duckdb", "") // In-memory
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory duckdb: %w", err)
	}

	db.SetMaxOpenConns(maxOpenConns)

	// Set memory limit for in-memory DB
	conn, err := db.Conn(c)
	if err == nil {
		memLimit := os.Getenv("DUCKDB_MEMORY_LIMIT")
		if memLimit == "" {
			memLimit = "512MB"
		}
		if _, err := conn.ExecContext(c, fmt.Sprintf("PRAGMA memory_limit='%s'", memLimit)); err != nil {
			l.Warnf("Failed to set duckdb memory limit for sql audit store: %v", err)
		}
		if _, err := conn.ExecContext(c, "SET allocator_background_threads=true"); err != nil {
			l.Warnf("Failed to set allocator_background_threads for sql audit store: %v", err)
		}
		conn.Close()
	} else {
		l.Warnf("Failed to get connection to set memory limit for sql audit store: %v", err)
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", path, err)
	}

	store := &SqlAuditStore{db: db, path: path, ctx: c, Logger: l}
	return store, nil
}

func (s *SqlAuditStore) GetLastRequestIDs() (map[string]uint64, error) {
	files, err := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob parquet files: %w", err)
	}

	if len(files) == 0 {
		return make(map[string]uint64), nil
	}

	sort.Slice(files, func(i, j int) bool {
		timeI, errI := parseTimeFromFileName(files[i])
		timeJ, errJ := parseTimeFromFileName(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.Before(timeJ)
	})

	mostRecentFile := files[len(files)-1]

	// Query only the most recent file.
	query := fmt.Sprintf(sqlconst.GetMaxRequestIDFromParquet, mostRecentFile)

	rows, err := s.db.Query(query)
	if err != nil {
		s.Logger.Warnf("Error querying latest parquet file %s: %v", mostRecentFile, err)
		return make(map[string]uint64), nil
	}
	defer rows.Close()

	lastRequestIDs := make(map[string]uint64)
	for rows.Next() {
		var svrIP string
		var maxRequestID uint64
		if err := rows.Scan(&svrIP, &maxRequestID); err != nil {
			return nil, err
		}
		lastRequestIDs[svrIP] = maxRequestID
	}
	return lastRequestIDs, nil
}

func (s *SqlAuditStore) InsertBatch(resultsSlices [][]model.SqlAudit) error {
	if len(resultsSlices) == 0 {
		return nil
	}

	conn, err := s.db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	tempTableName := "sql_audit_batch_" + uuid.New().String()[:8] // Use a unique temp table name
	if _, err := conn.ExecContext(context.Background(), fmt.Sprintf(sqlconst.CreateSqlAuditTempTableTemplate, tempTableName)); err != nil {
		return fmt.Errorf("failed to create temp table: %w", err)
	}
	defer conn.ExecContext(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", tempTableName))

	// Use the appender to load data into the temp table.
	err = conn.Raw(func(driverConn any) error {
		duckdbConn, ok := driverConn.(driver.Conn)
		if !ok {
			return fmt.Errorf("failed to get raw duckdb connection")
		}
		appender, err := duckdb.NewAppenderFromConn(duckdbConn, "", tempTableName)
		if err != nil {
			return fmt.Errorf("failed to create appender: %w", err)
		}
		defer appender.Close()

		collectTime := time.Now()

		for _, results := range resultsSlices {
			for _, r := range results {
				err := appender.AppendRow(
					r.SvrIP, r.SvrPort, r.TenantId, r.TenantName, r.UserId, r.UserName, r.DBId, r.DBName, r.SqlId, r.PlanId,
					r.QuerySql, r.ClientIp, r.Event, r.FormatSqlId, r.EffectiveTenantId, r.TraceId, r.Sid, r.UserClientIp, r.TxId,
					r.Executions, r.MinRequestTime, r.MaxRequestTime, r.MaxRequestId, r.MinRequestId,
					r.ElapsedTimeSum, r.ElapsedTimeMax, r.ElapsedTimeMin,
					r.ExecuteTimeSum, r.ExecuteTimeMax, r.ExecuteTimeMin,
					r.QueueTimeSum, r.QueueTimeMax, r.QueueTimeMin,
					r.GetPlanTimeSum, r.GetPlanTimeMax, r.GetPlanTimeMin,
					r.AffectedRowsSum, r.AffectedRowsMax, r.AffectedRowsMin,
					r.ReturnRowsSum, r.ReturnRowsMax, r.ReturnRowsMin,
					r.PartitionCountSum, r.PartitionCountMax, r.PartitionCountMin,
					r.RetryCountSum, r.RetryCountMax, r.RetryCountMin,
					r.DiskReadsSum, r.DiskReadsMax, r.DiskReadsMin,
					r.RpcCountSum, r.RpcCountMax, r.RpcCountMin,
					r.MemstoreReadRowCountSum, r.MemstoreReadRowCountMax, r.MemstoreReadRowCountMin,
					r.SSStoreReadRowCountSum, r.SSStoreReadRowCountMax, r.SSStoreReadRowCountMin,
					r.RequestMemoryUsedSum, r.RequestMemoryUsedMax, r.RequestMemoryUsedMin,
					r.WaitTimeMicroSum, r.WaitTimeMicroMax, r.WaitTimeMicroMin,
					r.TotalWaitTimeMicroSum, r.TotalWaitTimeMicroMax, r.TotalWaitTimeMicroMin,
					r.NetTimeSum, r.NetTimeMax, r.NetTimeMin,
					r.NetWaitTimeSum, r.NetWaitTimeMax, r.NetWaitTimeMin,
					r.DecodeTimeSum, r.DecodeTimeMax, r.DecodeTimeMin,
					r.ApplicationWaitTimeSum, r.ApplicationWaitTimeMax, r.ApplicationWaitTimeMin,
					r.ConcurrencyWaitTimeSum, r.ConcurrencyWaitTimeMax, r.ConcurrencyWaitTimeMin,
					r.UserIoWaitTimeSum, r.UserIoWaitTimeMax, r.UserIoWaitTimeMin,
					r.ScheduleTimeSum, r.ScheduleTimeMax, r.ScheduleTimeMin,
					r.RowCacheHitSum, r.RowCacheHitMax, r.RowCacheHitMin,
					r.BloomFilterCacheHitSum, r.BloomFilterCacheHitMax, r.BloomFilterCacheHitMin,
					r.BlockCacheHitSum, r.BlockCacheHitMax, r.BlockCacheHitMin,
					r.IndexBlockCacheHitSum, r.IndexBlockCacheHitMax, r.IndexBlockCacheHitMin,
					r.ExpectedWorkerCountSum, r.ExpectedWorkerCountMax, r.ExpectedWorkerCountMin,
					r.UsedWorkerCountSum, r.UsedWorkerCountMax, r.UsedWorkerCountMin,
					r.TableScanSum, r.TableScanMax, r.TableScanMin,
					r.ConsistencyLevelStrongCount,
					r.ConsistencyLevelWeakCount,
					r.FailCountSum,
					r.RetCode4012CountSum, r.RetCode4013CountSum, r.RetCode5001CountSum, r.RetCode5024CountSum,
					r.RetCode5167CountSum, r.RetCode5217CountSum, r.RetCode6002CountSum,
					r.Event0WaitTimeSum, r.Event1WaitTimeSum, r.Event2WaitTimeSum, r.Event3WaitTimeSum,
					r.PlanTypeLocalCount, r.PlanTypeRemoteCount, r.PlanTypeDistributedCount,
					r.InnerSqlCount,
					r.MissPlanCount,
					r.ExecutorRpcCount,
					collectTime,
					collectTime,
				)
				if err != nil {
					return errors.Wrapf(err, "Failed to append row for SvrIP %s, SvrPort %d. MinRequestId: %d, MaxRequestID: %d", r.SvrIP, r.SvrPort, r.MinRequestId, r.MaxRequestId)
				}
			}
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "Failed to append data")
	}

	// Determine the target Parquet file based on the current timestamp.
	currentTime := time.Now().Format(parquet.FileTimeFormat)
	targetParquetFile := filepath.Join(s.path, fmt.Sprintf("%s-%s.parquet", currentTime, uuid.New().String()[:8]))

	// Now, copy the data from the temp table to the new parquet file.
	copySql := fmt.Sprintf(
		"COPY %s TO '%s' (FORMAT PARQUET)",
		tempTableName, targetParquetFile,
	)
	_, err = conn.ExecContext(s.ctx, copySql)
	return err
}

func (s *SqlAuditStore) Compact() error {
	conn, err := s.db.Conn(s.ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get connection")
	}
	defer conn.Close()

	smallFiles, err := filepath.Glob(filepath.Join(s.path, parquet.SmallFilePattern))
	if err != nil {
		return fmt.Errorf("failed to glob small parquet files: %w", err)
	}

	// Filter out invalid small files by running a quick check on each one.
	var validSmallFiles []string
	for _, file := range smallFiles {
		var count int64
		// This is a fast way to check if the file is readable by DuckDB.
		err := s.db.QueryRowContext(s.ctx, fmt.Sprintf("SELECT count(*) FROM read_parquet('%s')", file)).Scan(&count)
		if err != nil {
			s.Logger.Warnf("File %s appears to be invalid, deleting. Error: %v", file, err)
			if removeErr := os.Remove(file); removeErr != nil {
				s.Logger.Errorf("Failed to delete invalid file %s: %v", file, removeErr)
			}
			continue // Skip to the next file
		}
		validSmallFiles = append(validSmallFiles, file)
	}

	if len(validSmallFiles) <= 1 {
		return nil // Nothing to compact
	}

	sort.Slice(validSmallFiles, func(i, j int) bool {
		timeI, errI := parseTimeFromFileName(validSmallFiles[i])
		timeJ, errJ := parseTimeFromFileName(validSmallFiles[j])
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.Before(timeJ)
	})

	filesToCompact := validSmallFiles[:len(validSmallFiles)-1]
	if len(filesToCompact) == 0 {
		return nil
	}
	lastFileInBatch := filesToCompact[len(filesToCompact)-1]
	timestamp, err := parseTimeFromFileName(lastFileInBatch)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp from file %s: %w", lastFileInBatch, err)
	}

	// Define the compacted file path and a temporary path for atomic operation.
	compactedFile := filepath.Join(s.path, "compacted-"+timestamp.Format(parquet.FileTimeFormat)+".parquet")
	tempCompactedFile := compactedFile + ".tmp"

	// Use a single, streaming COPY command instead of loading into a temp table.
	copySql := fmt.Sprintf("COPY (SELECT * FROM read_parquet(['%s'])) TO '%s' (FORMAT PARQUET)", strings.Join(filesToCompact, "','"), tempCompactedFile)

	if _, err := conn.ExecContext(context.Background(), copySql); err != nil {
		return fmt.Errorf("failed to compact files: %w", err)
	}

	// Atomically rename the temporary compacted file to the final name.
	if err := os.Rename(tempCompactedFile, compactedFile); err != nil {
		return fmt.Errorf("failed to rename temporary compacted file: %w", err)
	}

	// Delete the original files that were compacted.
	for _, file := range filesToCompact {
		if err := os.Remove(file); err != nil {
			s.Logger.Errorf("Failed to delete old file %s: %v", file, err)
		}
	}

	return nil
}

// QueryOptions holds all the parameters for a dynamic query.
type QueryOptions struct {
	SelectExpressions []string
	Filters           map[string]any
	GroupByColumns    []string
	OrderBy           string
	SortOrder         string
	Limit             int
	Offset            int
}

func (s *SqlAuditStore) CountSqlAudits(opts *QueryOptions) (int64, error) {
	if !s.hasParquetFiles() {
		return 0, nil
	}
	var args []any
	var whereClauses []string

	for key, value := range opts.Filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s ?", key))
		args = append(args, value)
	}

	fromClause := fmt.Sprintf("FROM read_parquet('%s/*.parquet')", s.path)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	groupByClause := ""
	if len(opts.GroupByColumns) > 0 {
		groupByClause = "GROUP BY " + strings.Join(opts.GroupByColumns, ", ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT 1 %s %s %s)", fromClause, whereClause, groupByClause)
	var totalCount int64
	err := s.db.QueryRowContext(s.ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to query total count: %w", err)
	}
	return totalCount, nil
}

func (s *SqlAuditStore) QuerySqlAudits(opts *QueryOptions) ([]map[string]any, error) {
	if !s.hasParquetFiles() {
		return []map[string]any{}, nil
	}
	var args []any
	var whereClauses []string

	for key, value := range opts.Filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s ?", key))
		args = append(args, value)
	}

	fromClause := fmt.Sprintf("FROM read_parquet('%s/*.parquet')", s.path)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	groupByClause := ""
	if len(opts.GroupByColumns) > 0 {
		groupByClause = "GROUP BY " + strings.Join(opts.GroupByColumns, ", ")
	}

	selectClause := "SELECT " + strings.Join(opts.SelectExpressions, ", ")

	var orderByClause string
	if opts.OrderBy != "" {
		safeOrderBy := strings.ReplaceAll(opts.OrderBy, ";", "")
		safeSortOrder := "ASC"
		if strings.ToUpper(opts.SortOrder) == "DESC" {
			safeSortOrder = "DESC"
		}
		orderByClause = fmt.Sprintf("ORDER BY %s %s", safeOrderBy, safeSortOrder)
	}

	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", opts.Limit, opts.Offset)

	dataQuery := fmt.Sprintf("%s %s %s %s %s %s", selectClause, fromClause, whereClause, groupByClause, orderByClause, limitClause)

	rows, err := s.db.QueryContext(s.ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sql audits: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		m := make(map[string]any)
		for i, colName := range cols {
			val := columnPointers[i].(*any)
			m[colName] = *val
		}
		results = append(results, m)
	}

	return results, nil
}

// Close closes the database connection.
func (s *SqlAuditStore) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *SqlAuditStore) hasParquetFiles() bool {
	files, _ := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	return len(files) > 0
}

// DeleteOldData deletes data from parquet files older than the retention period.
func (s *SqlAuditStore) DeleteOldData(retentionDays int) error {
	if retentionDays <= 0 {
		return nil
	}
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	files, err := filepath.Glob(filepath.Join(s.path, "*.parquet"))
	if err != nil {
		return fmt.Errorf("failed to glob parquet files: %w", err)
	}

	for _, filePath := range files {
		fileTime, err := parseTimeFromFileName(filePath)
		if err != nil {
			s.Logger.Warnf("Skipping file %s with invalid date format: %v", filePath, err)
			continue
		}

		if fileTime.Before(cutoffDate) {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete old file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func (s *SqlAuditStore) StartCleanupWorker() {
	// Start the cleanup routine for old data
	retentionStr := os.Getenv("DATA_RETENTION_DAYS")
	retentionDays, err := strconv.Atoi(retentionStr)
	if err != nil {
		s.Logger.Fatalf("Invalid or missing DATA_RETENTION_DAYS environment variable: %v", err)
	}

	go func() {
		// Run cleanup once at startup
		s.Logger.Info("Running initial cleanup of old data...")
		if err := s.DeleteOldData(retentionDays); err != nil {
			s.Logger.Errorf("Error during initial data cleanup: %v", err)
		}

		// Then run periodically
		cleanupTicker := time.NewTicker(24 * time.Hour)
		defer cleanupTicker.Stop()
		for {
			select {
			case <-cleanupTicker.C:
				s.Logger.Info("Running periodic cleanup of old data...")
				if err := s.DeleteOldData(retentionDays); err != nil {
					s.Logger.Errorf("Error during periodic data cleanup: %v", err)
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

func (s *SqlAuditStore) StartBackgroundWorkers() {
	s.StartCleanupWorker()
	// Start memory monitoring
	StartMemoryMonitoring(s.ctx, s.db, s.Logger)
}

func parseTimeFromFileName(fileName string) (time.Time, error) {
	baseName := filepath.Base(fileName)
	var dateStr string
	if strings.HasPrefix(baseName, "compacted-") {
		dateStr = strings.TrimSuffix(strings.TrimPrefix(baseName, "compacted-"), ".parquet")
	} else {
		parts := strings.Split(baseName, "-")
		if len(parts) > 6 {
			dateStr = strings.Join(parts[0:6], "-")
		} else {
			return time.Time{}, fmt.Errorf("invalid small file name format")
		}
	}
	return time.Parse(parquet.FileTimeFormat, dateStr)
}

func (s *SqlAuditStore) QueryRequestStatistics(req apimodel.RequestStatisticsRequest) (*apimodel.RequestStatisticsResponse, error) {
	if !s.hasParquetFiles() {
		return &apimodel.RequestStatisticsResponse{
			ExecutionTrend: []apimodel.DailyTrend{},
			LatencyTrend:   []apimodel.DailyTrend{},
		}, nil
	}
	var args []any
	var whereClauses []string

	if req.StartTime > 0 {
		whereClauses = append(whereClauses, "max_request_time >= ?")
		args = append(args, req.StartTime*1000) // convert to microseconds
	}
	if req.EndTime > 0 {
		whereClauses = append(whereClauses, "max_request_time <= ?")
		args = append(args, req.EndTime*1000) // convert to microseconds
	}
	if req.UserName != "" {
		whereClauses = append(whereClauses, "user_name = ?")
		args = append(args, req.UserName)
	}
	if req.DatabaseName != "" {
		whereClauses = append(whereClauses, "db_name = ?")
		args = append(args, req.DatabaseName)
	}
	if req.FilterInnerSql {
		whereClauses = append(whereClauses, "inner_sql_count = 0")
	}

	fromClause := fmt.Sprintf("FROM read_parquet('%s/*.parquet')", s.path)
	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Query for totals
	totalsQuery := fmt.Sprintf(sqlconst.QueryRequestStatisticsTotals, fromClause, whereClause)

	var totalExecutions, failedExecutions, totalLatency sql.NullFloat64
	err := s.db.QueryRowContext(s.ctx, totalsQuery, args...).Scan(&totalExecutions, &failedExecutions, &totalLatency)
	if err != nil {
		return nil, fmt.Errorf("failed to query request statistics totals: %w", err)
	}

	resp := &apimodel.RequestStatisticsResponse{
		TotalExecutions:  totalExecutions.Float64,
		FailedExecutions: failedExecutions.Float64,
		TotalLatency:     totalLatency.Float64,
		ExecutionTrend:   []apimodel.DailyTrend{},
		LatencyTrend:     []apimodel.DailyTrend{},
	}

	// Query for trends
	trendsQuery := fmt.Sprintf(sqlconst.QueryRequestStatisticsTrends, fromClause, whereClause)

	rows, err := s.db.QueryContext(s.ctx, trendsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query request statistics trends: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var day string
		var executions, latency float64
		if err := rows.Scan(&day, &executions, &latency); err != nil {
			return nil, fmt.Errorf("failed to scan trend row: %w", err)
		}
		resp.ExecutionTrend = append(resp.ExecutionTrend, apimodel.DailyTrend{Date: day, Value: executions})
		resp.LatencyTrend = append(resp.LatencyTrend, apimodel.DailyTrend{Date: day, Value: latency})
	}

	return resp, nil
}

func (s *SqlAuditStore) QuerySqlHistoryInfo(req apimodel.SqlHistoryRequest) (*apimodel.SqlHistoryResponse, error) {
	if !s.hasParquetFiles() {
		return &apimodel.SqlHistoryResponse{
			ExecutionTrend: []apimodel.PlanTypeTrend{},
			LatencyTrend:   []apimodel.LatencyTrendItem{},
		}, nil
	}
	resp := &apimodel.SqlHistoryResponse{
		ExecutionTrend: []apimodel.PlanTypeTrend{},
		LatencyTrend:   []apimodel.LatencyTrendItem{},
	}

	// Execution Trend
	execTrendQuery := fmt.Sprintf(sqlconst.QueryExecutionTrend, req.Interval, s.path)

	rows, err := s.db.QueryContext(s.ctx, execTrendQuery, req.SqlId, req.StartTime*1000000, req.EndTime*1000000)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution trend: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trend apimodel.PlanTypeTrend
		if err := rows.Scan(&trend.Time, &trend.Local, &trend.Remote, &trend.Distributed); err != nil {
			return nil, fmt.Errorf("failed to scan execution trend row: %w", err)
		}
		resp.ExecutionTrend = append(resp.ExecutionTrend, trend)
	}

	// Latency Trend
	if len(req.LatencyColumns) > 0 {
		var selectExpressions []string
		for _, col := range req.LatencyColumns {
			// Basic validation to prevent SQL injection
			safeCol := strings.ReplaceAll(col, ";", "")
			expr := common.BuildMetricExpression(safeCol)
			if expr != "" {
				selectExpressions = append(selectExpressions, fmt.Sprintf("%s AS %s", expr, safeCol))
			} else {
				// Fallback for unknown metrics, assume they are raw columns we want to sum
				selectExpressions = append(selectExpressions, fmt.Sprintf("sum(%s) AS %s", safeCol, safeCol))
			}
		}

		latencyTrendQuery := fmt.Sprintf(sqlconst.QueryLatencyTrend, req.Interval, strings.Join(selectExpressions, ", "), s.path)

		rows, err := s.db.QueryContext(s.ctx, latencyTrendQuery, req.SqlId, req.StartTime*1000000, req.EndTime*1000000)
		if err != nil {
			return nil, fmt.Errorf("failed to query latency trend: %w", err)
		}
		defer rows.Close()

		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for latency trend: %w", err)
		}

		for rows.Next() {
			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				return nil, fmt.Errorf("failed to scan latency trend row: %w", err)
			}

			item := apimodel.LatencyTrendItem{
				Value: make(map[string]float64),
			}
			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				if colName == "time_bucket" {
					item.Time = (*val).(int64)
				} else {
					if v, ok := (*val).(float64); ok {
						item.Value[colName] = v
					}
				}
			}
			resp.LatencyTrend = append(resp.LatencyTrend, item)
		}
	}

	return resp, nil
}

func (s *SqlAuditStore) QuerySqlDetailInfo(planStore *PlanStore, req apimodel.SqlDetailRequest) (*apimodel.SqlDetailResponse, error) {
	if !s.hasParquetFiles() {
		return &apimodel.SqlDetailResponse{
			Plans:   []apimodel.PlanStats{},
			Tables:  []apimodel.TableInfo{},
			Indexes: []apimodel.IndexInfo{},
		}, nil
	}
	resp := &apimodel.SqlDetailResponse{
		Plans:   []apimodel.PlanStats{},
		Tables:  []apimodel.TableInfo{},
		Indexes: []apimodel.IndexInfo{},
	}

	// Fetch QuerySql for the given SqlId
	querySqlQuery := fmt.Sprintf(sqlconst.QuerySqlById, s.path)

	var querySql string
	err := s.db.QueryRowContext(s.ctx, querySqlQuery, req.SqlId, req.StartTime*1000000, req.EndTime*1000000).Scan(&querySql)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query QuerySql: %w", err)
	}
	resp.QuerySql = querySql

	// Plans
	planStats, err := planStore.GetPlanStatsBySqlId(req.SqlId)
	if err != nil {
		return nil, err
	}

	for _, ps := range planStats {
		gmt, _ := time.Parse("2006-01-02 15:04:05", ps.GeneratedTime)
		resp.Plans = append(resp.Plans, apimodel.PlanStats{
			TenantID:      ps.TenantID,
			SvrIP:         ps.SvrIP,
			SvrPort:       ps.SvrPort,
			PlanID:        ps.PlanID,
			PlanHash:      ps.PlanHash,
			GeneratedTime: gmt.Unix(),
			IoCost:        ps.IoCost,
			CpuCost:       ps.CpuCost,
			Cost:          ps.Cost,
			RealCost:      ps.RealCost,
		})
	}

	// Tables
	tables, err := planStore.GetTableInfoBySqlId(req.SqlId)
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		resp.Tables = append(resp.Tables, apimodel.TableInfo{
			DatabaseName: table.DatabaseName,
			TableName:    table.TableName,
			TableID:      table.TableID,
		})
	}

	return resp, nil
}
