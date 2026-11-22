/*
Copyright (c) 2026 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package handler

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/oceanbase/ob-operator/internal/dashboard/model/response"
)

// The client runs in the same dashboard container. Keep its password out of
// Kubernetes exec request URLs, audit logs, and process arguments. Unlink the
// private file before writing it so even a crashed session leaves no pathname.
func databaseTerminalCommand(term *response.OBConnection) ([]string, *os.File, error) {
	command := []string{"mysql"}
	var options *os.File
	if term.Password != "" {
		if strings.ContainsRune(term.Password, 0) {
			return nil, nil, errors.New("invalid database password")
		}
		file, err := os.CreateTemp("", "mysql-client-*")
		if err != nil {
			return nil, nil, err
		}
		if err := os.Remove(file.Name()); err != nil {
			file.Close()
			return nil, nil, err
		}
		escape := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`, "\b", `\b`)
		if _, err := fmt.Fprintf(file, "[client]\npassword=\"%s\"\n", escape.Replace(term.Password)); err != nil {
			file.Close()
			return nil, nil, err
		}
		options = file
		command = append(command, fmt.Sprintf("--defaults-extra-file=/proc/%d/fd/%d", os.Getpid(), file.Fd()))
	}
	command = append(command, "--user="+term.User, "--host="+term.Host, "--port=2881", "-A")
	return command, options, nil
}
