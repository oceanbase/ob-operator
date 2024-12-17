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
package version

import (
	"os"
	"runtime"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// Injected by build script
var (
	OS         = "unknown"
	Arch       = "unknown"
	Version    = "unknown"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

// defaultVersionTemplate is the default template for displaying version information.
const defaultVersionTemplate = `
OceanBase Operator Cli:
 Version:    {{.Version}}
 OS/Arch:	   {{.OS}}/{{.Arch}}
 Go Version: {{.GoVersion}}
 Git Commit: {{.CommitHash}}
 Build:      {{.BuildTime}}
`

// VersionInfo stores the version information.
type VersionInfo struct {
	Version    string
	OS         string
	Arch       string
	GoVersion  string
	CommitHash string
	BuildTime  string
}

// NewCmd returns the version command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of OceanBase Cli",
		RunE: func(cmd *cobra.Command, args []string) error {
			return printCliInfo()
		},
	}
	return cmd
}

func reformatTime(buildTime string) string {
	t, err := time.Parse("20060102150405", buildTime)
	if err != nil {
		return buildTime
	}
	return t.Format("2006-01-02 15:04:05")
}

// printCliInfo prints the version information of OceanBase Cli.
func printCliInfo() error {
	t := tabwriter.NewWriter(os.Stdout, 10, 3, 1, ' ', 0)
	tmpl, err := template.New("version").Parse(defaultVersionTemplate)
	data := &VersionInfo{
		Version:    Version,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GoVersion:  runtime.Version(),
		CommitHash: CommitHash,
		BuildTime:  reformatTime(BuildTime),
	}
	if err != nil {
		return err
	}
	if err := tmpl.Execute(t, data); err != nil {
		return err
	}
	if _, err := t.Write([]byte("\n")); err != nil {
		return err
	}
	return t.Flush()
}
