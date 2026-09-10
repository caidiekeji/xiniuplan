package dto

// OverviewResponse 资源概览响应
type OverviewResponse struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    DiskInfo    `json:"disk"`
	Network NetworkInfo `json:"network"`
	System  SystemInfo  `json:"system"`
}

type CPUInfo struct {
	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`
}

type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

type NetworkInfo struct {
	BytesSentPerSec uint64 `json:"bytes_sent_per_sec"`
	BytesRecvPerSec uint64 `json:"bytes_recv_per_sec"`
}

type SystemInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	Uptime       uint64 `json:"uptime"`
	PanelVersion string `json:"panel_version"`
	PublicIP     string `json:"public_ip"`
}

// MetricPoint 单个指标点（同时包含 CPU 和内存）
type MetricPoint struct {
	Timestamp int64   `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
}

// MetricsResponse 历史趋势响应
type MetricsResponse struct {
	Type   string        `json:"type"`
	Range  string        `json:"range"`
	Points []MetricPoint `json:"points"`
}

// AuditLogItem 审计日志项
type AuditLogItem struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Result    string `json:"result"`
	CreatedAt string `json:"created_at"`
}
