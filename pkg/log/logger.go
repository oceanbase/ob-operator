package log

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const defaultTimestampFormat = "2006-01-02T15:04:05.99999-07:00"

var textFormatter = &TextFormatter{
	TimestampFormat:        defaultTimestampFormat,
	FullTimestamp:          true,
	DisableLevelTruncation: true,
	CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
		n := 0
		filename := frame.File
		for i := len(filename) - 1; i > 0; i-- {
			if filename[i] == '/' {
				n++
				if n >= 2 {
					filename = filename[i+1:]
					break
				}
			}
		}
		name := frame.Function
		idx := strings.LastIndex(name, ".")
		return name[idx+1:], fmt.Sprintf("%s:%d", filename, frame.Line)
	},
}

type LoggerConfig struct {
	Output     io.Writer
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`
	MaxAge     int    `yaml:"maxage"`
	MaxBackups int    `yaml:"maxbackups"`
	LocalTime  bool   `yaml:"localtime"`
	Compress   bool   `yaml:"compress"`
}

func InitLogger(config LoggerConfig) *logrus.Logger {
	logger := logrus.StandardLogger()
	if config.Output == nil {
		logger.SetOutput(&lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
	} else {
		logger.SetOutput(config.Output)
	}

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(fmt.Sprintf("parse log level: %+v", err))
	}
	logger.SetLevel(level)

	logger.SetFormatter(textFormatter)
	logger.SetReportCaller(true)
	return logger
}
