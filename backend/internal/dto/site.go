package dto

// CreateSiteRequest 创建网站请求
type CreateSiteRequest struct {
	PrimaryDomain string   `json:"primary_domain" binding:"required"`
	Domains       []string `json:"domains"`
	Type          string   `json:"type" binding:"required,oneof=static runtime reverse_proxy"`
	Runtime       string   `json:"runtime"`
	RootDir       string   `json:"root_dir"`
	ProxyTarget   string   `json:"proxy_target"`
	Remark        string   `json:"remark"`
}

// UpdateSiteRequest 更新网站请求（所有字段可选）
type UpdateSiteRequest struct {
	PrimaryDomain *string  `json:"primary_domain"`
	Domains       []string `json:"domains"`
	Type          *string  `json:"type"`
	Runtime       *string  `json:"runtime"`
	RootDir       *string  `json:"root_dir"`
	ProxyTarget   *string  `json:"proxy_target"`
	Remark        *string  `json:"remark"`
	// WAF 基础版（roadmap #10）
	WafEnabled  *bool   `json:"waf_enabled"`
	CCRpm       *int    `json:"cc_rpm"`
	UABlacklist *string `json:"ua_blacklist"`
}

// SiteSSLInfo SSL 信息
type SiteSSLInfo struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	CertPath  string `json:"cert_path"`
	KeyPath   string `json:"key_path"`
	ExpireAt  string `json:"expire_at"`
	AutoRenew bool   `json:"auto_renew"`
	Issuer    string `json:"issuer"`
	DaysLeft  int    `json:"days_left"`
}

// SiteResponse 网站响应
type SiteResponse struct {
	ID            uint        `json:"id"`
	PrimaryDomain string      `json:"primary_domain"`
	Domains       []string    `json:"domains"`
	Type          string      `json:"type"`
	Runtime       string      `json:"runtime"`
	RootDir       string      `json:"root_dir"`
	NginxConf     string      `json:"nginx_conf"`
	Status        string      `json:"status"`
	ContainerID   string      `json:"container_id"`
	Remark        string      `json:"remark"`
	SSL           SiteSSLInfo `json:"ssl"`
	WafEnabled    bool        `json:"waf_enabled"`
	CCRpm         int         `json:"cc_rpm"`
	UABlacklist   string      `json:"ua_blacklist"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

// NginxConfigRequest Nginx 配置更新请求
type NginxConfigRequest struct {
	Content string `json:"content" binding:"required"`
}

// NginxConfigResponse Nginx 配置响应
type NginxConfigResponse struct {
	Domain  string `json:"domain"`
	Content string `json:"content"`
}

// SSLUploadRequest 手动上传证书请求
type SSLUploadRequest struct {
	Cert string `json:"cert" binding:"required"`
	Key  string `json:"key" binding:"required"`
}

// SSLApplyRequest 申请 SSL 请求
type SSLApplyRequest struct {
	Provider string `json:"provider" binding:"required,oneof=manual letsencrypt"`
	Email    string `json:"email"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
}

// SSLInfoResponse SSL 信息响应
type SSLInfoResponse struct {
	SiteSSLInfo
}

// SiteLogResponse 日志响应
type SiteLogResponse struct {
	Type    string `json:"type"`
	Lines   int    `json:"lines"`
	Content string `json:"content"`
}
