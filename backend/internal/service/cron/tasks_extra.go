package cron

import (
	"fmt"
	"strings"
	"time"

	"cloudpivot/internal/model"
	alertsvc "cloudpivot/internal/service/alert"
	filesvc "cloudpivot/internal/service/file"
	sitesvc "cloudpivot/internal/service/site"

	"gorm.io/gorm"
)

// --- 新增计划任务类型 ---
// ssl_check      : 扫描全部站点证书，30/15/7 天梯度到期提醒（走通知渠道）
// tamper_check   : 校验全部已建基线站点的防篡改差异，发现篡改即告警
// recycle_cleanup: 清理 7 天前回收站文件

// execSSLCheck 证书到期检查
func (s *Service) execSSLCheck() (string, error) {
	var sites []model.Site
	s.db.Where("ssl_cert_path != ''").Find(&sites)

	var warnings []string
	for _, site := range sites {
		if site.SSLExpireAt == nil {
			continue // 无到期时间的站点（手动部署证书未登记）跳过
		}
		daysLeft := int(time.Until(*site.SSLExpireAt).Hours() / 24)
		if daysLeft < 0 {
			warnings = append(warnings, fmt.Sprintf("[已过期 %d 天] %s", -daysLeft, site.PrimaryDomain))
		} else if daysLeft <= 30 {
			warnings = append(warnings, fmt.Sprintf("[%d 天后到期] %s", daysLeft, site.PrimaryDomain))
		}
	}

	if len(warnings) == 0 {
		return fmt.Sprintf("证书检查完成：%d 个站点证书均在有效期内（>30 天）", len(sites)), nil
	}
	out := fmt.Sprintf("证书检查完成：%d 个证书临近或已过期\n%s", len(warnings), strings.Join(warnings, "\n"))
	// 告警走默认通知渠道（邮件等），失败不影响任务结果
	_ = s.notify("SSL 证书到期提醒", out)
	return out, nil
}

// execTamperCheck 全站防篡改校验
func (s *Service) execTamperCheck() (string, error) {
	var baselines []model.SiteTamperBaseline
	s.db.Group("site_id").Find(&baselines)
	if len(baselines) == 0 {
		return "防篡改检查完成：尚未建立任何站点基线，跳过", nil
	}

	siteSvc := sitesvc.NewService(s.db)
	var reports []string
	for _, b := range baselines {
		diffs, err := siteSvc.VerifyTamper(b.SiteID)
		if err != nil {
			reports = append(reports, fmt.Sprintf("站点 #%d 校验失败: %v", b.SiteID, err))
			continue
		}
		if len(diffs) > 0 {
			lines := make([]string, 0, len(diffs))
			for _, d := range diffs {
				lines = append(lines, fmt.Sprintf("  %s (%s)", d.Path, d.Status))
			}
			reports = append(reports, fmt.Sprintf("站点 #%d 发现 %d 处变更:\n%s", b.SiteID, len(diffs), strings.Join(lines, "\n")))
		}
	}

	if len(reports) == 0 {
		return fmt.Sprintf("防篡改检查完成：%d 个站点基线全部一致", len(baselines)), nil
	}
	out := "防篡改检查发现异常:\n" + strings.Join(reports, "\n")
	_ = s.notify("网站防篡改告警", out)
	return out, nil
}

// execRecycleCleanup 回收站过期清理
func (s *Service) execRecycleCleanup() (string, error) {
	// 与文件 API 一致：生产环境以 / 为虚拟根，Windows 开发环境由调用方决定
	m, err := filesvc.NewManager("/")
	if err != nil {
		return "", err
	}
	n, err := m.PurgeExpired(s.db)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("回收站清理完成：清理 %d 项过期文件（保留 7 天）", n), nil
}

// EnsureDefaultTasks 启动时补齐内置计划任务（按 类型+名称 判重，用户删除后不会重建）
func EnsureDefaultTasks(db *gorm.DB) {
	defaults := []model.CronTask{
		{Name: "SSL 证书到期检查", Type: "ssl_check", Schedule: "0 9 * * *", Status: "enabled",
			Remark: "每日检查站点证书，30/15/7 天梯度邮件提醒"},
		{Name: "Docker 资源清理", Type: "docker_cleanup", Schedule: "0 3 * * 0", Status: "enabled",
			Remark: "每周清理停止容器、悬空镜像与构建缓存"},
		{Name: "回收站过期清理", Type: "recycle_cleanup", Schedule: "0 4 * * *", Status: "enabled",
			Remark: "清理 7 天前删除的文件"},
		{Name: "网站防篡改检查", Type: "tamper_check", Schedule: "0 10 * * *", Status: "enabled",
			Remark: "校验已建基线站点的文件完整性，发现变更即告警"},
		{Name: "站点病毒扫描", Type: "virus_scan", Schedule: "0 4 * * 0", Target: "/var/www", Status: "disabled",
			Remark: "ClamAV 扫描站点目录（需安装 clamscan，确认路径后启用）"},
	}
	for _, t := range defaults {
		var count int64
		db.Model(&model.CronTask{}).Where("type = ? AND name = ?", t.Type, t.Name).Count(&count)
		if count == 0 {
			db.Create(&t)
		}
	}
}

// notify 向所有启用的通知渠道广播告警（全部失败时仅返回错误不阻断任务）
func (s *Service) notify(title, content string) error {
	return alertsvc.NewNotifierManager(s.db).SendToAll(title, content)
}
