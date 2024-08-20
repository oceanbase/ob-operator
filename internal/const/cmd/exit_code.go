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

package cmd

type ExitCode int

const (
	ExitCodeOK           ExitCode = 0
	ExitCodeErr          ExitCode = 1
	ExitCodeBadArgs      ExitCode = 2
	ExitCodeNotExecuted  ExitCode = 4
	ExitCodeIgnorableErr ExitCode = 8
	ExitCodeNotSupport   ExitCode = 10
	ExitCannotExecute    ExitCode = 126
	ExitCodeNotFound     ExitCode = 127
	ExitCodeSigInt       ExitCode = 130
	ExitCodeSigKill      ExitCode = 137
	ExitCodeSegFault     ExitCode = 139
	ExitCodePipeErr      ExitCode = 141
	ExitCodeSigTerm      ExitCode = 143
)
