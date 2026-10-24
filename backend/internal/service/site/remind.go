package site

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// SSL 到期提醒（roadmap #2）：每日扫描证书到期时间，梯度通知
type SSLReminder struct {
	db           *gorm.DB
	notifier     Notifier
	stopCh       chan struct{}
	lastNotified map[uint]string // siteID -> 最近提醒的等级（30/15/7），防重复
	mu           sync.Mutex
	running      bool
}

// Notifier 通知接口（alert NotifierManager 适配）
type Notifier interface {
	SendToAll(title, content string) error
}

// NewSSLReminder 创建 SSL 提醒服务
func NewSSLReminder(db *gorm.DB, notifier Notifier) *SSLReminder {
	return &SSLReminder{
		db:           db,
		notifier:     notifier,
		stopCh:       make(chan struct{}),
		lastNotified: map[uint]string{},
	}
}

// Start 启动每日检查（每 6 小时轮询一次，按天去重）
func (r *SSLReminder) Start() {
	if r.running {
		return
	}
	r.running = true
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		r.check()
		for {
			select {
			case <-ticker.C:
				r.check()
			case <-r.stopCh:
				return
			}
		}
	}()
	slog.Info("[ssl-reminder] SSL 到期提醒服务已启动")
}

// Stop 停止
func (r *SSLReminder) Stop() {
	if r.running {
		close(r.stopCh)
		r.running = false
	}
}

// remindDays 告警梯度（天）
var remindDays = []int{30, 15, 7, 3}

func (r *SSLReminder) check() {
	var sites []model.Site
	if err := r.db.Where("ssl_enabled = ? AND ssl_expire_at IS NOT NULL", true).Find(&sites).Error; err != nil {
		return
	}
	for _, site := range sites {
		if site.SSLExpireAt == nil {
			continue
		}
		days := int(time.Until(*site.SSLExpireAt).Hours() / 24)
		level := 0
		for _, d := range remindDays {
			if days <= d {
				level = d
				break
			}
		}
		if level == 0 || days < 0 {
			// 未到提醒窗口或已过期（已过期由证书状态本身体现）
			if days >= 0 {
				continue
			}
			level = 0
		}
		// 同一站点同一等级只提醒一次
		r.mu.Lock()
		last, seen := r.lastNotified[site.ID]
		key := fmt.Sprintf("%d", level)
		if seen && last == key {
			r.mu.Unlock()
			continue
		}
		r.lastNotified[site.ID] = key
		r.mu.Unlock()

		if r.notifier != nil {
			title := fmt.Sprintf("SSL 证书到期提醒: %s", site.PrimaryDomain)
			content := fmt.Sprintf("站点 %s 的 SSL 证书将于 %d 天后到期（%s）。\n建议尽快续期（Let's Encrypt 自动续期可在站点设置中开启）。",
				site.PrimaryDomain, days, site.SSLExpireAt.Format("2006-01-02"))
			if err := r.notifier.SendToAll(title, content); err != nil {
				slog.Error("[ssl-reminder] 通知发送失败", "site", site.PrimaryDomain, "error", err)
			} else {
				slog.Info("[ssl-reminder] 已发送到期提醒", "site", site.PrimaryDomain, "days", days)
			}
		}
	}
}
