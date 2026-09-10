package model

import "time"

// DNSAccount DNS 服务商账号模型，存储云厂商 API 凭证。
// credentials 字段以 base64 编码的 JSON 存储，不存明文。
// domains 字段缓存该账号下的域名列表（JSON 格式）。
type DNSAccount struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:128;not null;index" json:"name"`
	ProviderType string    `gorm:"size:32;not null;index" json:"provider_type"` // aliyun | tencent | cloudflare
	Credentials  string    `gorm:"type:text;not null" json:"-"`                 // base64 编码的 JSON 凭证，不对外返回
	Domains      string    `gorm:"type:text" json:"domains"`                    // 域名列表 JSON 缓存
	Enabled      bool      `gorm:"default:true" json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (DNSAccount) TableName() string { return "dns_accounts" }
