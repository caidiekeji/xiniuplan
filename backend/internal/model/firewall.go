package model

import "time"

// FirewallRule 防火墙规则模型
type FirewallRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:16;not null;index" json:"type"`             // port | ip
	Direction string    `gorm:"size:16;not null" json:"direction"`              // inbound | outbound
	Protocol  string    `gorm:"size:16;not null" json:"protocol"`               // tcp | udp | icmp | all
	Port      int       `gorm:"default:0" json:"port"`                          // 端口号（type=port 时有效）
	IPAddress string    `gorm:"size:64" json:"ip_address"`                      // IP 地址/网段（type=ip 时有效）
	Action    string    `gorm:"size:16;not null" json:"action"`                 // accept | drop
	Status    string    `gorm:"size:16;not null;default:enabled" json:"status"` // enabled | disabled
	Remark    string    `gorm:"size:255" json:"remark"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

func (FirewallRule) TableName() string { return "firewall_rules" }
