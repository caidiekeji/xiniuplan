//go:build windows

package daemon

import (
	"os"
	"os/exec"
)

// setDetach Windows 下无 Setsid，空实现（守护功能仅面向 Linux 部署）
func setDetach(cmd *exec.Cmd) {}

// pidSignal Windows 探活占位（Windows 部署非目标场景）

type winSignal struct{}

func (winSignal) Signal()        {}
func (winSignal) String() string { return "" }

func pidSignal() os.Signal { return winSignal{} }
