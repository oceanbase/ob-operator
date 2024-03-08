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
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/oceanbase/ob-operator/pkg/helper"
)

const (
	DefaultLogDiskSize   = "40G"
	DefaultDatafileSize  = "40G"
	DefaultCpuCount      = "16"
	DefaultDevName       = "eth0"
	DefaultHomePath      = "/home/admin/oceanbase"
	DefaultDataFilePath  = "/home/admin/data-file"
	DefaultClogPath      = "/home/admin/data-log"
	DefaultLogPath       = "/home/admin/log"
	BackupConfigFileName = "observer.conf.bin"
	ConfigFileName       = "observer.config.bin"
)

const (
	DefaultSqlPort = 2881
	DefaultRpcPort = 2882
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start OceanBase",
	Run: func(cmd *cobra.Command, args []string) {
		err := prepareDir()
		if err != nil {
			fmt.Printf("Prepare observer dir failed, %v \n", err)
			os.Exit(1)
		}
		err = startOBServer()
		if err != nil {
			fmt.Printf("Start observer failed, %v \n", err)
			os.Exit(1)
		}
	},
}

var MinStandaloneVersion *helper.OceanBaseVersion
var MinServiceVersion *helper.OceanBaseVersion

func init() {
	rootCmd.AddCommand(startCmd)
	MinStandaloneVersion, _ = helper.ParseOceanBaseVersion("4.2.0.0")
	MinServiceVersion, _ = helper.ParseOceanBaseVersion("4.2.3.0")
}

func prepareDir() error {
	makeLogDirCmd := fmt.Sprintf("mkdir -p %s/log && ln -sf %s/log %s/log", DefaultLogPath, DefaultLogPath, DefaultHomePath)
	makeStoreDirCmd := fmt.Sprintf("mkdir -p %s/store", DefaultHomePath)
	makeCLogDirCmd := fmt.Sprintf("mkdir -p %s/clog && ln -sf %s/clog %s/store/clog", DefaultClogPath, DefaultClogPath, DefaultHomePath)
	makeILogDirCmd := fmt.Sprintf("mkdir -p %s/ilog && ln -sf %s/ilog %s/store/ilog", DefaultClogPath, DefaultClogPath, DefaultHomePath)
	makeSLogDirCmd := fmt.Sprintf("mkdir -p %s/slog && ln -sf %s/slog %s/store/slog", DefaultDataFilePath, DefaultDataFilePath, DefaultHomePath)
	makeEtcDirCmd := fmt.Sprintf("mkdir -p %s/etc && ln -sf %s/etc %s/store/etc", DefaultDataFilePath, DefaultDataFilePath, DefaultHomePath)
	makeSortDirCmd := fmt.Sprintf("mkdir -p %s/sort_dir && ln -sf %s/sort_dir %s/store/sort_dir", DefaultDataFilePath, DefaultDataFilePath, DefaultHomePath)
	makeSstableDirCmd := fmt.Sprintf("mkdir -p %s/sstable && ln -sf %s/sstable %s/store/sstable", DefaultDataFilePath, DefaultDataFilePath, DefaultHomePath)
	cmd := fmt.Sprintf("%s && %s && %s && %s && %s && %s && %s && %s", makeLogDirCmd, makeStoreDirCmd, makeCLogDirCmd, makeILogDirCmd, makeSLogDirCmd, makeEtcDirCmd, makeSortDirCmd, makeSstableDirCmd)
	return exec.Command("bash", "-c", cmd).Run()
}

func startOBServer() error {
	configFile := fmt.Sprintf("%s/store/etc/%s", DefaultHomePath, BackupConfigFileName)
	_, err := os.Stat(configFile)
	if err == nil {
		fmt.Println("Found backup config file, start without parameter")
		return startOBServerWithConfig()
	}
	if os.IsNotExist(err) {
		fmt.Println("Backup config file not found, start with parameter")
		return startOBServerWithParam()
	}
	return errors.Wrap(err, "Failed to check config file")
}

func startOBServerWithConfig() error {
	cmd := fmt.Sprintf("cp %s/store/etc/%s %s/etc/%s && %s/bin/observer --nodaemon", DefaultHomePath, BackupConfigFileName, DefaultHomePath, ConfigFileName, DefaultHomePath)
	return exec.Command("bash", "-c", cmd).Run()
}

func startOBServerWithParam() error {
	logDiskSizeOpt, found := os.LookupEnv("LOG_DISK_SIZE")
	if !found {
		logDiskSizeOpt = DefaultLogDiskSize
	}
	datafileSizeOpt, found := os.LookupEnv("DATAFILE_SIZE")
	if !found {
		datafileSizeOpt = DefaultDatafileSize
	}
	cpuCountOpt, found := os.LookupEnv("CPU_COUNT")
	if !found {
		cpuCountOpt = DefaultCpuCount
	}
	clusterName, found := os.LookupEnv("CLUSTER_NAME")
	if !found {
		return errors.New("cluster name is required")
	}
	clusterId, found := os.LookupEnv("CLUSTER_ID")
	if !found {
		return errors.New("cluster id is required")
	}
	zoneName, found := os.LookupEnv("ZONE_NAME")
	if !found {
		return errors.New("zone name is required")
	}
	extraOptStr, found := os.LookupEnv("EXTRA_OPTION")
	if !found {
		extraOptStr = ""
	}
	standalone, found := os.LookupEnv("STANDALONE")
	if found {
		standalone = "true"
	}
	optStr := fmt.Sprintf("cpu_count=%s,datafile_size=%s,log_disk_size=%s,enable_syslog_recycle=true,max_syslog_file_count=4", cpuCountOpt, datafileSizeOpt, logDiskSizeOpt)
	if extraOptStr != "" {
		optStr = fmt.Sprintf("%s,%s", optStr, extraOptStr)
	}
	ver, err := helper.GetCurrentVersion(DefaultHomePath)
	if err != nil {
		return errors.Wrap(err, "Failed to get current version")
	}
	obv, err := helper.ParseOceanBaseVersion(ver)
	if err != nil {
		return errors.Wrap(err, "Failed to parse current version")
	}
	var cmd string
	svcIP := os.Getenv("SVC_IP")
	if standalone != "" && obv.Cmp(MinStandaloneVersion) >= 0 {
		cmd = fmt.Sprintf("cd %s && %s/bin/observer --nodaemon --appname %s --cluster_id %s --zone %s --devname lo -p %d -P %d -d %s/store -l info -o config_additional_dir=%s/store/etc,%s", DefaultHomePath, DefaultHomePath, clusterName, clusterId, zoneName, DefaultSqlPort, DefaultRpcPort, DefaultHomePath, DefaultHomePath, optStr)
	} else if svcIP != "" {
		cmd = fmt.Sprintf("cd %s && %s/bin/observer --nodaemon --appname %s --cluster_id %s --zone %s -I %s -p %d -P %d -d %s/store -l info -o config_additional_dir=%s/store/etc,%s", DefaultHomePath, DefaultHomePath, clusterName, clusterId, zoneName, svcIP, DefaultSqlPort, DefaultRpcPort, DefaultHomePath, DefaultHomePath, optStr)
	} else {
		cmd = fmt.Sprintf("cd %s && %s/bin/observer --nodaemon --appname %s --cluster_id %s --zone %s -i %s -p %d -P %d -d %s/store -l info -o config_additional_dir=%s/store/etc,%s", DefaultHomePath, DefaultHomePath, clusterName, clusterId, zoneName, DefaultDevName, DefaultSqlPort, DefaultRpcPort, DefaultHomePath, DefaultHomePath, optStr)
	}
	fmt.Println("Start commands: ", cmd)
	return exec.Command("bash", "-c", cmd).Run()
}
