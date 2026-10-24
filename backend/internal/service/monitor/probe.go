package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// ProbeService 网站拨测（roadmap #6）
type ProbeService struct {
	db       *gorm.DB
	notifier Notifier
	stopCh   chan struct{}
	running  bool
}

// Notifier 拨测告警通知接口（由 alert NotifierManager 适配实现）
type Notifier interface {
	SendToAll(title, content string) error
}

// NewProbeService 创建拨测服务
func NewProbeService(db *gorm.DB, notifier Notifier) *ProbeService {
	return &ProbeService{db: db, notifier: notifier}
}

// Start 启动拨测循环
func (s *ProbeService) Start() {
	if s.running {
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	go s.run()
	slog.Info("[probe] 拨测服务已启动")
}

// Stop 停止
func (s *ProbeService) Stop() {
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

func (s *ProbeService) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	s.runAll()
	for {
		select {
		case <-ticker.C:
			s.runAll()
		case <-s.stopCh:
			return
		}
	}
}

// runAll 对所有启用的探针执行探测（按各自 interval 控制节流）
func (s *ProbeService) runAll() {
	var probes []model.ProbeMonitor
	if err := s.db.Where("enabled = ?", true).Find(&probes).Error; err != nil {
		return
	}
	for i := range probes {
		p := &probes[i]
		if p.LastChecked != nil && time.Since(*p.LastChecked) < time.Duration(p.IntervalSec)*time.Second {
			continue
		}
		s.runOne(p)
	}
}

// runOne 执行单个探测：写结果、更新状态、状态翻转时告警
func (s *ProbeService) runOne(p *model.ProbeMonitor) {
	status, code, latency, msg := s.doProbe(p)
	now := time.Now()

	result := model.ProbeResult{
		ProbeID: p.ID, Status: status, Code: code, LatencyMs: latency, Message: msg, CheckedAt: now,
	}
	s.db.Create(&result)
	// 清理旧结果：仅保留最近 200 条
	s.db.Where("probe_id = ? AND id < (SELECT COALESCE(MAX(id),0) - 200 FROM probe_results WHERE probe_id = ?)", p.ID, p.ID).
		Delete(&model.ProbeResult{})

	prev := p.LastStatus
	p.LastStatus, p.LastCode, p.LastLatency, p.LastChecked = status, code, latency, &now
	s.db.Model(p).Updates(map[string]any{
		"last_status": status, "last_code": code, "last_latency": latency, "last_checked": now,
	})

	// 状态翻转告警（up→down 立即告警；down→up 发恢复通知）
	if prev != "" && prev != status && s.notifier != nil {
		title := fmt.Sprintf("拨测告警: %s %s", p.Name, status)
		content := fmt.Sprintf("探针 [%s] 状态由 %s 变为 %s\n目标: %s\n状态码: %d\n延迟: %.0fms\n信息: %s",
			p.Name, prev, status, p.Target, code, latency, msg)
		if err := s.notifier.SendToAll(title, content); err != nil {
			slog.Error("[probe] 告警发送失败", "probe", p.Name, "error", err)
		}
	}
}

// doProbe 执行探测
func (s *ProbeService) doProbe(p *model.ProbeMonitor) (status string, code int, latency float64, msg string) {
	start := time.Now()
	switch p.Type {
	case "tcp":
		host := p.Target
		if !strings.Contains(host, ":") {
			host += ":80"
		}
		conn, err := net.DialTimeout("tcp", host, 10*time.Second)
		latency = float64(time.Since(start).Milliseconds())
		if err != nil {
			return "down", 0, latency, err.Error()
		}
		conn.Close()
		return "up", 0, latency, ""
	default: // http
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(p.Target)
		latency = float64(time.Since(start).Milliseconds())
		if err != nil {
			return "down", 0, latency, err.Error()
		}
		defer resp.Body.Close()
		code = resp.StatusCode
		// 2xx/3xx 视为 up；4xx/5xx 视为 down
		if code >= 400 {
			return "down", code, latency, fmt.Sprintf("HTTP %d", code)
		}
		if p.Keyword != "" {
			buf := make([]byte, 64*1024)
			n, _ := resp.Body.Read(buf)
			if !strings.Contains(string(buf[:n]), p.Keyword) {
				return "down", code, latency, "响应体未包含关键词: " + p.Keyword
			}
		}
		return "up", code, latency, ""
	}
}

// ===== CRUD =====

func (s *ProbeService) List() ([]model.ProbeMonitor, error) {
	list := make([]model.ProbeMonitor, 0)
	err := s.db.Order("created_at ASC").Find(&list).Error
	return list, err
}

func (s *ProbeService) Create(p *model.ProbeMonitor) error {
	if p.Name == "" || p.Target == "" {
		return fmt.Errorf("名称和目标不能为空")
	}
	if p.Type != "http" && p.Type != "tcp" {
		return fmt.Errorf("类型必须为 http 或 tcp")
	}
	if p.IntervalSec <= 0 {
		p.IntervalSec = 60
	}
	p.LastStatus = "unknown"
	return s.db.Create(p).Error
}

func (s *ProbeService) Update(id uint, req *model.ProbeMonitor) error {
	var p model.ProbeMonitor
	if err := s.db.First(&p, id).Error; err != nil {
		return err
	}
	return s.db.Model(&p).Updates(map[string]any{
		"name": req.Name, "type": req.Type, "target": req.Target,
		"interval_sec": req.IntervalSec, "keyword": req.Keyword, "enabled": req.Enabled,
	}).Error
}

func (s *ProbeService) Delete(id uint) error {
	s.db.Where("probe_id = ?", id).Delete(&model.ProbeResult{})
	return s.db.Delete(&model.ProbeMonitor{}, id).Error
}

func (s *ProbeService) Results(probeID uint, limit int) ([]model.ProbeResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	list := make([]model.ProbeResult, 0)
	err := s.db.Where("probe_id = ?", probeID).Order("checked_at DESC").Limit(limit).Find(&list).Error
	return list, err
}

// CheckNow 立即探测一次
func (s *ProbeService) CheckNow(id uint) error {
	var p model.ProbeMonitor
	if err := s.db.First(&p, id).Error; err != nil {
		return err
	}
	s.runOne(&p)
	return nil
}

var _ = context.Background
