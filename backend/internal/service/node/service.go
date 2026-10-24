package node

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Service 多节点管理（roadmap #13）：主控端通过节点面板签发的 API Key 拉取只读资源概览。
type Service struct {
	db     *gorm.DB
	client *http.Client
	stopCh chan struct{}
	mu     sync.Mutex
	busy   bool
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, client: &http.Client{Timeout: 10 * time.Second}}
}

// List 节点列表
func (s *Service) List() ([]model.Node, error) {
	var nodes []model.Node
	err := s.db.Order("id ASC").Find(&nodes).Error
	return nodes, err
}

// Create 添加节点并立即握手（拉取一次概览验证连通性与 Key 有效性）
func (s *Service) Create(n *model.Node) error {
	n.BaseURL = strings.TrimRight(strings.TrimSpace(n.BaseURL), "/")
	if !strings.HasPrefix(n.BaseURL, "http://") && !strings.HasPrefix(n.BaseURL, "https://") {
		return fmt.Errorf("节点地址需以 http:// 或 https:// 开头")
	}
	if strings.TrimSpace(n.APIKey) == "" {
		return fmt.Errorf("请填写节点面板签发的 API Key")
	}
	var count int64
	s.db.Model(&model.Node{}).Where("base_url = ?", n.BaseURL).Count(&count)
	if count > 0 {
		return fmt.Errorf("该节点已存在")
	}
	if err := s.refreshOne(n); err != nil {
		return fmt.Errorf("节点握手失败: %w", err)
	}
	return s.db.Create(n).Error
}

// Delete 删除节点
func (s *Service) Delete(id uint) error {
	return s.db.Delete(&model.Node{}, id).Error
}

// SetEnabled 启用/停用节点的定时刷新
func (s *Service) SetEnabled(id uint, enabled bool) error {
	return s.db.Model(&model.Node{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// Refresh 手动刷新单个节点
func (s *Service) Refresh(id uint) (*model.Node, error) {
	var n model.Node
	if err := s.db.First(&n, id).Error; err != nil {
		return nil, fmt.Errorf("节点不存在")
	}
	if err := s.refreshOne(&n); err != nil {
		n.LastStatus = "offline"
		n.LastError = err.Error()
		s.db.Save(&n)
		return &n, err
	}
	return &n, s.db.Save(&n).Error
}

// refreshOne 拉取节点概览并更新快照字段
func (s *Service) refreshOne(n *model.Node) error {
	req, err := http.NewRequest(http.MethodGet, n.BaseURL+"/api/v1/dashboard/overview", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", n.APIKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("节点返回 HTTP %d", resp.StatusCode)
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			CPU struct {
				UsagePercent float64 `json:"usage_percent"`
			} `json:"cpu"`
			Memory struct {
				UsagePercent float64 `json:"usage_percent"`
			} `json:"memory"`
			Disk struct {
				UsagePercent float64 `json:"usage_percent"`
			} `json:"disk"`
			System struct {
				Hostname     string `json:"hostname"`
				PanelVersion string `json:"panel_version"`
			} `json:"system"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("解析节点响应失败: %w", err)
	}
	if body.Code != 0 {
		return fmt.Errorf("节点拒绝访问（code=%d），请确认 API Key 有效", body.Code)
	}
	now := time.Now()
	n.Hostname = body.Data.System.Hostname
	n.Version = body.Data.System.PanelVersion
	n.CPUPercent = body.Data.CPU.UsagePercent
	n.MemPercent = body.Data.Memory.UsagePercent
	n.DiskPercent = body.Data.Disk.UsagePercent
	n.LastStatus = "online"
	n.LastError = ""
	n.LastSeenAt = &now
	return nil
}

// Start 启动定时刷新循环（每分钟一轮，仅刷新启用节点）
func (s *Service) Start() {
	s.stopCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.RefreshAll()
			case <-s.stopCh:
				return
			}
		}
	}()
	slog.Info("[node] 多节点刷新已启动")
}

// Stop 停止刷新循环
func (s *Service) Stop() {
	if s.stopCh != nil {
		close(s.stopCh)
	}
}

// RefreshAll 刷新全部启用节点（并发上限 1，避免重入）
func (s *Service) RefreshAll() {
	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		return
	}
	s.busy = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.busy = false
		s.mu.Unlock()
	}()

	var nodes []model.Node
	s.db.Where("enabled = ?", true).Find(&nodes)
	for i := range nodes {
		n := &nodes[i]
		if err := s.refreshOne(n); err != nil {
			n.LastStatus = "offline"
			n.LastError = err.Error()
		}
		s.db.Save(n)
	}
}
