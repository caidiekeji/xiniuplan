package audit

import (
	"fmt"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// --- 异常登录检测（S-012） ---

// AnomalyType 异常类型
type AnomalyEvent struct {
	Type      string    `json:"type"`    // brute_force_ip | brute_force_user | new_ip_login
	Subject   string    `json:"subject"` // IP 或用户名
	Count     int64     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Detail    string    `json:"detail"`
}

// DetectAnomalies 检测最近 24h 的异常登录行为：
// 1) 同一 IP 登录失败 ≥ 5 次（疑似暴力破解）
// 2) 同一用户名登录失败 ≥ 5 次（疑似撞库）
// 3) 成功登录的 IP 在该用户历史成功 IP 中首次出现（非常用地登录）
func DetectAnomalies(db *gorm.DB) ([]AnomalyEvent, error) {
	events := make([]AnomalyEvent, 0)
	cutoff := time.Now().Add(-24 * time.Hour)

	// 1. 按 IP 聚合失败登录
	type agg struct {
		IP        string
		Count     int64
		FirstSeen time.Time
		LastSeen  time.Time
	}
	var byIP []agg
	if err := db.Model(&model.AuditLog{}).
		Select("ip, count(*) as count, min(created_at) as first_seen, max(created_at) as last_seen").
		Where("action = ? AND result = ? AND created_at > ?", "login", "fail", cutoff).
		Group("ip").Having("count(*) >= ?", 5).
		Scan(&byIP).Error; err == nil {
		for _, r := range byIP {
			events = append(events, AnomalyEvent{
				Type: "brute_force_ip", Subject: r.IP, Count: r.Count,
				FirstSeen: r.FirstSeen, LastSeen: r.LastSeen,
				Detail: fmt.Sprintf("IP %s 在 24 小时内登录失败 %d 次，疑似暴力破解", r.IP, r.Count),
			})
		}
	}

	// 2. 按用户名聚合失败登录（Resource 存用户名）
	var byUser []agg
	if err := db.Model(&model.AuditLog{}).
		Select("resource as ip, count(*) as count, min(created_at) as first_seen, max(created_at) as last_seen").
		Where("action = ? AND result = ? AND created_at > ?", "login", "fail", cutoff).
		Group("resource").Having("count(*) >= ?", 5).
		Scan(&byUser).Error; err == nil {
		for _, r := range byUser {
			events = append(events, AnomalyEvent{
				Type: "brute_force_user", Subject: r.IP, Count: r.Count,
				FirstSeen: r.FirstSeen, LastSeen: r.LastSeen,
				Detail: fmt.Sprintf("用户 %s 在 24 小时内登录失败 %d 次，疑似撞库攻击", r.IP, r.Count),
			})
		}
	}

	// 3. 非常用 IP 成功登录：成功 IP 不在该用户 30 天内历史成功 IP 集合的前置部分
	//    简化实现：对比最近 1h 成功登录 IP 与此前 30 天成功登录 IP 集合
	var recentLogins []model.AuditLog
	if err := db.Where("action = ? AND result = ? AND created_at > ?", "login", "success", time.Now().Add(-time.Hour)).
		Find(&recentLogins).Error; err == nil {
		for _, login := range recentLogins {
			var historyCount int64
			db.Model(&model.AuditLog{}).
				Where("action = ? AND result = ? AND user_id = ? AND ip = ? AND created_at < ?",
					"login", "success", login.UserID, login.IP, login.CreatedAt).
				Count(&historyCount)
			if historyCount == 0 {
				// 检查更早（30 天前）是否有过
				var olderCount int64
				db.Model(&model.AuditLog{}).
					Where("action = ? AND result = ? AND user_id = ? AND ip = ? AND created_at <= ?",
						"login", "success", login.UserID, login.IP, time.Now().Add(-30*24*time.Hour)).
					Count(&olderCount)
				if olderCount == 0 {
					events = append(events, AnomalyEvent{
						Type: "new_ip_login", Subject: login.IP, Count: 1,
						FirstSeen: login.CreatedAt, LastSeen: login.CreatedAt,
						Detail: fmt.Sprintf("用户 (ID %d) 首次从新 IP %s 成功登录", login.UserID, login.IP),
					})
				}
			}
		}
	}

	return events, nil
}

// OnLoginFail 登录失败后的实时检测钩子：同一 IP 短窗口失败过多时返回 true（供告警扩展）
func OnLoginFail(db *gorm.DB, ip string) bool {
	var count int64
	db.Model(&model.AuditLog{}).
		Where("action = ? AND result = ? AND ip = ? AND created_at > ?", "login", "fail", ip, time.Now().Add(-10*time.Minute)).
		Count(&count)
	return count >= 10
}
