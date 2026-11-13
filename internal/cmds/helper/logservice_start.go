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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

const (
	DefaultLogServiceStoreMountPath = "/home/admin/oblogservice/store"
	DefaultLogServiceLogMountPath   = "/home/admin/oblogservice/log"
	DefaultLogServiceBinPath        = "/home/admin/oblogservice/bin/oblogservice"
	DefaultLogServiceEtcPath        = "/home/admin/oblogservice/etc"
	DefaultLogServiceRpcPort        = "50051"
	DefaultLogServiceHttpPort       = "50052"
	logServiceStartTimeoutEnv       = "LOGSERVICE_START_TIMEOUT_SECONDS"
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
	startTimeout := getLogServiceStartTimeout()
	deadline, err := startAndWaitForLauncher(cmd, startTimeout)
	if err != nil {
		return err
	}
	return monitorDaemon(deadline, startTimeout)
}

func startAndWaitForLauncher(cmd *exec.Cmd, startTimeout time.Duration) (time.Time, error) {
	if err := cmd.Start(); err != nil {
		return time.Time{}, errors.Wrap(err, "start oblogservice launcher")
	}
	deadline := time.Now().Add(startTimeout)
	waitResult := make(chan error, 1)
	go func() {
		waitResult <- cmd.Wait()
	}()

	timer := time.NewTimer(startTimeout)
	defer timer.Stop()
	select {
	case err := <-waitResult:
		if err != nil {
			return time.Time{}, errors.Wrap(err, "oblogservice launcher exited with an error")
		}
		return deadline, nil
	case <-timer.C:
		// Prefer a concurrently completed Wait result over reporting a timeout.
		select {
		case err := <-waitResult:
			if err != nil {
				return time.Time{}, errors.Wrap(err, "oblogservice launcher exited with an error")
			}
			return deadline, nil
		default:
		}
		killErr := cmd.Process.Kill()
		<-waitResult // Reap the launcher after Kill so no process or goroutine leaks.
		if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			return time.Time{}, errors.Wrap(killErr, "kill timed-out oblogservice launcher")
		}
		return time.Time{}, errors.Errorf("oblogservice launcher failed to exit within %s", startTimeout)
	}
}

// monitorDaemon polls for the oblogservice daemon process and exits if it dies.
// It forwards SIGTERM/SIGINT and bounds shutdown with an explicit kill fallback.
func monitorDaemon(deadline time.Time, startTimeout time.Duration) error {
	daemonPid, err := waitForDaemon(deadline, startTimeout)
	if err != nil {
		return err
	}
	log.Printf("oblogservice daemon running with pid %d", daemonPid)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sig)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case s := <-sig:
			return stopLogServiceDaemon(daemonPid, s.(syscall.Signal), 20*time.Second)
		case <-ticker.C:
			running, err := logServiceProcessRunning("/proc", daemonPid)
			if err != nil {
				return errors.Wrap(err, "inspect oblogservice daemon")
			}
			if !running {
				return errors.New("oblogservice daemon exited unexpectedly")
			}
		}
	}
}

func waitForDaemon(deadline time.Time, startTimeout time.Duration) (int, error) {
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 0, errors.Errorf("oblogservice daemon failed to start within %s", startTimeout)
	}
	startTimer := time.NewTimer(remaining)
	pollTicker := time.NewTicker(time.Second)
	defer startTimer.Stop()
	defer pollTicker.Stop()

	for {
		daemonPid, err := findLogServiceDaemon("/proc")
		if err != nil {
			return 0, errors.Wrap(err, "inspect oblogservice processes")
		}
		if daemonPid > 0 {
			return daemonPid, nil
		}
		select {
		case <-pollTicker.C:
		case <-startTimer.C:
			return 0, errors.Errorf("oblogservice daemon failed to start within %s", startTimeout)
		}
	}
}

func getLogServiceStartTimeout() time.Duration {
	rawTimeout := os.Getenv(logServiceStartTimeoutEnv)
	if rawTimeout == "" {
		return time.Duration(oceanbaseconst.LogServiceStartTimeoutSeconds) * time.Second
	}
	seconds, err := strconv.ParseInt(rawTimeout, 10, 32)
	if err != nil || seconds <= 0 {
		log.Printf("Invalid %s value %q, using default %ds", logServiceStartTimeoutEnv, rawTimeout, oceanbaseconst.LogServiceStartTimeoutSeconds)
		return time.Duration(oceanbaseconst.LogServiceStartTimeoutSeconds) * time.Second
	}
	return time.Duration(seconds) * time.Second
}

// Inspect procfs directly: minimal runtime images need not contain pgrep.
// A failed inspection is an error, not evidence that the daemon is absent.
func findLogServiceDaemon(procRoot string) (int, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return 0, err
	}
	oldest := 0
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		running, err := logServiceProcessRunning(procRoot, pid)
		if err != nil {
			return 0, err
		}
		if running && (oldest == 0 || pid < oldest) {
			oldest = pid
		}
	}
	return oldest, nil
}

func logServiceProcessRunning(procRoot string, pid int) (bool, error) {
	data, err := os.ReadFile(filepath.Join(procRoot, strconv.Itoa(pid), "stat"))
	if os.IsNotExist(err) {
		return false, nil
	} // Process exited during enumeration.
	if err != nil {
		return false, err
	}
	stat := string(data)
	open, close := strings.IndexByte(stat, '('), strings.LastIndexByte(stat, ')')
	if open < 0 || close <= open {
		return false, errors.New("invalid process stat")
	}
	fields := strings.Fields(stat[close+1:])
	if len(fields) == 0 {
		return false, errors.New("missing process state in stat")
	}
	return stat[open+1:close] == "oblogservice" && fields[0] != "Z" && fields[0] != "X", nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// Bound shutdown below Kubernetes' default 30-second termination grace period.
// Some LogService builds handle TERM without exiting; report the forced fallback.
func stopLogServiceDaemon(pid int, sig syscall.Signal, grace time.Duration) error {
	log.Printf("Stopping LogService pid %d with %v", pid, sig)
	if err := syscall.Kill(pid, sig); err != nil && err != syscall.ESRCH {
		return err
	}
	deadline := time.Now().Add(grace)
	forced := false
	for {
		running, err := logServiceProcessRunning("/proc", pid)
		if err != nil {
			return err
		}
		if !running {
			return nil
		}
		if time.Now().After(deadline) {
			if forced {
				return errors.New("LogService remains alive after SIGKILL")
			}
			log.Printf("WARNING: LogService pid %d did not stop within %s; forcing SIGKILL", pid, grace)
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
				return err
			}
			forced = true
			deadline = time.Now().Add(3 * time.Second)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
