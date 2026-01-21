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

package main

import (
	"context"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	webserver "github.com/oceanbase/ob-operator/internal/server"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/collector"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/config"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/handler"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/router"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/store"
)

func newLogger(filename string, level string) *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	}))

	lLevel, err := logrus.ParseLevel(level)
	if err != nil {
		lLevel = logrus.InfoLevel
	}
	l.SetLevel(lLevel)
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	return l
}

func startHttpServer(ctx context.Context, logger *logrus.Logger) *webserver.HTTPServer {
	httpServer := webserver.NewHTTPServer(ctx)
	httpServer.Router.Use(gin.LoggerWithWriter(logger.Out), gin.RecoveryWithWriter(logger.Out))
	router.Register(httpServer.Router)
	logger.Info("Successfully registered router")
	go func() {
		err := httpServer.Run()
		if err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Errorln("Start server failed")
			os.Exit(1)
		}
	}()
	return httpServer
}

func main() {
	// Read configuration from environment variables.
	namespace := os.Getenv("NAMESPACE")
	obtenant := os.Getenv("OBTENANT")
	dataPath := os.Getenv("DATA_PATH")
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	if namespace == "" || obtenant == "" {
		logrus.Fatalf("NAMESPACE, OBTENANT environment variables must be set.")
	}
	if dataPath == "" {
		dataPath = "."
	}

	// Initialize Loggers
	logDir := filepath.Join(dataPath, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logrus.Fatalf("Failed to create log directory: %v", err)
	}
	collectorLogPath := filepath.Join(logDir, "collector.log")
	analyzerLogPath := filepath.Join(logDir, "analyzer.log")

	collectorLogger := newLogger(collectorLogPath, logLevel)
	analyzerLogger := newLogger(analyzerLogPath, logLevel)

	// Set analyzer logger for handlers
	handler.HandlerLogger = analyzerLogger

	// Start pprof if enabled
	if os.Getenv("ENABLE_PPROF") == "true" {
		go func() {
			analyzerLogger.Info("Starting pprof server on :6060")
			if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
				analyzerLogger.WithError(err).Error("Failed to start pprof server")
			}
		}()
	}

	// Set up a context that is canceled on interruption signals.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Configure collection interval
	collectionIntervalSeconds := 30
	intervalStr := os.Getenv("COLLECTION_INTERVAL_SECONDS")
	if intervalStr != "" {
		if val, err := strconv.Atoi(intervalStr); err == nil && val > 0 {
			collectionIntervalSeconds = val
		} else {
			analyzerLogger.Warnf("Invalid COLLECTION_INTERVAL_SECONDS value '%s', using default of 30 seconds.", intervalStr)
		}
	}

	// Configure compaction interval
	compactionIntervalSeconds := 3600
	compactionIntervalStr := os.Getenv("COMPACTION_INTERVAL_SECONDS")
	if compactionIntervalStr != "" {
		if val, err := strconv.Atoi(compactionIntervalStr); err == nil && val > 0 {
			compactionIntervalSeconds = val
		} else {
			analyzerLogger.Warnf("Invalid COMPACTION_INTERVAL_SECONDS value '%s', using default of 3600 seconds.", compactionIntervalStr)
		}
	}

	compactionThreshold := compactionIntervalSeconds / collectionIntervalSeconds
	if compactionThreshold < 1 {
		compactionThreshold = 1 // Ensure compaction runs at least after one collection cycle if interval is short
	}

	// Configure SQL audit limit
	sqlAuditLimit := 10000
	sqlAuditLimitStr := os.Getenv("SQL_AUDIT_LIMIT")
	if sqlAuditLimitStr != "" {
		if val, err := strconv.Atoi(sqlAuditLimitStr); err == nil && val > 0 {
			sqlAuditLimit = val
		} else {
			analyzerLogger.Warnf("Invalid SQL_AUDIT_LIMIT value '%s', using default of 10000.", sqlAuditLimitStr)
		}
	}

	// Configure slow SQL threshold
	slowSqlThresholdMilliSeconds := 1000 // milliseconds
	slowSqlThresholdMilliSecondsStr := os.Getenv("SLOW_SQL_THRESHOLD_MILLISECONDS")
	if slowSqlThresholdMilliSecondsStr != "" {
		if val, err := strconv.Atoi(slowSqlThresholdMilliSecondsStr); err == nil && val >= 0 {
			slowSqlThresholdMilliSeconds = val
		} else {
			analyzerLogger.Warnf("Invalid SLOW_SQL_THRESHOLD_MILLISECONDS value '%s', using default of 1000ms.", slowSqlThresholdMilliSecondsStr)
		}
	}

	// Configure worker num
	workerNum := 1
	workerNumStr := os.Getenv("PLAN_WORKER_NUM")
	if workerNumStr != "" {
		if val, err := strconv.Atoi(workerNumStr); err == nil && val > 0 {
			workerNum = val
		} else {
			analyzerLogger.Warnf("Invalid PLAN_WORKER_NUM value '%s', using default of 1.", workerNumStr)
		}
	}

	// Configure queue size
	queueSize := 100
	queueSizeStr := os.Getenv("PLAN_QUEUE_SIZE")
	if queueSizeStr != "" {
		if val, err := strconv.Atoi(queueSizeStr); err == nil && val > 0 {
			queueSize = val
		} else {
			analyzerLogger.Warnf("Invalid PLAN_QUEUE_SIZE value '%s', using default of 100.", queueSizeStr)
		}
	}

	// Configure plan cache size
	planCacheSize := 10000
	planCacheSizeStr := os.Getenv("PLAN_CACHE_SIZE")
	if planCacheSizeStr != "" {
		if val, err := strconv.Atoi(planCacheSizeStr); err == nil && val > 0 {
			planCacheSize = val
		} else {
			analyzerLogger.Warnf("Invalid PLAN_CACHE_SIZE value '%s', using default of 10000.", planCacheSizeStr)
		}
	}

	// Configure DuckDB max open conns
	duckDBMaxOpenConns := 1
	duckDBMaxOpenConnsStr := os.Getenv("DUCKDB_MAX_OPEN_CONNS")
	if duckDBMaxOpenConnsStr != "" {
		if val, err := strconv.Atoi(duckDBMaxOpenConnsStr); err == nil && val > 0 {
			duckDBMaxOpenConns = val
		} else {
			analyzerLogger.Warnf("Invalid DUCKDB_MAX_OPEN_CONNS value '%s', using default of 1.", duckDBMaxOpenConnsStr)
		}
	}

	config := &config.Config{
		Namespace:                    namespace,
		OBTenant:                     obtenant,
		Interval:                     time.Duration(collectionIntervalSeconds) * time.Second,
		DataPath:                     dataPath,
		CompactionThreshold:          compactionThreshold,
		SqlAuditLimit:                sqlAuditLimit,
		SlowSqlThresholdMilliSeconds: slowSqlThresholdMilliSeconds,
		QueueSize:                    queueSize,
		WorkerNum:                    workerNum,
		PlanCacheSize:                planCacheSize,
		DuckDBMaxOpenConns:           duckDBMaxOpenConns,
	}

	// Initialize Stores
	if err := store.InitGlobalStores(ctx, config, collectorLogger); err != nil {
		collectorLogger.Fatalf("Failed to initialize stores: %v", err)
	}

	collector := collector.NewCollector(ctx, config, collectorLogger)
	if err := collector.Init(); err != nil {
		collectorLogger.Fatalf("Failed to initialize collector: %v", err)
	}
	go collector.Start()

	httpServer := startHttpServer(ctx, analyzerLogger)
	// Wait for a shutdown signal
	<-sigChan
	analyzerLogger.Info("Shutdown signal received, stopping...")
	// Trigger graceful shutdown
	httpServer.Stop()
	cancel()
	// Stop the collector and close stores
	collector.Stop()
	store.CloseGlobalStores()
	analyzerLogger.Info("Shutdown complete.")
}
