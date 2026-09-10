package monitor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"github.com/shirou/gopsutil/v3/host"
	"gorm.io/gorm"
)

const PanelVersion = "1.0.0"

// Service 监控服务
type Service struct {
	db        *gorm.DB
	collector *Collector
}

// NewService 创建监控服务
func NewService(db *gorm.DB, collector *Collector) *Service {
	return &Service{
		db:        db,
		collector: collector,
	}
}

// GetOverview 获取资源概览
func (s *Service) GetOverview(ctx context.Context) (*dto.OverviewResponse, error) {
	snap, err := s.collector.Snapshot(ctx)
	if err != nil {
		return nil, fmt.Errorf("collect snapshot: %w", err)
	}

	sysInfo, err := s.getSystemInfo()
	if err != nil {
		sysInfo = &dto.SystemInfo{}
	}

	return &dto.OverviewResponse{
		CPU: dto.CPUInfo{
			UsagePercent: snap.CPUUsage,
			Cores:        snap.CPUCores,
		},
		Memory: dto.MemoryInfo{
			Total:        snap.MemTotal,
			Used:         snap.MemUsed,
			Free:         snap.MemFree,
			UsagePercent: snap.MemPercent,
		},
		Disk: dto.DiskInfo{
			Total:        snap.DiskTotal,
			Used:         snap.DiskUsed,
			Free:         snap.DiskFree,
			UsagePercent: snap.DiskPercent,
		},
		Network: dto.NetworkInfo{
			BytesSentPerSec: snap.NetSentPerSec,
			BytesRecvPerSec: snap.NetRecvPerSec,
		},
		System: *sysInfo,
	}, nil
}

// GetMetrics 获取历史趋势数据（同时返回 CPU 和内存）
// timeRange: 1h | 6h | 24h | 7d
func (s *Service) GetMetrics(metricType, timeRange string) (*dto.MetricsResponse, error) {
	var duration time.Duration
	var aggregationSec int64 // 聚合粒度（秒）

	switch timeRange {
	case "1h":
		duration = 1 * time.Hour
		aggregationSec = 60 // 分钟聚合
	case "6h":
		duration = 6 * time.Hour
		aggregationSec = 300 // 5分钟聚合
	case "24h":
		duration = 24 * time.Hour
		aggregationSec = 3600 // 小时聚合
	case "7d":
		duration = 7 * 24 * time.Hour
		aggregationSec = 3600 // 小时聚合
	default:
		duration = 24 * time.Hour
		aggregationSec = 3600
	}

	now := time.Now()
	cutoff := now.Add(-duration).Unix()

	// 同时查询 CPU 和内存
	var rawMetrics []model.MonitorMetric
	if err := s.db.Where("metric_type IN ? AND timestamp >= ?", []string{"cpu", "memory"}, cutoff).
		Order("timestamp ASC").Find(&rawMetrics).Error; err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}

	// 聚合
	points := aggregateDualMetrics(rawMetrics, aggregationSec)

	return &dto.MetricsResponse{
		Type:   "cpu_memory",
		Range:  timeRange,
		Points: points,
	}, nil
}

// aggregateDualMetrics 按时间桶聚合 CPU 和内存双指标
func aggregateDualMetrics(metrics []model.MonitorMetric, bucketSec int64) []dto.MetricPoint {
	if len(metrics) == 0 {
		return []dto.MetricPoint{}
	}

	type bucketVals struct {
		cpuVals []float64
		memVals []float64
	}
	buckets := make(map[int64]*bucketVals)

	for _, m := range metrics {
		bucket := (m.Timestamp / bucketSec) * bucketSec
		if _, ok := buckets[bucket]; !ok {
			buckets[bucket] = &bucketVals{}
		}
		if m.MetricType == "cpu" {
			buckets[bucket].cpuVals = append(buckets[bucket].cpuVals, m.Value)
		} else if m.MetricType == "memory" {
			buckets[bucket].memVals = append(buckets[bucket].memVals, m.Value)
		}
	}

	points := make([]dto.MetricPoint, 0, len(buckets))
	for ts, bv := range buckets {
		var cpuAvg, memAvg float64
		if len(bv.cpuVals) > 0 {
			for _, v := range bv.cpuVals {
				cpuAvg += v
			}
			cpuAvg /= float64(len(bv.cpuVals))
		}
		if len(bv.memVals) > 0 {
			for _, v := range bv.memVals {
				memAvg += v
			}
			memAvg /= float64(len(bv.memVals))
		}
		points = append(points, dto.MetricPoint{
			Timestamp: ts,
			CPU:       cpuAvg,
			Memory:    memAvg,
		})
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp < points[j].Timestamp
	})

	return points
}

// GetRecentAuditLogs 获取最近20条操作日志
func (s *Service) GetRecentAuditLogs() ([]dto.AuditLogItem, error) {
	logs := make([]model.AuditLog, 0)
	if err := s.db.Order("created_at DESC").Limit(20).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}

	items := make([]dto.AuditLogItem, 0, len(logs))
	for _, l := range logs {
		items = append(items, dto.AuditLogItem{
			ID:        l.ID,
			UserID:    l.UserID,
			Action:    l.Action,
			Resource:  l.Resource,
			IP:        l.IP,
			UserAgent: l.UserAgent,
			Result:    l.Result,
			CreatedAt: l.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}

// getSystemInfo 获取系统信息
func (s *Service) getSystemInfo() (*dto.SystemInfo, error) {
	info := &dto.SystemInfo{
		PanelVersion: PanelVersion,
	}

	if h, err := host.Info(); err == nil {
		info.Hostname = h.Hostname
		info.OS = h.OS + " " + h.PlatformVersion
		info.Kernel = h.KernelVersion
		info.Uptime = h.Uptime
	}

	// 尝试获取公网 IP
	info.PublicIP = getPublicIP()

	return info, nil
}

// getPublicIP 尝试获取公网 IP（超时3秒）
func getPublicIP() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}
