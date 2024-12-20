/*
Copyright (c) 2024 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:

	http://license.coscl.org.cn/MulanPSL2

THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/
package utils

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
)

var fLog *log.Logger
var tbLog *log.Logger
var tbw = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

// BinaryName injected by ldflags
var BinaryName string

func init() {
	logHead := fmt.Sprintf("[%s]: ", BinaryName)
	fLog = log.New(os.Stdout, logHead, 0)
	tbLog = log.New(tbw, "", 0)
}

// GetDefaultLoggerInstance return a default logger instance
func GetDefaultLoggerInstance() *log.Logger {
	return fLog
}

// GetTableLoggerInstance return a table logger instance
func GetTableLoggerInstance() (*tabwriter.Writer, *log.Logger) {
	return tbw, tbLog
}
