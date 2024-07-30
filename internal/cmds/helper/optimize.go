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
package helper

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	cmdconst "github.com/oceanbase/ob-operator/internal/const/cmd"
	"github.com/oceanbase/ob-operator/pkg/helper/model"
)

const (
	OptimizedParameterConfigFile = "/home/admin/oceanbase/etc/default_parameter.json"
	OptimizedVariableConfigFile  = "/home/admin/oceanbase/etc/default_system_variable.json"
)

// optimizeCmd represents the optimize command
var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Optimize parameters and variables",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.PrintErr("No enough args, scope and scenario are needed")
			os.Exit(int(cmdconst.ExitCodeBadArgs))
		}
		if len(args) > 2 {
			cmd.PrintErr("Too many args")
			os.Exit(int(cmdconst.ExitCodeBadArgs))
		}
		scope := args[0]
		scenario := args[1]
		ignorableErr := make([]error, 0)
		parameters, err := readOptimizedParams(scope, scenario)
		if err != nil {
			cmd.PrintErrf("Failed to read optimized parameters, %v", err)
			ignorableErr = append(ignorableErr, err)
		}
		variables, err := readOptimizedVariables(scope, scenario)
		if err != nil {
			cmd.PrintErrf("Failed to read optimized variables, %v", err)
			ignorableErr = append(ignorableErr, err)
		}
		optimizationResult := &model.OptimizationResponse{
			Parameters: parameters,
			Variables:  variables,
		}
		resultBytes, err := json.Marshal(optimizationResult)
		resultStr := string(resultBytes)
		if err != nil {
			os.Exit(int(cmdconst.ExitCodeErr))
		}
		cmd.Print(resultStr)
		if len(ignorableErr) > 0 {
			os.Exit(int(cmdconst.ExitCodeIgnorableErr))
		}
	},
}

func init() {
	rootCmd.AddCommand(optimizeCmd)
}

func readOptimizedParams(scope, scenario string) ([]model.OBConfig, error) {
	obconfigs := make([]model.OBConfig, 0)
	content, err := os.ReadFile(OptimizedParameterConfigFile)
	if err != nil {
		return obconfigs, err
	}
	parameters := make([]model.OptimizedParameters, 0)
	err = json.Unmarshal(content, &parameters)
	if err != nil {
		return obconfigs, err
	}
	for _, parameter := range parameters {
		if parameter.Scenario != scenario {
			continue
		}
		switch scope {
		case "cluster":
			obconfigs = parameter.Parameters.Cluster
		case "tenant":
			obconfigs = parameter.Parameters.Tenant
		}

	}
	return obconfigs, nil
}

func readOptimizedVariables(scope, scenario string) ([]model.OBConfig, error) {
	obconfigs := make([]model.OBConfig, 0)
	content, err := os.ReadFile(OptimizedVariableConfigFile)
	if err != nil {
		return obconfigs, err
	}
	variables := make([]model.OptimizedVariables, 0)
	err = json.Unmarshal(content, &variables)
	if err != nil {
		return obconfigs, err
	}
	for _, variable := range variables {
		if variable.Scenario != scenario {
			continue
		}
		switch scope {
		case "cluster":
			obconfigs = variable.Variables.Cluster
		case "tenant":
			obconfigs = variable.Variables.Tenant
		}

	}
	return obconfigs, nil
}
