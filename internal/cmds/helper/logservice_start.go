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
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	DefaultLogServiceStoreMountPath = "/home/admin/oblogservice/store"
	DefaultLogServiceLogMountPath   = "/home/admin/oblogservice/log"
	DefaultLogServiceBinPath        = "/home/admin/oblogservice/bin/oblogservice"
	DefaultLogServiceEtcPath        = "/home/admin/oblogservice/etc"
	DefaultLogServiceRpcPort        = "50051"
	DefaultLogServiceHttpPort       = "50052"
)

var logserviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start OBLogService",
	Run: func(cmd *cobra.Command, _ []string) {
		if err := prepareLogServiceDir(); err != nil {
			cmd.PrintErrf("Prepare logservice dir failed: %v\n", err)
			os.Exit(1)
		}
		if err := startLogService(); err != nil {
			cmd.PrintErrf("Start logservice failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	logserviceCmd.AddCommand(logserviceStartCmd)
}

func prepareLogServiceDir() error {
	storeMountPath := getEnvOrDefault("STORE_MOUNT_PATH", DefaultLogServiceStoreMountPath)
	logMountPath := getEnvOrDefault("LOG_MOUNT_PATH", DefaultLogServiceLogMountPath)
	etcPath := DefaultLogServiceEtcPath
	persistEtcDir := storeMountPath + "/etc"
	localStorageDir := storeMountPath + "/data"

	cmd := fmt.Sprintf(
		"mkdir -p %s %s %s && (cp %s/*.yml %s/ 2>/dev/null || true) && rm -rf %s && ln -s %s %s",
		localStorageDir, persistEtcDir, logMountPath,
		etcPath, persistEtcDir,
		etcPath, persistEtcDir, etcPath,
	)
	log.Println("Prepare dirs:", cmd)
	return exec.Command("bash", "-c", cmd).Run()
}

func startLogService() error {
	storeMountPath := getEnvOrDefault("STORE_MOUNT_PATH", DefaultLogServiceStoreMountPath)
	persistEtcDir := storeMountPath + "/etc"
	configFile := persistEtcDir + "/config.toml"

	_, err := os.Stat(configFile)
	if err == nil {
		log.Println("Found config.toml, starting logservice without -g (reuse config)")
		return startLogServiceWithConfig()
	}
	if os.IsNotExist(err) {
		log.Println("config.toml not found, starting logservice with -g (first boot)")
		return startLogServiceWithParam()
	}
	return errors.Wrap(err, "failed to check config file")
}

func startLogServiceWithConfig() error {
	cmd := exec.Command(DefaultLogServiceBinPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return runAndSupervise(cmd)
}

func startLogServiceWithParam() error {
	storeMountPath := getEnvOrDefault("STORE_MOUNT_PATH", DefaultLogServiceStoreMountPath)
	localStorageDir := getEnvOrDefault("LOCAL_STORAGE_DIR", storeMountPath+"/data")

	clusterId := os.Getenv("CLUSTER_ID")
	if clusterId == "" {
		return errors.New("CLUSTER_ID is required")
	}
	localIP := os.Getenv("LOCAL_IP")
	if localIP == "" {
		return errors.New("LOCAL_IP is required")
	}
	rpcPort := getEnvOrDefault("RPC_PORT", DefaultLogServiceRpcPort)
	httpPort := getEnvOrDefault("HTTP_PORT", DefaultLogServiceHttpPort)

	params := []string{
		fmt.Sprintf("cluster_id=%s", clusterId),
		fmt.Sprintf("local_ip=%s", localIP),
		fmt.Sprintf("rpc_port=%s", rpcPort),
		fmt.Sprintf("http_port=%s", httpPort),
		fmt.Sprintf("local_storage_dir=%s", localStorageDir),
	}
	if logDiskSize := os.Getenv("LOG_DISK_SIZE"); logDiskSize != "" {
		params = append(params, fmt.Sprintf("log_disk_size=%s", logDiskSize))
	}
	if extra := os.Getenv("EXTRA_PARAMETERS"); extra != "" {
		params = append(params, extra)
	}

	paramStr := strings.Join(params, ",")
	log.Printf("Start command: %s -g \"%s\"\n", DefaultLogServiceBinPath, paramStr)
	cmd := exec.Command(DefaultLogServiceBinPath, "-g", paramStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return runAndSupervise(cmd)
}

// runAndSupervise starts the command (which daemonizes itself), then monitors
// the daemon process. It forwards SIGTERM/SIGINT to the daemon and exits when
// the daemon exits, so kubelet can detect the failure and restart the container.
func runAndSupervise(cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start oblogservice")
	}
	// Wait for the parent process to exit (oblogservice daemonizes via fork)
	_ = cmd.Wait()

	// The daemon has forked; find and monitor the actual daemon process
	return monitorDaemon()
}

// monitorDaemon polls for the oblogservice daemon process and exits if it dies.
// It also forwards SIGTERM/SIGINT to the daemon PID for graceful shutdown.
func monitorDaemon() error {
	// Wait for daemon to appear (up to 30s for slow startup)
	var daemonPid int
	for range 30 {
		time.Sleep(time.Second)
		pid := getDaemonPid()
		if pid > 0 {
			daemonPid = pid
			break
		}
	}
	if daemonPid == 0 {
		return errors.New("oblogservice daemon failed to start within 30s")
	}
	log.Printf("oblogservice daemon running with pid %d", daemonPid)

	// Forward termination signals to the daemon
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("Received signal %v, forwarding to daemon pid %d", s, daemonPid)
		_ = syscall.Kill(daemonPid, s.(syscall.Signal))
	}()

	for {
		time.Sleep(5 * time.Second)
		if getDaemonPid() == 0 {
			return errors.New("oblogservice daemon exited unexpectedly")
		}
	}
}

func getDaemonPid() int {
	out, err := exec.Command("pgrep", "-x", "oblogservice").Output()
	if err != nil {
		return 0
	}
	pidStr := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	if pidStr == "" {
		return 0
	}
	pid := 0
	fmt.Sscanf(pidStr, "%d", &pid)
	return pid
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
