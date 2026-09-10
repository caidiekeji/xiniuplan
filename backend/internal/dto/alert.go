package dto

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name       string   `json:"name" binding:"required"`
	MetricType string   `json:"metric_type" binding:"required,oneof=cpu memory disk network"`
	Condition  string   `json:"condition" binding:"required,oneof=gt lt eq"`
	Threshold  float64  `json:"threshold" binding:"required"`
	Duration   int      `json:"duration"`
	Enabled    bool     `json:"enabled"`
	Channels   []string `json:"channels"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name       string   `json:"name"`
	MetricType string   `json:"metric_type" binding:"omitempty,oneof=cpu memory disk network"`
	Condition  string   `json:"condition" binding:"omitempty,oneof=gt lt eq"`
	Threshold  float64  `json:"threshold"`
	Duration   int      `json:"duration"`
	Enabled    *bool    `json:"enabled"`
	Channels   []string `json:"channels"`
}

// AlertRuleResponse 告警规则响应
type AlertRuleResponse struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	MetricType string   `json:"metric_type"`
	Condition  string   `json:"condition"`
	Threshold  float64  `json:"threshold"`
	Duration   int      `json:"duration"`
	Enabled    bool     `json:"enabled"`
	Channels   []string `json:"channels"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// CreateAlertChannelRequest 创建告警渠道请求
type CreateAlertChannelRequest struct {
	Name    string                 `json:"name" binding:"required"`
	Type    string                 `json:"type" binding:"required,oneof=email dingtalk wechat feishu telegram webhook"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// UpdateAlertChannelRequest 更新告警渠道请求
type UpdateAlertChannelRequest struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type" binding:"omitempty,oneof=email dingtalk wechat feishu telegram webhook"`
	Config  map[string]interface{} `json:"config"`
	Enabled *bool                  `json:"enabled"`
}

// AlertChannelResponse 告警渠道响应
type AlertChannelResponse struct {
	ID        uint                   `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt string                 `json:"created_at"`
}

// AlertHistoryResponse 告警历史响应
type AlertHistoryResponse struct {
	ID                 uint    `json:"id"`
	RuleID             uint    `json:"rule_id"`
	RuleName           string  `json:"rule_name"`
	MetricType         string  `json:"metric_type"`
	TriggerValue       float64 `json:"trigger_value"`
	Threshold          float64 `json:"threshold"`
	Status             string  `json:"status"`
	TriggeredAt        string  `json:"triggered_at"`
	RecoveredAt        string  `json:"recovered_at"`
	NotificationStatus string  `json:"notification_status"`
	NotificationError  string  `json:"notification_error"`
	CreatedAt          string  `json:"created_at"`
}

// TestChannelRequest 测试渠道请求
type TestChannelRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
