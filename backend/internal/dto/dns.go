package dto

import "time"

// CreateDNSAccountRequest 创建 DNS 账号请求。
// credentials 为各厂商所需的凭证字段（JSON 对象）：
//   - aliyun:     {"access_key_id": "...", "access_key_secret": "..."}
//   - tencent:    {"secret_id": "...", "secret_key": "..."}
//   - cloudflare: {"api_token": "..."}
type CreateDNSAccountRequest struct {
	Name         string                 `json:"name" binding:"required,max=128"`
	ProviderType string                 `json:"provider_type" binding:"required,oneof=aliyun tencent cloudflare"`
	Credentials  map[string]interface{} `json:"credentials" binding:"required"`
}

// UpdateDNSAccountRequest 更新 DNS 账号请求，所有字段可选。
type UpdateDNSAccountRequest struct {
	Name        *string                `json:"name,omitempty"`
	Credentials map[string]interface{} `json:"credentials,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
}

// DNSAccountResponse DNS 账号响应，不包含 credentials 凭证信息。
type DNSAccountResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"provider_type"`
	Enabled      bool      `json:"enabled"`
	Domains      string    `json:"domains"`
	CreatedAt    time.Time `json:"created_at"`
}

// DNSRecordResponse DNS 解析记录响应。
type DNSRecordResponse struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	ID    string `json:"id"`
}

// AddDNSRecordRequest 添加 DNS 记录请求。
type AddDNSRecordRequest struct {
	Type  string `json:"type" binding:"required,oneof=A AAAA CNAME TXT MX NS SRV PTR"`
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
	TTL   int    `json:"ttl"`
}

// DeleteDNSRecordRequest 删除 DNS 记录请求。
type DeleteDNSRecordRequest struct {
	Type string `json:"type" binding:"required,oneof=A AAAA CNAME TXT MX NS SRV PTR"`
	Name string `json:"name" binding:"required"`
}
