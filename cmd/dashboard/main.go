package main

import (
	"os"

	"github.com/oceanbase/ob-operator/internal/dashboard/server"
	"github.com/oceanbase/ob-operator/pkg/log"
	logger "github.com/sirupsen/logrus"
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
	HTTPServer := server.NewHTTPServer()
	err := HTTPServer.RegisterRouter()
	if err != nil {
		logger.WithError(err).Errorln("Register router failed")
		os.Exit(1)
	}
	logger.Info("Successfully registered router")
	err = HTTPServer.Run()
	if err != nil {
		logger.WithError(err).Errorln("Start server failed")
		os.Exit(1)
	}
}
