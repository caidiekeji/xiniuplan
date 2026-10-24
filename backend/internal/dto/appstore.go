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
	// Success 安装是否成功——前端 InstallDialog 依此判定成败，
	// 此前该字段缺失导致「安装成功却提示失败」（2026-10-23 闭环审计）
	Success         bool              `json:"success"`
	ID              uint              `json:"id"`
	Key             string            `json:"key"`
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	InstallPath     string            `json:"install_path"`
	Status          string            `json:"status"`
	EnvParams       map[string]string `json:"env_params"`
	ContainerStatus string            `json:"container_status"`
	// ContainerID compose 容器 ID（前端「查看日志」依赖）
	ContainerID string `json:"container_id"`
	// HostPort 宿主机映射端口（前端拼装访问地址依赖）
	HostPort int `json:"host_port"`
	// AccessURL 应用访问地址（由 api 层按请求 Host 拼装）
	AccessURL string `json:"access_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
