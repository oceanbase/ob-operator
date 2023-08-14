/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

func init() {
	rootCmd.AddCommand(startCmd)
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
	optStr := fmt.Sprintf("cpu_count=%s,datafile_size=%s,log_disk_size=%s", cpuCountOpt, datafileSizeOpt, logDiskSizeOpt)
	cmd := fmt.Sprintf("cd %s && %s/bin/observer --nodaemon --appname %s --cluster_id %s --zone %s --devname %s -p %d -P %d -d %s/store -l info -o config_additional_dir=%s/store/etc,%s", DefaultHomePath, DefaultHomePath, clusterName, clusterId, zoneName, DefaultDevName, DefaultSqlPort, DefaultRpcPort, DefaultHomePath, DefaultHomePath, optStr)
	return exec.Command("bash", "-c", cmd).Run()
}
