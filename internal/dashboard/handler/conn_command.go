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

import "github.com/oceanbase/ob-operator/internal/dashboard/model/response"

func databaseTerminalCommand(term *response.OBConnection) []string {
	// Execute mysql directly: connection values must never become shell syntax.
	// Attach each value to its option so even a value starting with '-' stays data.
	cmdArgs := []string{"mysql", "--user=" + term.User, "--host=" + term.Host, "--port=2881", "-A"}
	if term.Password != "" {
		cmdArgs = append(cmdArgs, "--password="+term.Password)
	}
	return cmdArgs
}
