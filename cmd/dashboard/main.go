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
	log.InitLogger(
		log.LoggerConfig{
			Level:      logLevel,
			Filename:   "log/oceanbase-dashboard.log",
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
