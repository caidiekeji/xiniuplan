package monitor

import (
	"fmt"
	"os"
	"sort"
	"time"

	"cloudpivot/internal/dto"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ListProcesses 获取进程列表，按 CPU/内存排序，返回 Top N
func (s *Service) ListProcesses(topN int, sortBy string) ([]dto.ProcessInfo, error) {
	if topN <= 0 || topN > 200 {
		topN = 20
	}
	if sortBy != "cpu" && sortBy != "memory" {
		sortBy = "cpu"
	}

	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("get processes: %w", err)
	}

	infos := make([]dto.ProcessInfo, 0, len(procs))
	for _, p := range procs {
		info := dto.ProcessInfo{PID: p.Pid}

		if name, err := p.Name(); err == nil {
			info.Name = name
		}

		if cpu, err := p.CPUPercent(); err == nil {
			info.CPUPercent = cpu
		}

		if memPercent, err := p.MemoryPercent(); err == nil {
			info.MemoryPercent = float64(memPercent)
		}

		if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
			info.MemoryRSS = memInfo.RSS
		}

		if status, err := p.Status(); err == nil && len(status) > 0 {
			info.Status = status[0]
		}

		if createTime, err := p.CreateTime(); err == nil {
			info.CreateTime = createTime / 1000 // 毫秒转秒
		}

		infos = append(infos, info)
	}

	// 排序
	sort.Slice(infos, func(i, j int) bool {
		if sortBy == "cpu" {
			return infos[i].CPUPercent > infos[j].CPUPercent
		}
		return infos[i].MemoryPercent > infos[j].MemoryPercent
	})

	// 取 Top N
	if len(infos) > topN {
		infos = infos[:topN]
	}

	return infos, nil
}

// KillProcess 终止进程（M-002），支持 SIGTERM（signal=15）与 SIGKILL（signal=9）
// 保护性拦截：PID 1（init）、面板自身进程不允许终止
func (s *Service) KillProcess(pid int32, signal int32) error {
	if pid <= 1 {
		return fmt.Errorf("不允许终止 init 进程 (PID %d)", pid)
	}
	if int(pid) == os.Getpid() {
		return fmt.Errorf("不允许终止面板自身进程 (PID %d)", pid)
	}
	if signal != 9 && signal != 15 {
		signal = 15 // 默认 SIGTERM
	}

	proc, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("进程 %d 不存在", pid)
	}
	name, _ := proc.Name()
	if signal == 9 {
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("强制结束进程 %d (%s) 失败: %w", pid, name, err)
		}
		return nil
	}
	if err := proc.Terminate(); err != nil {
		return fmt.Errorf("终止进程 %d (%s) 失败: %w", pid, name, err)
	}

	// SIGTERM 后短暂等待，若仍存活提示可用 SIGKILL
	time.Sleep(1500 * time.Millisecond)
	if running, _ := proc.IsRunning(); running {
		return fmt.Errorf("已发送 SIGTERM，进程 %d (%s) 仍未退出，可尝试强制结束 (signal=9)", pid, name)
	}
	return nil
}

// GetDiskAnalysis 各挂载点使用量
func (s *Service) GetDiskAnalysis() ([]dto.DiskUsage, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("get disk partitions: %w", err)
	}

	usages := make([]dto.DiskUsage, 0, len(parts))
	for _, part := range parts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		usages = append(usages, dto.DiskUsage{
			Mountpoint:   part.Mountpoint,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}

	return usages, nil
}

// GetNetworkConnections 网络连接统计：TCP 状态分布（含 TIME_WAIT/CLOSE_WAIT/OTHER）
func (s *Service) GetNetworkConnections() (*dto.NetworkStats, error) {
	conns, err := net.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("get network connections: %w", err)
	}

	stats := &dto.NetworkStats{TotalConnections: len(conns)}
	for _, conn := range conns {
		switch conn.Status {
		case "ESTABLISHED":
			stats.Established++
		case "LISTEN":
			stats.Listening++
		case "TIME_WAIT":
			stats.TimeWait++
		case "CLOSE_WAIT":
			stats.CloseWait++
		default:
			stats.Other++
		}
	}

	return stats, nil
}
