package dto

// SettingsResponse 设置响应
type SettingsResponse struct {
	Settings map[string]string `json:"settings"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// VersionResponse 版本信息响应
type VersionResponse struct {
	PanelVersion  string `json:"panel_version"`
	GoVersion     string `json:"go_version"`
	BuildTime     string `json:"build_time"`
	Uptime        uint64 `json:"uptime"`
	LatestVersion string `json:"latest_version,omitempty"`
	HasUpdate     bool   `json:"has_update"`
}

// CleanRequest 清理请求
type CleanRequest struct {
	CleanType string `json:"clean_type" binding:"required,oneof=cache docker logs"`
}

// SystemInfoResponse 详细系统信息响应
type SystemInfoResponse struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	CPUModel     string `json:"cpu_model"`
	CPUCores     int    `json:"cpu_cores"`
	MemoryTotal  uint64 `json:"memory_total"`
	DiskTotal    uint64 `json:"disk_total"`
	PublicIP     string `json:"public_ip"`
	PanelVersion string `json:"panel_version"`
}
