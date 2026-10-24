package model

import "time"

// ============ 拨测监控（roadmap #6） ============

// ProbeMonitor 拨测监控配置
type ProbeMonitor struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:64;not null" json:"name"`
	Type        string     `gorm:"size:16;not null" json:"type"`    // http | tcp
	Target      string     `gorm:"size:512;not null" json:"target"` // http(s)://URL 或 host:port
	IntervalSec int        `gorm:"not null;default:60" json:"interval_sec"`
	Keyword     string     `gorm:"size:128" json:"keyword"` // http 类型可选：响应体必须包含的关键词
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	LastStatus  string     `gorm:"size:16" json:"last_status"` // up | down | unknown
	LastCode    int        `json:"last_code"`
	LastLatency float64    `json:"last_latency_ms"`
	LastChecked *time.Time `json:"last_checked"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ProbeResult 拨测结果（每探针保留最近 N 条，由服务层清理）
type ProbeResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProbeID   uint      `gorm:"not null;index" json:"probe_id"`
	Status    string    `gorm:"size:16;not null" json:"status"` // up | down
	Code      int       `json:"code"`
	LatencyMs float64   `json:"latency_ms"`
	Message   string    `gorm:"size:512" json:"message"`
	CheckedAt time.Time `gorm:"index" json:"checked_at"`
}

// ============ 进程守护（roadmap #11） ============

// DaemonProcess 守护进程配置
type DaemonProcess struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:64;not null;uniqueIndex" json:"name"`
	Command      string     `gorm:"size:512;not null" json:"command"`
	WorkDir      string     `gorm:"size:256" json:"work_dir"`
	Enabled      bool       `gorm:"not null;default:true" json:"enabled"`
	AutoStart    bool       `gorm:"not null;default:true" json:"auto_start"` // 面板启动时自动拉起
	PID          int        `json:"pid"`
	Restarts     int        `json:"restarts"`
	LastStarted  *time.Time `json:"last_started"`
	LastExitCode int        `json:"last_exit_code"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ============ Git 部署（roadmap #9，PaaS 化第一步） ============

// DeployApp Git 部署应用
type DeployApp struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"size:64;not null;uniqueIndex" json:"name"`
	RepoURL       string     `gorm:"size:512;not null" json:"repo_url"` // 可含凭据 https://user:token@host/...
	Branch        string     `gorm:"size:64;not null;default:main" json:"branch"`
	Secret        string     `gorm:"size:128;not null" json:"secret"`    // webhook 校验令牌
	BuildType     string     `gorm:"size:16;not null" json:"build_type"` // static | node | dockerfile
	OutDir        string     `gorm:"size:128" json:"out_dir"`            // node 构建产物目录（默认 dist）
	DestPath      string     `gorm:"size:256;not null" json:"dest_path"` // 部署目标目录（static/node）
	ContainerPort int        `json:"container_port"`                     // dockerfile 类型：容器端口
	Enabled       bool       `gorm:"not null;default:true" json:"enabled"`
	LastDeployAt  *time.Time `json:"last_deploy_at"`
	LastStatus    string     `gorm:"size:16" json:"last_status"` // success | failed | deploying
	CreatedAt     time.Time  `json:"created_at"`
}

// DeployLog 部署日志
type DeployLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AppID     uint      `gorm:"not null;index" json:"app_id"`
	Status    string    `gorm:"size:16;not null" json:"status"` // success | failed
	Output    string    `gorm:"type:text" json:"output"`
	CreatedAt time.Time `json:"created_at"`
}
