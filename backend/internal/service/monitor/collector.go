package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloudpivot/internal/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"gorm.io/gorm"
)

// Collector 监控采集器
type Collector struct {
	db       *gorm.DB
	interval time.Duration
	stopCh   chan struct{}
}

// NewCollector 创建采集器
func NewCollector(db *gorm.DB, intervalSec int) *Collector {
	if intervalSec <= 0 {
		intervalSec = 15
	}
	return &Collector{
		db:       db,
		interval: time.Duration(intervalSec) * time.Second,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动后台采集 goroutine
func (c *Collector) Start() {
	go c.run()
	slog.Info("[monitor] 采集器已启动", "interval", c.interval)
}

// Stop 停止采集器
func (c *Collector) Stop() {
	close(c.stopCh)
}

func (c *Collector) run() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 立即采集一次
	c.collect()

	for {
		select {
		case <-ticker.C:
			c.collect()
		case <-c.stopCh:
			slog.Info("[monitor] 采集器已停止")
			return
		}
	}
}

func (c *Collector) collect() {
	now := time.Now().Unix()
	metrics := make([]model.MonitorMetric, 0, 8)

	// CPU 使用率
	if cpuPercent, err := cpu.Percent(c.interval, false); err == nil && len(cpuPercent) > 0 {
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "cpu",
			Value:      cpuPercent[0],
			Timestamp:  now,
		})
	}

	// 内存使用率（使用 Available 计算，与 Linux 内核 MemAvailable 一致）
	if vm, err := mem.VirtualMemory(); err == nil {
		memPercent := float64(vm.Total-vm.Available) / float64(vm.Total) * 100
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "memory",
			Value:      memPercent,
			Timestamp:  now,
		})
	}

	// 磁盘使用率（根分区）
	if d, err := disk.Usage("/"); err == nil {
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "disk",
			Value:      d.UsedPercent,
			Timestamp:  now,
		})
	}

	// 网络速率（计算每秒字节数）
	if ioStats, err := net.IOCounters(false); err == nil && len(ioStats) > 0 {
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "network_sent",
			Value:      float64(ioStats[0].BytesSent),
			Timestamp:  now,
		})
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "network_recv",
			Value:      float64(ioStats[0].BytesRecv),
			Timestamp:  now,
		})
	}

	if len(metrics) > 0 && c.db != nil {
		if err := c.db.Create(&metrics).Error; err != nil {
			slog.Error("[monitor] 保存指标失败", "error", err)
		}
	}
}

// Snapshot 实时采集一次（用于 overview 接口）
func (c *Collector) Snapshot(ctx context.Context) (*SnapshotData, error) {
	data := &SnapshotData{}

	// CPU
	if cpuPercent, err := cpu.Percent(1*time.Second, false); err == nil && len(cpuPercent) > 0 {
		data.CPUUsage = cpuPercent[0]
	}
	if counts, err := cpu.Counts(true); err == nil {
		data.CPUCores = counts
	}

	// 内存（使用 Available 计算使用率，与 Linux 内核 MemAvailable 一致）
	if vm, err := mem.VirtualMemory(); err == nil {
		data.MemTotal = vm.Total
		data.MemUsed = vm.Total - vm.Available
		data.MemFree = vm.Available
		data.MemPercent = float64(vm.Total-vm.Available) / float64(vm.Total) * 100
	}

	// 磁盘
	if d, err := disk.Usage("/"); err == nil {
		data.DiskTotal = d.Total
		data.DiskUsed = d.Used
		data.DiskFree = d.Free
		data.DiskPercent = d.UsedPercent
	}

	// 网络（两次采样计算速率）
	if stats1, err := net.IOCounters(false); err == nil && len(stats1) > 0 {
		time.Sleep(1 * time.Second)
		if stats2, err := net.IOCounters(false); err == nil && len(stats2) > 0 {
			data.NetSentPerSec = stats2[0].BytesSent - stats1[0].BytesSent
			data.NetRecvPerSec = stats2[0].BytesRecv - stats1[0].BytesRecv
		}
	}

	return data, nil
}

// SnapshotData 实时快照数据
type SnapshotData struct {
	CPUUsage      float64
	CPUCores      int
	MemTotal      uint64
	MemUsed       uint64
	MemFree       uint64
	MemPercent    float64
	DiskTotal     uint64
	DiskUsed      uint64
	DiskFree      uint64
	DiskPercent   float64
	NetSentPerSec uint64
	NetRecvPerSec uint64
}

// CleanOldData 清理过期监控数据（保留7天）
func (c *Collector) CleanOldData() {
	if c.db == nil {
		return
	}
	cutoff := time.Now().AddDate(0, 0, -7).Unix()
	c.db.Where("timestamp < ?", cutoff).Delete(&model.MonitorMetric{})
	fmt.Println("[monitor] 过期监控数据已清理")
}
