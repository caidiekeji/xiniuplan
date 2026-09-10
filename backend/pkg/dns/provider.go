// Package dns 提供云厂商 DNS API 的统一抽象接口，
// 目前支持阿里云、腾讯云 DNSPod、Cloudflare 三家提供商。
package dns

// DNSProvider DNS 提供商统一接口，所有云厂商实现均需满足该接口。
type DNSProvider interface {
	// Name 返回提供商标识名称，如 aliyun / tencent / cloudflare。
	Name() string
	// ListDomains 列出该账号下所有域名。
	ListDomains() ([]Domain, error)
	// ListRecords 列出指定域名下的所有 DNS 解析记录。
	ListRecords(domain string) ([]DNSRecord, error)
	// AddRecord 添加一条 DNS 解析记录。
	AddRecord(domain, recordType, name, value string, ttl int) error
	// DeleteRecord 删除一条 DNS 解析记录（按类型和名称匹配）。
	DeleteRecord(domain, recordType, name string) error
	// GetRecord 获取指定域名下指定类型和名称的单条解析记录。
	GetRecord(domain, recordType, name string) (*DNSRecord, error)
}

// Domain 域名信息
type Domain struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// DNSRecord DNS 解析记录
type DNSRecord struct {
	Type  string `json:"type"`  // A, AAAA, CNAME, TXT, MX
	Name  string `json:"name"`  // 主机记录，如 www、@、_acme-challenge
	Value string `json:"value"` // 记录值
	TTL   int    `json:"ttl"`   // 生存时间（秒）
	ID    string `json:"id"`    // 提供商侧记录 ID
}
