package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// resetTokenTTL 重置令牌有效期
const resetTokenTTL = 30 * time.Minute

// RequestPasswordReset 忘记密码：校验用户名+邮箱匹配后发送重置邮件。
// 无论匹配与否都返回 nil（防账号枚举），由调用方统一提示"如果信息正确将收到邮件"。
func (s *Service) RequestPasswordReset(username, email, panelOrigin string) error {
	var user model.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil // 用户不存在：静默成功
	}
	if user.Email == "" || !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(email)) {
		return nil // 邮箱不匹配：静默成功
	}

	// 生成明文 token 与哈希（只存哈希，防库泄露后直接用）
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	// 旧的未用令牌作废
	s.db.Where("username = ? AND used_at IS NULL", user.Username).
		Delete(&model.PasswordResetToken{})

	record := model.PasswordResetToken{
		Username:  user.Username,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(resetTokenTTL),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return fmt.Errorf("save reset token: %w", err)
	}

	// 重置链接：{origin}{entry}/reset-password?token=xxx
	link := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(panelOrigin, "/"), token)

	// 异步发送邮件，避免 SMTP 超时阻塞 HTTP 请求
	cfg, cfgErr := s.getSMTPConfig()
	go func(username, email, link string, cfg map[string]string, cfgErr error) {
		if cfgErr != nil {
			fmt.Printf("[忘记密码] 读取 SMTP 配置失败: %v\n", cfgErr)
			return
		}
		if err := sendResetMail(cfg, username, email, link); err != nil {
			fmt.Printf("[忘记密码] 发送重置邮件失败: user=%s err=%v\n", username, err)
		}
	}(user.Username, user.Email, link, cfg, cfgErr)

	return nil
}

// ResetPassword 使用令牌重置密码
func (s *Service) ResetPassword(token, newPassword string) error {
	if len(token) != 64 {
		return errors.New("重置链接无效或已过期")
	}
	if len(newPassword) < 6 {
		return errors.New("新密码至少 6 位")
	}

	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	var record model.PasswordResetToken
	if err := s.db.Where("token_hash = ?", tokenHash).First(&record).Error; err != nil {
		return errors.New("重置链接无效或已过期")
	}
	if record.UsedAt != nil {
		return errors.New("重置链接已使用")
	}
	if time.Now().After(record.ExpiresAt) {
		return errors.New("重置链接已过期，请重新申请")
	}

	var user model.User
	if err := s.db.Where("username = ?", record.Username).First(&user).Error; err != nil {
		return errors.New("用户不存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx := s.db.Begin()
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"password_hash":        string(hash),
		"must_change_password": false,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update password: %w", err)
	}
	now := time.Now()
	if err := tx.Model(&record).Update("used_at", &now).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("mark token used: %w", err)
	}
	return tx.Commit().Error
}

// getSMTPConfig 从设置表读取 SMTP 配置
func (s *Service) getSMTPConfig() (map[string]string, error) {
	var settings []model.Setting
	if err := s.db.Find(&settings).Error; err != nil {
		return nil, err
	}
	cfg := make(map[string]string, len(settings))
	for _, item := range settings {
		cfg[item.Key] = item.Value
	}
	return cfg, nil
}

// sendResetMail 发送密码重置邮件（cfg 来自系统设置表：smtp_host/port/user/pass/from）
func sendResetMail(cfg map[string]string, username, to, link string) error {
	host := cfg["smtp_host"]
	port := cfg["smtp_port"]
	user := cfg["smtp_user"]
	pass := cfg["smtp_pass"]
	from := cfg["smtp_from"]
	if from == "" {
		from = user
	}
	if host == "" || user == "" || to == "" {
		return errors.New("SMTP 未配置或用户未绑定邮箱")
	}
	if port == "" {
		port = "465"
	}

	subject := "云枢面板 - 密码重置"
	body := fmt.Sprintf(
		"您好，%s：\r\n\r\n"+
			"我们收到了您重置面板密码的请求。请点击以下链接设置新密码（30 分钟内有效）：\r\n\r\n"+
			"%s\r\n\r\n"+
			"如果这不是您本人的操作，请忽略本邮件，您的密码不会被修改。\r\n\r\n"+
			"—— 云枢面板",
		username, link,
	)

	msg := buildMailMessage(from, to, subject, body)
	addr := host + ":" + port

	auth := smtp.PlainAuth("", user, pass, host)
	// 465 端口走 TLS 直连，其余走 STARTTLS/明文回退
	if port == "465" {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		defer conn.Close()
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}
		if err = client.Mail(from); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err = w.Write(msg); err != nil {
			return err
		}
		if err = w.Close(); err != nil {
			return err
		}
		return client.Quit()
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// buildMailMessage 构造 MIME 邮件正文
func buildMailMessage(from, to, subject, body string) []byte {
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, to, subject,
	)
	return []byte(headers + body)
}
