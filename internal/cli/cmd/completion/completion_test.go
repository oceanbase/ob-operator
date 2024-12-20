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
package completion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompletions(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name: "bash",
			args: []string{"bash"},
		},
		{
			name: "zsh",
			args: []string{"zsh"},
		},
		{
			name: "fish",
			args: []string{"fish"},
		},
		{
			name: "powershell",
			args: []string{"powershell"},
		},
		{
			name:          "no args",
			args:          []string{},
			expectedError: "shell not specified. See 'okctl completion -h' for help and examples",
		},
		{
			name:          "too many args",
			args:          []string{"bash", "zsh"},
			expectedError: "too many arguments. Expected only the shell type. See 'okctl completion -h' for help and examples",
		},
		{
			name:          "unsupported shell",
			args:          []string{"foo"},
			expectedError: (`unsupported shell type "foo"`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			parentCmd := &cobra.Command{
				Use: "okctl",
			}
			out := new(bytes.Buffer)
			cmd := NewCmd(out, defaultBoilerPlate)
			parentCmd.AddCommand(cmd)
			err := RunCompletionE(out, defaultBoilerPlate, cmd, tc.args)
			if tc.expectedError == "" {
				if err != nil {
					tt.Fatalf("Unexpected error: %v", err)
				}
				if out.Len() == 0 {
					tt.Fatalf("Output was not written")
				}
				if !strings.Contains(out.String(), defaultBoilerPlate) {
					tt.Fatalf("Output does not contain boilerplate:\n%s", out.String())
				}
			} else {
				if err == nil {
					tt.Fatalf("An error was expected but no error was returned")
				}
				if err.Error() != tc.expectedError {
					tt.Fatalf("unexpected error: %v\n expected: %v\n", err, tc.expectedError)
				}
			}
		})
	}
}
