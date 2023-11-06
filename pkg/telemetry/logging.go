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

package telemetry

import (
	"io"
	"log"
	"os"
	"sync"
)

var lg *log.Logger
var loggerOnce sync.Once
var debugMode = os.Getenv(TelemetryDebugEnvName) == "true"

func configLogger() {
	if debugMode {
		file, err := os.OpenFile("/tmp/telemetry.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// log.Println("Failed to open log file:", err)
			lg = log.New(io.Discard, "[Telemetry] ", log.LstdFlags|log.Lshortfile)
		} else {
			lg = log.New(file, "[Telemetry] ", log.LstdFlags|log.Lshortfile)
		}
	} else {
		// if not in debug mode, discard all logs
		lg = log.New(io.Discard, "[Telemetry] ", log.LstdFlags|log.Lshortfile)
	}
}

func getLogger() *log.Logger {
	loggerOnce.Do(configLogger)
	return lg
}
