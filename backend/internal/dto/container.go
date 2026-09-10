package dto

// ContainerResponse 容器响应
type ContainerResponse struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	Ports   string            `json:"ports"`
	Created int64             `json:"created"`
	State   string            `json:"state"`
	Labels  map[string]string `json:"labels,omitempty"`
}

// ContainerLogResponse 容器日志响应
type ContainerLogResponse struct {
	Content string `json:"content"`
}

// ImageResponse 镜像响应
type ImageResponse struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created int64    `json:"created"`
}

// ContainerStatsResponse 容器实时资源占用响应
type ContainerStatsResponse struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   uint64  `json:"memory_usage"`
	MemoryLimit   uint64  `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
}
