package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
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
	// 网络速率差值计算的跨 tick 状态（collect 仅在单一 goroutine 中调用，无需加锁）
	lastSent uint64
	lastRecv uint64
	lastTime int64
	hasLast  bool
	// CPU 使用率跨 tick 差值状态（避免 cpu.Percent(interval) 阻塞采集周期）
	lastCPUTimes    cpu.TimesStat
	hasLastCPUTimes bool
	// snap 保存最近一次采集结果，供 Overview 接口零阻塞读取
	snap atomic.Pointer[SnapshotData]
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

// collect 后台采集一轮：采样 → 更新内存快照 → 落库
//
// CPU 使用率改用跨 tick 差值计算（原 cpu.Percent(c.interval) 会阻塞整个采集
// 周期 15 秒，导致实际采集间隔变成 30 秒），口径与 gopsutil cpu.Percent 一致。
func (c *Collector) collect() {
	now := time.Now().Unix()
	snap := &SnapshotData{}
	metrics := make([]model.MonitorMetric, 0, 8)

	// CPU 使用率（跨 tick 差值，不阻塞采集周期）
	if usage, ok := c.cpuUsage(); ok {
		snap.CPUUsage = usage
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "cpu",
			Value:      usage,
			Timestamp:  now,
		})
	}
	if counts, err := cpu.Counts(true); err == nil {
		snap.CPUCores = counts
	}

	// 内存使用率（使用 Available 计算，与 Linux 内核 MemAvailable 一致）
	if vm, err := mem.VirtualMemory(); err == nil {
		memPercent := float64(vm.Total-vm.Available) / float64(vm.Total) * 100
		snap.MemTotal = vm.Total
		snap.MemUsed = vm.Total - vm.Available
		snap.MemFree = vm.Available
		snap.MemPercent = memPercent
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "memory",
			Value:      memPercent,
			Timestamp:  now,
		})
	}

	// 磁盘使用率（根分区）
	if d, err := disk.Usage("/"); err == nil {
		snap.DiskTotal = d.Total
		snap.DiskUsed = d.Used
		snap.DiskFree = d.Free
		snap.DiskPercent = d.UsedPercent
		metrics = append(metrics, model.MonitorMetric{
			MetricType: "disk",
			Value:      d.UsedPercent,
			Timestamp:  now,
		})
	}

	// 网络速率：用两次采样的差值换算为每秒字节数（B/s），
	// 不能直接存累计字节数——累计值画趋势图只会是单调递增直线
	if ioStats, err := net.IOCounters(false); err == nil && len(ioStats) > 0 {
		totalSent := ioStats[0].BytesSent
		totalRecv := ioStats[0].BytesRecv

		if c.hasLast && now > c.lastTime {
			elapsed := float64(now - c.lastTime)
			sentRate := float64(totalSent-c.lastSent) / elapsed
			recvRate := float64(totalRecv-c.lastRecv) / elapsed
			// 计数器回绕（重启）时差值为负，跳过本点
			if totalSent >= c.lastSent && totalRecv >= c.lastRecv {
				snap.NetSentPerSec = uint64(sentRate)
				snap.NetRecvPerSec = uint64(recvRate)
				metrics = append(metrics, model.MonitorMetric{
					MetricType: "network_sent",
					Value:      sentRate,
					Timestamp:  now,
				})
				metrics = append(metrics, model.MonitorMetric{
					MetricType: "network_recv",
					Value:      recvRate,
					Timestamp:  now,
				})
			}
		}

		c.lastSent = totalSent
		c.lastRecv = totalRecv
		c.lastTime = now
		c.hasLast = true
	}

	// 内存快照：供 Overview 接口零阻塞读取
	c.snap.Store(snap)

	if len(metrics) > 0 && c.db != nil {
		if err := c.db.Create(&metrics).Error; err != nil {
			slog.Error("[monitor] 保存指标失败", "error", err)
		}
	}
}

// cpuUsage 计算当前 CPU 使用率。
//
// 首轮：间隔 1 秒取两次样本建立基准（与 gopsutil cpu.Percent 同口径）；
// 后续轮次：用上一轮的时间片与本轮时间片做差值，零阻塞。
func (c *Collector) cpuUsage() (float64, bool) {
	times, err := cpu.Times(false)
	if err != nil || len(times) == 0 {
		return 0, false
	}
	now := times[0]
	if !c.hasLastCPUTimes {
		time.Sleep(time.Second)
		later, err := cpu.Times(false)
		if err != nil || len(later) == 0 {
			return 0, false
		}
		c.lastCPUTimes = later[0]
		c.hasLastCPUTimes = true
		return busyPercent(now, later[0]), true
	}
	usage := busyPercent(c.lastCPUTimes, now)
	c.lastCPUTimes = now
	return usage, true
}

// busyPercent 计算两次 CPU 时间片采样之间的使用率百分比。
// 口径与 gopsutil cpu.Percent 完全一致（getAllBusy + calculateBusy），
// 保证落库的历史趋势数据与原实现可比。
func busyPercent(t1, t2 cpu.TimesStat) float64 {
	all1, busy1 := allBusy(t1)
	all2, busy2 := allBusy(t2)
	if busy2 <= busy1 {
		return 0
	}
	if all2 <= all1 {
		return 100
	}
	v := (busy2 - busy1) / (all2 - all1) * 100
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// allBusy 计算总时间片与忙碌时间片（口径同 gopsutil getAllBusy）
func allBusy(t cpu.TimesStat) (float64, float64) {
	tot := t.Total() - t.Guest - t.GuestNice // Linux 上 Guest 已计入 User/System，需扣除避免重复
	busy := tot - t.Idle - t.Iowait
	return tot, busy
}

// Snapshot 返回最近一次采集的快照（零阻塞）
//
// 历史实现是"现场采样"：cpu.Percent(1*time.Second) 阻塞 1 秒 + 网络速率的
// 两次采样间隔 sleep 1 秒，固定阻塞 2 秒，直接导致仪表盘首屏接口
// /dashboard/overview 耗时 2.2 秒（2026-10-23 线上实测，192.168.31.18）。
//
// 现改为读后台采集器的内存快照（每 interval=15s 更新一次，最新鲜度 15s）。
// 监控面板展示 15 秒粒度的实时值语义上完全可接受。
func (c *Collector) Snapshot(ctx context.Context) (*SnapshotData, error) {
	if s := c.snap.Load(); s != nil {
		return s, nil
	}
	// 采集器尚未完成首轮采集（进程刚启动）：现场采一次兜底，
	// 只取瞬时指标（内存/磁盘读取无阻塞），不等待 CPU 采样间隔
	return c.sampleInstant(), nil
}

// sampleInstant 现场采集瞬时指标（不等待任何采样间隔）
func (c *Collector) sampleInstant() *SnapshotData {
	data := &SnapshotData{}
	if counts, err := cpu.Counts(true); err == nil {
		data.CPUCores = counts
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		data.MemTotal = vm.Total
		data.MemUsed = vm.Total - vm.Available
		data.MemFree = vm.Available
		data.MemPercent = float64(vm.Total-vm.Available) / float64(vm.Total) * 100
	}
	if d, err := disk.Usage("/"); err == nil {
		data.DiskTotal = d.Total
		data.DiskUsed = d.Used
		data.DiskFree = d.Free
		data.DiskPercent = d.UsedPercent
	}
	return data
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
