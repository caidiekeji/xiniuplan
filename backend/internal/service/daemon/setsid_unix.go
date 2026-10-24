//go:build !windows

package daemon

import (
	"os"
	"os/exec"
	"syscall"
)

// setDetach 脱离面板进程组（Setsid）：面板重启不殃及被守护进程
func setDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// pidSignal 探活信号（unix: signal 0）
func pidSignal() os.Signal {
	return syscall.Signal(0)
}
