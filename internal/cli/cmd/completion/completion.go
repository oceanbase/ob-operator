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
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/internal/cli/utils"
)

const defaultBoilerPlate = `
# Copyright (c) 2024 OceanBase
# ob-operator is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#
#	http://license.coscl.org.cn/MulanPSL2
#
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
`

var completionLong = `To load completions:
Bash:
 
  $ source <(obocli completion bash)
 
  # To load completions for each session, execute once:
  # Linux:
  $ obocli completion bash > /etc/bash_completion.d/obocli
  # macOS:
  $ obocli completion bash > /usr/local/etc/bash_completion.d/obocli
 
Zsh:
 
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
 
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
 
  # To load completions for each session, execute once:
  $ obocli completion zsh > "${fpath[1]}/_obocli"
 
  # You will need to start a new shell for this setup to take effect.
 
fish:
 
  $ obocli completion fish | source
 
  # To load completions for each session, execute once:
  $ obocli completion fish > ~/.config/fish/completions/obocli.fish
 
PowerShell:
 
  PS> obocli completion powershell | Out-String | Invoke-Expression
 
  # To load completions for every new session, run:
  PS> obocli completion powershell > obocli.ps1
  # and source this file from your PowerShell profile.
`
var (
	completionShells = map[string]func(out io.Writer, boilerPlate string, cmd *cobra.Command) error{
		"bash":       runCompletionBash,
		"zsh":        runCompletionZsh,
		"fish":       runCompletionFish,
		"powershell": runCompletionPwsh,
	}
)

// NewCmd creates the instruction command for the completion of commands
func NewCmd(out io.Writer, boilerPlate string) *cobra.Command {
	logger := utils.GetDefaultLoggerInstance()
	shells := make([]string, 0, len(completionShells))
	for shell := range completionShells {
		shells = append(shells, shell)
	}
	cmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script for the specified shell (bash, zsh, fish, powershell)",
		Long:                  completionLong,
		DisableFlagsInUseLine: true,
		ValidArgs:             shells,
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			if err := RunCompletionE(out, boilerPlate, cmd, args); err != nil {
				logger.Fatalln(err)
			}
		},
	}
	return cmd
}

// RunCompletionE is the entry point for the completion command
func RunCompletionE(out io.Writer, boilerPlate string, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("shell not specified. See 'obocli completion -h' for help and examples")
	}
	if len(args) > 1 {
		return errors.New("too many arguments. Expected only the shell type. See 'obocli completion -h' for help and examples")
	}
	run, found := completionShells[args[0]]
	if !found {
		return fmt.Errorf("unsupported shell type %q", args[0])
	}
	return run(out, boilerPlate, cmd)
}

func runCompletionBash(out io.Writer, boilerPlate string, cmd *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}
	return cmd.GenBashCompletionV2(out, true)
}

func runCompletionZsh(out io.Writer, boilerPlate string, cmd *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}
	return cmd.GenZshCompletion(out)
}

func runCompletionFish(out io.Writer, boilerPlate string, cmd *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}
	return cmd.GenFishCompletion(out, true)
}

func runCompletionPwsh(out io.Writer, boilerPlate string, cmd *cobra.Command) error {
	if len(boilerPlate) == 0 {
		boilerPlate = defaultBoilerPlate
	}
	if _, err := out.Write([]byte(boilerPlate)); err != nil {
		return err
	}
	return cmd.GenPowerShellCompletionWithDesc(out)
}
