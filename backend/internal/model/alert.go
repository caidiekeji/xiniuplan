package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// JSONMap 自定义 JSON map 类型，用于渠道配置
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("JSONMap.Scan: unsupported type")
	}
	if len(b) == 0 {
		*m = JSONMap{}
		return nil
	}
	return json.Unmarshal(b, m)
}

// AlertRule 告警规则模型
type AlertRule struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	Name       string      `gorm:"size:128;not null" json:"name"`
	MetricType string      `gorm:"size:32;not null;index" json:"metric_type"` // cpu | memory | disk | network
	Condition  string      `gorm:"size:8;not null" json:"condition"`          // gt | lt | eq
	Threshold  float64     `gorm:"not null" json:"threshold"`
	Duration   int         `gorm:"not null;default:1" json:"duration"` // 持续分钟数
	Enabled    bool        `gorm:"default:true" json:"enabled"`
	Channels   StringSlice `gorm:"type:text;default:'[]'" json:"channels"` // 渠道ID列表
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

func (AlertRule) TableName() string { return "alert_rules" }

// AlertChannel 告警渠道模型
type AlertChannel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Type      string    `gorm:"size:32;not null;index" json:"type"` // email | dingtalk | wechat | feishu | telegram | webhook
	Config    JSONMap   `gorm:"type:text;default:'{}'" json:"config"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

func (AlertChannel) TableName() string { return "alert_channels" }

// AlertHistory 告警历史记录模型
type AlertHistory struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	RuleID             uint       `gorm:"not null;index" json:"rule_id"`
	RuleName           string     `gorm:"size:128;not null" json:"rule_name"`
	MetricType         string     `gorm:"size:32;not null" json:"metric_type"`
	TriggerValue       float64    `gorm:"not null" json:"trigger_value"`
	Threshold          float64    `gorm:"not null" json:"threshold"`
	Status             string     `gorm:"size:16;not null;default:triggered;index" json:"status"` // triggered | recovered
	TriggeredAt        time.Time  `json:"triggered_at"`
	RecoveredAt        *time.Time `json:"recovered_at"`
	NotificationStatus string     `gorm:"size:16;default:pending" json:"notification_status"` // pending | sent | failed
	NotificationError  string     `gorm:"type:text" json:"notification_error"`
	CreatedAt          time.Time  `json:"created_at"`
}

func (AlertHistory) TableName() string { return "alert_histories" }
