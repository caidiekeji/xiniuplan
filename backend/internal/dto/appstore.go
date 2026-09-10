package dto

// EnvParam 环境变量参数定义
type EnvParam struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // string | password | number | boolean
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
}

// AppInfo 应用商店应用信息
type AppInfo struct {
	Key           string     `json:"key"`
	Name          string     `json:"name"`
	Version       string     `json:"version"`
	Description   string     `json:"description"`
	Icon          string     `json:"icon"`
	Category      string     `json:"category"`
	Tags          []string   `json:"tags"`
	RequiredPorts []int      `json:"required_ports"`
	EnvParams     []EnvParam `json:"env_params"`
	ComposeFile   string     `json:"compose_file"`
	MinMemoryMB   int        `json:"min_memory_mb"`
	MinDiskGB     int        `json:"min_disk_gb"`
	SourceID      uint       `json:"source_id"`
	SourceName    string     `json:"source_name"`
}

// InstallRequest 安装应用请求
type InstallRequest struct {
	SourceID  uint              `json:"source_id"`
	EnvParams map[string]string `json:"env_params"`
}

// InstalledAppResponse 已安装应用响应
type InstalledAppResponse struct {
	ID              uint              `json:"id"`
	Key             string            `json:"key"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	InstallPath     string            `json:"install_path"`
	Status          string            `json:"status"`
	EnvParams       map[string]string `json:"env_params"`
	ContainerStatus string            `json:"container_status"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

// AppStatusResponse 应用安装状态响应
type AppStatusResponse struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	ContainerStatus string `json:"container_status"`
	ContainerCount  int    `json:"container_count"`
	RunningCount    int    `json:"running_count"`
}
