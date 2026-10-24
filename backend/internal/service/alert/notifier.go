package alert

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// Notifier 通知发送器接口
type Notifier interface {
	Send(channel model.AlertChannel, title, content string) error
}

// NotifierManager 通知器管理器，按渠道类型分发
type NotifierManager struct {
	db        *gorm.DB
	notifiers map[string]Notifier
}

// NewNotifierManager 创建通知器管理器
func NewNotifierManager(db *gorm.DB) *NotifierManager {
	m := &NotifierManager{
		db:        db,
		notifiers: make(map[string]Notifier),
	}
	m.notifiers["email"] = &EmailNotifier{db: db}
	m.notifiers["dingtalk"] = &DingTalkNotifier{}
	m.notifiers["wechat"] = &WeChatNotifier{}
	m.notifiers["feishu"] = &FeishuNotifier{}
	m.notifiers["telegram"] = &TelegramNotifier{}
	m.notifiers["webhook"] = &WebhookNotifier{}
	return m
}

// Send 通过指定渠道发送通知
func (m *NotifierManager) Send(channel model.AlertChannel, title, content string) error {
	n, ok := m.notifiers[channel.Type]
	if !ok {
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
	return n.Send(channel, title, content)
}

// SendToAll 向所有启用的渠道广播（拨测/防篡改/部署结果等场景通用）
func (m *NotifierManager) SendToAll(title, content string) error {
	var channels []model.AlertChannel
	if err := m.db.Where("enabled = ?", true).Find(&channels).Error; err != nil {
		return err
	}
	if len(channels) == 0 {
		return fmt.Errorf("无可用通知渠道")
	}
	ids := make([]string, 0, len(channels))
	for _, ch := range channels {
		ids = append(ids, strconv.Itoa(int(ch.ID)))
	}
	return m.SendByIDs(ids, title, content)
}

// SendByIDs 通过渠道ID列表批量发送通知，返回错误汇总
func (m *NotifierManager) SendByIDs(channelIDs []string, title, content string) error {
	if len(channelIDs) == 0 {
		return nil
	}
	var channels []model.AlertChannel
	if err := m.db.Where("id IN ? AND enabled = ?", channelIDs, true).Find(&channels).Error; err != nil {
		return fmt.Errorf("query channels: %w", err)
	}
	var errs []string
	for _, ch := range channels {
		if err := m.Send(ch, title, content); err != nil {
			slog.Warn("[alert] 渠道发送失败", "name", ch.Name, "type", ch.Type, "error", err)
			errs = append(errs, fmt.Sprintf("%s: %v", ch.Name, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// httpClient 统一 5s 超时的 HTTP 客户端
var httpClient = &http.Client{Timeout: 5 * time.Second}

// postJSON 发送 JSON POST 请求
func postJSON(rawURL string, body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// ---------------------------------------------------------------------------
// EmailNotifier 邮件通知器
// ---------------------------------------------------------------------------

// EmailNotifier 邮件通知器，从 Setting 表读取 SMTP 配置
type EmailNotifier struct {
	db *gorm.DB
}

func (e *EmailNotifier) Send(channel model.AlertChannel, title, content string) error {
	cfg, err := e.getSMTPConfig()
	if err != nil {
		return err
	}
	if cfg["smtp_host"] == "" || cfg["smtp_user"] == "" {
		return fmt.Errorf("SMTP 配置不完整")
	}

	// 收件人：渠道 config.to 或 config.recipients
	to := e.getRecipients(channel)
	if len(to) == 0 {
		return fmt.Errorf("邮件渠道未配置收件人")
	}

	port := cfg["smtp_port"]
	if port == "" {
		port = "587"
	}
	addr := cfg["smtp_host"] + ":" + port

	auth := smtp.PlainAuth("", cfg["smtp_user"], cfg["smtp_pass"], cfg["smtp_host"])

	from := cfg["smtp_from"]
	if from == "" {
		from = cfg["smtp_user"]
	}

	msg := buildEmailMessage(from, to, title, content)
	if err := smtp.SendMail(addr, auth, from, to, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func (e *EmailNotifier) getSMTPConfig() (map[string]string, error) {
	keys := []string{"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from"}
	var settings []model.Setting
	if err := e.db.Where("`key` IN ?", keys).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("query smtp settings: %w", err)
	}
	result := make(map[string]string, len(keys))
	for _, s := range settings {
		result[s.Key] = s.Value
	}
	return result, nil
}

func (e *EmailNotifier) getRecipients(channel model.AlertChannel) []string {
	if v, ok := channel.Config["to"]; ok {
		if s, ok := v.(string); ok && s != "" {
			return []string{s}
		}
	}
	if v, ok := channel.Config["recipients"]; ok {
		// 前端保存的是逗号分隔字符串，数组与字符串都要认
		if s, ok := v.(string); ok && s != "" {
			result := make([]string, 0)
			for _, part := range strings.Split(s, ",") {
				if p := strings.TrimSpace(part); p != "" {
					result = append(result, p)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
		if arr, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok && s != "" {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

func buildEmailMessage(from string, to []string, subject, body string) []byte {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(k + ": " + v + "\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return []byte(msg.String())
}

// ---------------------------------------------------------------------------
// DingTalkNotifier 钉钉通知器
// ---------------------------------------------------------------------------

// DingTalkNotifier 钉钉机器人通知器，支持加签
type DingTalkNotifier struct{}

func (d *DingTalkNotifier) Send(channel model.AlertChannel, title, content string) error {
	webhook, _ := channel.Config["webhook"].(string)
	if webhook == "" {
		return fmt.Errorf("钉钉渠道未配置 webhook")
	}

	// 支持加签
	if secret, ok := channel.Config["secret"].(string); ok && secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		stringToSign := timestamp + "\n" + secret
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(stringToSign))
		sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
		sign = url.QueryEscape(sign)
		sep := "&"
		if !strings.Contains(webhook, "?") {
			sep = "?"
		}
		webhook = webhook + sep + "timestamp=" + timestamp + "&sign=" + sign
	}

	fullContent := title + "\n" + content
	body := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": fullContent},
	}
	return postJSON(webhook, body)
}

// ---------------------------------------------------------------------------
// WeChatNotifier 企业微信通知器
// ---------------------------------------------------------------------------

// WeChatNotifier 企业微信机器人通知器
type WeChatNotifier struct{}

func (w *WeChatNotifier) Send(channel model.AlertChannel, title, content string) error {
	webhook, _ := channel.Config["webhook"].(string)
	if webhook == "" {
		return fmt.Errorf("企微渠道未配置 webhook")
	}
	fullContent := title + "\n" + content
	body := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": fullContent},
	}
	return postJSON(webhook, body)
}

// ---------------------------------------------------------------------------
// FeishuNotifier 飞书通知器
// ---------------------------------------------------------------------------

// FeishuNotifier 飞书机器人通知器
type FeishuNotifier struct{}

func (f *FeishuNotifier) Send(channel model.AlertChannel, title, content string) error {
	webhook, _ := channel.Config["webhook"].(string)
	if webhook == "" {
		return fmt.Errorf("飞书渠道未配置 webhook")
	}
	fullContent := title + "\n" + content
	body := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": fullContent},
	}
	return postJSON(webhook, body)
}

// ---------------------------------------------------------------------------
// TelegramNotifier Telegram 通知器
// ---------------------------------------------------------------------------

// TelegramNotifier Telegram Bot 通知器
type TelegramNotifier struct{}

func (t *TelegramNotifier) Send(channel model.AlertChannel, title, content string) error {
	token, _ := channel.Config["bot_token"].(string)
	chatID, _ := channel.Config["chat_id"].(string)
	if token == "" || chatID == "" {
		return fmt.Errorf("Telegram 渠道未配置 bot_token 或 chat_id")
	}
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	fullText := title + "\n" + content
	body := map[string]interface{}{
		"chat_id": chatID,
		"text":    fullText,
	}
	return postJSON(apiURL, body)
}

// ---------------------------------------------------------------------------
// WebhookNotifier 自定义 Webhook 通知器
// ---------------------------------------------------------------------------

// WebhookNotifier 自定义 Webhook 通知器
type WebhookNotifier struct{}

func (w *WebhookNotifier) Send(channel model.AlertChannel, title, content string) error {
	rawURL, _ := channel.Config["url"].(string)
	if rawURL == "" {
		return fmt.Errorf("Webhook 渠道未配置 url")
	}

	// 允许自定义请求体模板，默认发送标准结构
	body := map[string]interface{}{
		"title":   title,
		"content": content,
		"time":    time.Now().Format(time.RFC3339),
	}

	// 如果配置了自定义 body 字段，合并
	if custom, ok := channel.Config["body"].(map[string]interface{}); ok {
		for k, v := range custom {
			body[k] = v
		}
	}

	return postJSON(rawURL, body)
}
