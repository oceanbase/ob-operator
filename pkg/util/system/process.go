/*
Copyright (c) 2021 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package system

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/oceanbase/ob-operator/pkg/util/shell"
	"github.com/pkg/errors"
	psutil "github.com/shirou/gopsutil/v3/process"
)

type ProcessManager struct {
}

type ProcessInfo struct {
	Pid               int32     `json:"pid"`
	Name              string    `json:"name"`              // process name
	StartCommand      string    `json:"startCommand"`      // process command line, with arguments
	Username          string    `json:"username"`          // username of the process
	Ports             []int     `json:"ports"`             // the host ports the process occupied
	CreateTime        time.Time `json:"createTime"`        // process create time
	ElapsedTimeMillis int64     `json:"elapsedTimeMillis"` // elapsed time since process created
}

const GetProcessTcpListenCommand = "netstat -tunlp | { grep '%d/' || true; }"

func (p ProcessManager) listProcesses() ([]*psutil.Process, error) {
	processes, err := psutil.Processes()
	if err != nil {
		return nil, errors.Errorf("failed to list processes: %s", err)
	}
	return processes, nil
}

func (p ProcessManager) findProcessByName(name string) ([]*psutil.Process, error) {
	processes, err := p.listProcesses()
	if err != nil {
		return nil, errors.Wrapf(err, "find process by name %s", name)
	}
	var result []*psutil.Process
	for _, proc := range processes {
		if procName, err := proc.Name(); err == nil && procName == name {
			result = append(result, proc)
		}
	}
	return result, nil
}

func (p ProcessManager) FindProcessInfoByName(name string) ([]*ProcessInfo, error) {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return nil, errors.Wrapf(err, "find process info by name %s", name)
	}
	result := make([]*ProcessInfo, 0)
	for _, proc := range processes {
		result = append(result, processToProcessInfo(proc))
	}
	return result, nil
}

func (p ProcessManager) ProcessExists(name string) (bool, error) {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return false, errors.Wrapf(err, "check process exists by name %s", name)
	}
	return len(processes) > 0, nil
}

func (p ProcessManager) ProcessIsRunningByName(name string) bool {
	processes, err := p.findProcessByName(name)
	if err != nil {
		log.Printf("find process info by name %s %s", name, err)
		return false
	}
	var status bool
	for _, proc := range processes {
		procName, _ := proc.Name()
		procStatus, _ := proc.Status()
		log.Println(procName, proc.Pid, procStatus)
		for _, ps := range procStatus {
			if ps == "zombie" {
				return false
			}
		}
		status, err = proc.IsRunning()
		if err != nil {
			return false
		}
	}
	return status
}

func (p ProcessManager) terminateProcess(proc *psutil.Process) error {
	err := proc.Terminate()
	if err != nil {
		return errors.Errorf("failed to terminate process with pid %d: %s", proc.Pid, err)
	}
	return nil
}

func (p ProcessManager) TerminateProcessByName(name string) error {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return errors.Wrapf(err, "terminate process by name %s", name)
	}
	for _, proc := range processes {
		err = p.terminateProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "terminate process by name %s", name)
		}
	}
	return nil
}

func (p ProcessManager) killProcess(proc *psutil.Process) error {
	err := proc.Kill()
	if err != nil {
		return errors.Errorf("failed to kill process with pid %d, %s", proc.Pid, err)
	}
	return nil
}

func (p ProcessManager) KillProcessByName(name string) error {
	processes, err := p.findProcessByName(name)
	if err != nil {
		return errors.Wrapf(err, "kill process by name %s", name)
	}
	for _, proc := range processes {
		err = p.killProcess(proc)
		if err != nil {
			return errors.Wrapf(err, "kill process by name %s", name)
		}
	}
	return nil
}

func processToProcessInfo(p *psutil.Process) *ProcessInfo {
	pi := ProcessInfo{
		Ports: []int{},
	}
	pi.Pid = p.Pid
	if createTimeMillis, err := p.CreateTime(); err == nil {
		createTime := time.Unix(createTimeMillis/1000, (createTimeMillis%1000)*1000)
		pi.CreateTime = createTime
		elapsedTime := time.Now().Sub(createTime)
		pi.ElapsedTimeMillis = int64(elapsedTime / time.Millisecond)
	}
	if name, err := p.Name(); err == nil {
		pi.Name = name
	}
	if cmdline, err := p.Cmdline(); err == nil {
		pi.StartCommand = cmdline
	}
	if username, err := p.Username(); err == nil {
		pi.Username = username
	}
	if ports, err := getProcessOccupiedPorts(p.Pid); err == nil {
		pi.Ports = ports
	}
	return &pi
}

func getProcessOccupiedPorts(pid int32) ([]int, error) {
	cmd := fmt.Sprintf(GetProcessTcpListenCommand, pid)
	executeResult, err := shell.NewCommand(cmd).WithUser(shell.RootUser).Execute()
	if err != nil {
		return nil, errors.Wrap(err, "get process occupied ports")
	}
	output := strings.TrimSpace(executeResult.Output)
	if output == "" {
		return []int{}, nil
	}
	occupiedPorts := make([]int, 0)
	tcpLines := strings.Split(output, "\n")
	for _, tcpLine := range tcpLines {
		port, ok := parseNetstatLine(tcpLine)
		if !ok {
			return nil, errors.Errorf("failed to get process occupied ports, invalid output line: %s", tcpLine)
		}
		occupiedPorts = append(occupiedPorts, port)
	}
	// One occupied port can corresponds to multiple netstat lines, so remove duplicate ports.
	return removeDuplicate(occupiedPorts), nil
}

// Example:
// Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
// tcp        0      0 0.0.0.0:62888           0.0.0.0:*               LISTEN      104668/pos_proxy
// tcp6       0      0 :::62888                :::*                    LISTEN      104668/pos_proxy
func parseNetstatLine(line string) (int, bool) {
	fields := strings.Fields(line)
	if len(fields) != 7 {
		return 0, false
	}
	field := fields[3]
	if len(field) == 0 || !strings.Contains(field, ":") {
		return 0, false
	}
	i := strings.LastIndex(field, ":")
	if i == -1 {
		return 0, false
	}
	portStr := field[i+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, false
	}
	return port, true
}

func removeDuplicate(ports []int) []int {
	if len(ports) == 0 {
		return []int{}
	}
	result := make([]int, 0)
	var seen = make(map[int]bool, len(ports))
	for _, port := range ports {
		if _, exists := seen[port]; !exists {
			result = append(result, port)
			seen[port] = true
		}
	}
	return result
}
