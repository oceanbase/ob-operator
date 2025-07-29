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

package main

import (
	"os"

	logger "github.com/sirupsen/logrus"

	"github.com/oceanbase/ob-operator/internal/dashboard/server"
	"github.com/oceanbase/ob-operator/pkg/log"
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "log/oceanbase-dashboard.log"
	}
	log.InitLogger(
		log.LoggerConfig{
			Level:      logLevel,
			Filename:   logFile,
			MaxSize:    256,
			MaxAge:     7,
			MaxBackups: 5,
			LocalTime:  true,
			Compress:   true,
		},
	)
}

// @title OceanBase Dashboard API
// @version 1.0
// @description OceanBase Dashboard
// @BasePath /api/v1
// @SecurityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Cookie
func main() {
	httpServer := server.NewHTTPServer()
	err := httpServer.RegisterRouter()
	if err != nil {
		logger.WithError(err).Errorln("Register router failed")
		os.Exit(1)
	}
	logger.Info("Successfully registered router")
	err = httpServer.Run()
	if err != nil {
		logger.WithError(err).Errorln("Start server failed")
		os.Exit(1)
	}
}
