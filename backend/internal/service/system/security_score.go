package system

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"cloudpivot/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// --- 安全评分（D-007） ---

// ScoreCheck 评分检查项
type ScoreCheck struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Score  int    `json:"score"`  // 该项得分
	Max    int    `json:"max"`    // 满分
	Status string `json:"status"` // good | warn | bad
	Advice string `json:"advice"`
}

// SecurityScore 安全评分结果
type SecurityScore struct {
	Total  int          `json:"total"`
	Grade  string       `json:"grade"` // A | B | C | D
	Checks []ScoreCheck `json:"checks"`
}

// GetSecurityScore 综合评估系统安全状态
func (s *Service) GetSecurityScore() (*SecurityScore, error) {
	score := &SecurityScore{Checks: make([]ScoreCheck, 0)}

	addCheck := func(key, label string, got, max int, advice string) {
		status := "good"
		switch {
		case got >= max:
			status = "good"
		case got > 0:
			status = "warn"
		default:
			status = "bad"
		}
		score.Checks = append(score.Checks, ScoreCheck{
			Key: key, Label: label, Score: got, Max: max, Status: status, Advice: advice,
		})
	}

	// 1. 默认弱密码检查（-25 分档，弱密码直接 0 分）
	weakAdmin := false
	var admin model.User
	if err := s.db.Where("username = ?", "admin").First(&admin).Error; err == nil {
		if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte("admin123")) == nil {
			weakAdmin = true
		}
	}
	if weakAdmin {
		addCheck("default_password", "默认密码", 0, 25,
			"admin 仍在使用默认密码 admin123，请立即修改！")
	} else {
		addCheck("default_password", "默认密码", 25, 25, "")
	}

	// 2. 安全入口
	settings, _ := s.GetSettings()
	if entry := settings["security_entry"]; strings.TrimSpace(entry) != "" {
		addCheck("security_entry", "安全入口", 10, 10, "")
	} else {
		addCheck("security_entry", "安全入口", 0, 10,
			"未配置安全入口路径，面板地址可被扫描发现。建议在系统设置中开启。")
	}

	// 3. TOTP 二次验证
	var totpCount int64
	s.db.Model(&model.User{}).Where("totp_secret != '' AND status = 'active'").Count(&totpCount)
	if totpCount > 0 {
		addCheck("totp", "二次验证", 15, 15, "")
	} else {
		addCheck("totp", "二次验证", 0, 15,
			"没有任何用户启用 TOTP 二次验证，建议为管理员账号开启。")
	}

	// 4. 防火墙规则
	var fwCount int64
	s.db.Model(&model.FirewallRule{}).Where("status = 'enabled'").Count(&fwCount)
	if fwCount > 0 {
		addCheck("firewall", "防火墙规则", 10, 10, "")
	} else {
		addCheck("firewall", "防火墙规则", 0, 10,
			"防火墙未配置任何启用规则，建议至少限制高危端口。")
	}

	// 5. SSH 端口（22 → warn）
	sshPort := getSSHConfiguredPort()
	if sshPort != "" && sshPort != "22" {
		addCheck("ssh_port", "SSH 端口", 15, 15, "")
	} else {
		addCheck("ssh_port", "SSH 端口", 0, 15,
			"SSH 使用默认 22 端口，容易被暴力扫描。建议修改为非标准端口。")
	}

	// 6. SSH 密码登录
	if sshConfig := readSSHPasswordAuth(); sshConfig == "no" {
		addCheck("ssh_password_auth", "SSH 密码登录", 15, 15, "")
	} else {
		addCheck("ssh_password_auth", "SSH 密码登录", 5, 15,
			"SSH 仍允许密码登录，建议配置密钥后禁用密码登录。")
	}

	// 7. 站点 SSL 覆盖率
	var totalSites, sslSites int64
	s.db.Model(&model.Site{}).Count(&totalSites)
	s.db.Model(&model.Site{}).Where("ssl_enabled = ?", true).Count(&sslSites)
	if totalSites == 0 {
		addCheck("ssl", "站点 HTTPS", 10, 10, "")
	} else {
		ratio := int(float64(sslSites) / float64(totalSites) * 10)
		if ratio >= 10 {
			addCheck("ssl", "站点 HTTPS", 10, 10, "")
		} else {
			addCheck("ssl", "站点 HTTPS", ratio, 10,
				fmt.Sprintf("%d 个站点中仅 %d 个启用 HTTPS，建议为全部站点申请 SSL 证书。", totalSites, sslSites))
		}
	}

	for _, chk := range score.Checks {
		score.Total += chk.Score
	}
	switch {
	case score.Total >= 85:
		score.Grade = "A"
	case score.Total >= 70:
		score.Grade = "B"
	case score.Total >= 50:
		score.Grade = "C"
	default:
		score.Grade = "D"
	}
	return score, nil
}

// getSSHConfiguredPort 从 sshd_config 读取 SSH 端口
func getSSHConfiguredPort() string {
	for _, path := range []string{"/etc/ssh/sshd_config", "./runtime/sshd_config"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		re := regexp.MustCompile(`(?m)^\s*Port\s+(\d+)`)
		if m := re.FindStringSubmatch(string(data)); m != nil {
			return m[1]
		}
		return "22" // 配置文件存在但未显式指定端口 → 默认 22
	}
	return "22"
}

// readSSHPasswordAuth 读取 PasswordAuthentication 配置
func readSSHPasswordAuth() string {
	for _, path := range []string{"/etc/ssh/sshd_config", "./runtime/sshd_config"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		re := regexp.MustCompile(`(?m)^\s*PasswordAuthentication\s+(\w+)`)
		if m := re.FindStringSubmatch(string(data)); m != nil {
			return strings.ToLower(m[1])
		}
		return "yes" // 默认允许密码登录
	}
	return "yes"
}
