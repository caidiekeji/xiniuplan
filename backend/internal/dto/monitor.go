package dto

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryRSS     uint64  `json:"memory_rss"`
	Status        string  `json:"status"`
	CreateTime    int64   `json:"create_time"`
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Mountpoint   string  `json:"mountpoint"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkStats 网络连接统计
type NetworkStats struct {
	Established int `json:"established"`
	Listening   int `json:"listening"`
	Total       int `json:"total"`
}
