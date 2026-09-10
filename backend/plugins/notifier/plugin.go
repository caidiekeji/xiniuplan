// Package notifier 通知插件：提供统一的消息发送接口，
// 向 notify 扩展点注册通知通道，支持日志、Webhook 等多种通知方式。
package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"cloudpivot/pkg/plugin"
)

// NotifyLevel 通知级别。
type NotifyLevel string

const (
	LevelInfo     NotifyLevel = "info"
	LevelWarning  NotifyLevel = "warning"
	LevelError    NotifyLevel = "error"
	LevelCritical NotifyLevel = "critical"
)

// Notification 通知消息结构。
type Notification struct {
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	Level   NotifyLevel            `json:"level"`
	Tags    []string               `json:"tags,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// NotifierPlugin 通知插件实现。
type NotifierPlugin struct {
	config     map[string]interface{}
	stopCh     chan struct{}
	stopOnce   sync.Once
	running    bool
	mu         sync.RWMutex
	history    []Notification
	maxHistory int
	httpClient *http.Client
}

// New 创建通知插件实例。
func New() *NotifierPlugin {
	return &NotifierPlugin{
		stopCh:     make(chan struct{}),
		history:    make([]Notification, 0),
		maxHistory: 100,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Info 返回插件元信息。
func (p *NotifierPlugin) Info() plugin.PluginInfo {
	return plugin.PluginInfo{
		Name:        "notifier",
		Version:     "1.0.0",
		Description: "统一消息通知插件，支持日志输出与 Webhook 推送，向系统注册通知扩展点",
		Author:      "cloudpivot",
		Icon:        "N",
	}
}

// Init 初始化插件，读取配置并向 notify 扩展点注册发送接口。
func (p *NotifierPlugin) Init(config map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

	// 读取历史记录上限
	if v, ok := config["max_history"]; ok {
		if n, ok := v.(float64); ok && n > 0 {
			p.maxHistory = int(n)
		}
	}

	// 向 notify 扩展点注册发送函数（通过闭包包装）
	mgr := plugin.DefaultManager()

	// 清理旧的 notifier 扩展
	exts := mgr.GetExtensions(plugin.ExtNotify)
	filtered := make([]interface{}, 0, len(exts))
	for _, e := range exts {
		if m, ok := e.(map[string]interface{}); ok {
			if m["plugin"] != "notifier" {
				filtered = append(filtered, e)
			}
		}
	}

	// 注册通知通道描述
	channel := map[string]interface{}{
		"plugin":       "notifier",
		"name":         "default",
		"display_name": "默认通知通道",
		"levels":       []string{string(LevelInfo), string(LevelWarning), string(LevelError), string(LevelCritical)},
		"methods":      p.enabledMethods(),
		"send":         p.Send, // 暴露发送函数供宿主调用
	}
	mgr.RegisterExtension(plugin.ExtNotify, channel)

	slog.Info("[plugin:notifier] initialized", "methods", p.enabledMethods())
	return nil
}

// Run 启动插件：维护通知队列与历史记录。
func (p *NotifierPlugin) Run() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("notifier plugin already running")
	}
	p.running = true
	p.stopCh = make(chan struct{})
	p.mu.Unlock()

	slog.Info("[plugin:notifier] running, notification service ready")

	// 发送启动通知
	p.Send(Notification{
		Title:   "通知插件已启动",
		Content: fmt.Sprintf("通知服务已就绪，支持方式: %v", p.enabledMethods()),
		Level:   LevelInfo,
		Tags:    []string{"system", "plugin"},
	})

	<-p.stopCh
	slog.Info("[plugin:notifier] stopped")
	return nil
}

// Stop 停止插件。
func (p *NotifierPlugin) Stop() error {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()
	return nil
}

// Send 发送通知（核心方法，可被宿主通过扩展点调用）。
func (p *NotifierPlugin) Send(n Notification) error {
	p.mu.Lock()
	// 记录历史
	p.history = append(p.history, n)
	if len(p.history) > p.maxHistory {
		p.history = p.history[len(p.history)-p.maxHistory:]
	}
	p.mu.Unlock()

	// 1. 日志输出（始终启用）
	slog.Info("[notifier] notification", "level", n.Level, "title", n.Title, "content", n.Content)

	// 2. Webhook 推送（如配置）
	if webhookURL, ok := p.configString("webhook_url"); ok && webhookURL != "" {
		go p.sendWebhook(webhookURL, n)
	}

	return nil
}

// History 返回最近的通知历史。
func (p *NotifierPlugin) History(limit int) []Notification {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if limit <= 0 || limit > len(p.history) {
		limit = len(p.history)
	}
	result := make([]Notification, limit)
	copy(result, p.history[len(p.history)-limit:])
	return result
}

// sendWebhook 通过 Webhook 发送通知。
func (p *NotifierPlugin) sendWebhook(url string, n Notification) {
	body, err := json.Marshal(n)
	if err != nil {
		slog.Error("[plugin:notifier] webhook marshal error", "error", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		slog.Error("[plugin:notifier] webhook request error", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		slog.Error("[plugin:notifier] webhook send error", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("[plugin:notifier] webhook returned non-2xx status", "status", resp.StatusCode)
	}
}

// enabledMethods 返回当前启用的通知方式列表。
func (p *NotifierPlugin) enabledMethods() []string {
	methods := []string{"log"}
	if url, ok := p.configString("webhook_url"); ok && url != "" {
		methods = append(methods, "webhook")
	}
	return methods
}

// configString 从配置中读取字符串值。
func (p *NotifierPlugin) configString(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.config == nil {
		return "", false
	}
	v, ok := p.config[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
