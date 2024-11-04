package webshell

import (
	"fmt"
	"os/exec"
	"strconv"
)

var (
	WebShellPids = make(map[int]struct{})
)

type WebShellInfo struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Pid     int    `json:"pid"` // 为了方便后续可能进行 webshell 的关闭
}

func StartWebShell(addr string, port int, writeable bool, initCommand string, initArgs []string, timeoutMinute int) (*WebShellInfo, error) {
	startGottyCmd := []string{
		"./gotty",
		"-a", addr,
		"-p", strconv.Itoa(port),
		"--timeout", strconv.Itoa(60 * timeoutMinute),
		"--title-format", "WebShell",
	}
	if writeable {
		startGottyCmd = append(startGottyCmd, "-w")
	}
	startGottyCmd = append(startGottyCmd, initCommand)
	startGottyCmd = append(startGottyCmd, initArgs...)
	execCmd := exec.Command(startGottyCmd[0], startGottyCmd[1:]...)
	err := execCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("webshell start failed: %w", err)
	}
	newWebShellInfo := &WebShellInfo{
		Address: addr,
		Port:    port,
		Pid:     execCmd.Process.Pid,
	}
	WebShellPids[newWebShellInfo.Pid] = struct{}{}
	return &WebShellInfo{
		Address: addr,
		Port:    port,
		Pid:     execCmd.Process.Pid,
	}, nil
}

func StopWebShell(pid int) error {
	delete(WebShellPids, pid)
	killCmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
	err := killCmd.Start()
	if err != nil {
		return fmt.Errorf("webshell kill failed: %w", err)
	}
	return nil
}
