package model

// ContainerInfo 容器信息（实时从 Docker 获取，不持久化到 DB）
type ContainerInfo struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Ports   string            `json:"ports"`
	Created int64             `json:"created"`
	Labels  map[string]string `json:"labels"`
}
