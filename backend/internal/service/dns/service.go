// Package dns 提供 DNS 账号管理与云厂商 DNS 操作的服务层。
package dns

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	dnsprovider "cloudpivot/pkg/dns"

	"gorm.io/gorm"
)

// Service DNS 管理服务。
type Service struct {
	db *gorm.DB
}

// NewService 创建 DNS 管理服务实例。
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// encodeCredentials 将凭证 map 序列化为 JSON 后做 base64 编码。
func encodeCredentials(creds map[string]interface{}) (string, error) {
	if len(creds) == 0 {
		return "", fmt.Errorf("credentials cannot be empty")
	}
	raw, err := json.Marshal(creds)
	if err != nil {
		return "", fmt.Errorf("marshal credentials: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

// decodeCredentials 解码 base64 编码的凭证 JSON。
func decodeCredentials(encoded string) (map[string]interface{}, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode credentials: %w", err)
	}
	var creds map[string]interface{}
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, fmt.Errorf("unmarshal credentials: %w", err)
	}
	return creds, nil
}

// getMapString 从 map 中安全获取字符串值。
func getMapString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// newProvider 根据账号信息创建对应的 DNS 提供商实例。
func newProvider(account *model.DNSAccount) (dnsprovider.DNSProvider, error) {
	creds, err := decodeCredentials(account.Credentials)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials for account %d: %w", account.ID, err)
	}

	switch account.ProviderType {
	case "aliyun":
		accessKeyID := getMapString(creds, "access_key_id")
		accessKeySecret := getMapString(creds, "access_key_secret")
		if accessKeyID == "" || accessKeySecret == "" {
			return nil, fmt.Errorf("aliyun credentials require access_key_id and access_key_secret")
		}
		return dnsprovider.NewAliyunProvider(accessKeyID, accessKeySecret), nil

	case "tencent":
		secretID := getMapString(creds, "secret_id")
		secretKey := getMapString(creds, "secret_key")
		if secretID == "" || secretKey == "" {
			return nil, fmt.Errorf("tencent credentials require secret_id and secret_key")
		}
		return dnsprovider.NewTencentProvider(secretID, secretKey), nil

	case "cloudflare":
		apiToken := getMapString(creds, "api_token")
		if apiToken == "" {
			return nil, fmt.Errorf("cloudflare credentials require api_token")
		}
		return dnsprovider.NewCloudflareProvider(apiToken), nil

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", account.ProviderType)
	}
}

// toResponse 将模型转换为响应 DTO（不包含 credentials）。
func toResponse(a *model.DNSAccount) dto.DNSAccountResponse {
	return dto.DNSAccountResponse{
		ID:           a.ID,
		Name:         a.Name,
		ProviderType: a.ProviderType,
		Enabled:      a.Enabled,
		Domains:      a.Domains,
		CreatedAt:    a.CreatedAt,
	}
}

// getAccountByID 根据 ID 获取账号，不存在则返回错误。
func (s *Service) getAccountByID(id uint) (*model.DNSAccount, error) {
	var account model.DNSAccount
	if err := s.db.First(&account, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("dns account not found: %d", id)
		}
		return nil, fmt.Errorf("query dns account: %w", err)
	}
	return &account, nil
}

// ListAccounts 列出所有 DNS 账号（不返回 credentials）。
func (s *Service) ListAccounts() ([]dto.DNSAccountResponse, error) {
	accounts := make([]model.DNSAccount, 0)
	if err := s.db.Order("id DESC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("query dns accounts: %w", err)
	}

	result := make([]dto.DNSAccountResponse, 0, len(accounts))
	for i := range accounts {
		result = append(result, toResponse(&accounts[i]))
	}
	return result, nil
}

// CreateAccount 创建 DNS 账号，credentials 以 base64 编码存储。
func (s *Service) CreateAccount(req dto.CreateDNSAccountRequest) (*dto.DNSAccountResponse, error) {
	encodedCreds, err := encodeCredentials(req.Credentials)
	if err != nil {
		return nil, err
	}

	account := model.DNSAccount{
		Name:         req.Name,
		ProviderType: req.ProviderType,
		Credentials:  encodedCreds,
		Enabled:      true,
	}

	if err := s.db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("create dns account: %w", err)
	}

	resp := toResponse(&account)
	return &resp, nil
}

// UpdateAccount 更新 DNS 账号信息。
func (s *Service) UpdateAccount(id uint, req dto.UpdateDNSAccountRequest) (*dto.DNSAccountResponse, error) {
	account, err := s.getAccountByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Enabled != nil {
		account.Enabled = *req.Enabled
	}
	if req.Credentials != nil {
		encodedCreds, err := encodeCredentials(req.Credentials)
		if err != nil {
			return nil, err
		}
		account.Credentials = encodedCreds
	}

	if err := s.db.Save(account).Error; err != nil {
		return nil, fmt.Errorf("update dns account: %w", err)
	}

	resp := toResponse(account)
	return &resp, nil
}

// DeleteAccount 删除 DNS 账号。
func (s *Service) DeleteAccount(id uint) error {
	account, err := s.getAccountByID(id)
	if err != nil {
		return err
	}
	if err := s.db.Delete(account).Error; err != nil {
		return fmt.Errorf("delete dns account: %w", err)
	}
	return nil
}

// GetDomains 调用对应提供商获取域名列表，并缓存到 account.domains。
func (s *Service) GetDomains(accountID uint) ([]dnsprovider.Domain, error) {
	account, err := s.getAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	if !account.Enabled {
		return nil, fmt.Errorf("dns account is disabled: %d", accountID)
	}

	provider, err := newProvider(account)
	if err != nil {
		return nil, err
	}

	domains, err := provider.ListDomains()
	if err != nil {
		return nil, fmt.Errorf("list domains via %s: %w", account.ProviderType, err)
	}

	// 缓存域名列表到数据库
	domainsJSON, _ := json.Marshal(domains)
	account.Domains = string(domainsJSON)
	if err := s.db.Model(account).Update("domains", account.Domains).Error; err != nil {
		// 缓存失败不影响主流程，仅记录
		fmt.Printf("[dns] warn: cache domains failed: %v\n", err)
	}

	return domains, nil
}

// GetRecords 获取指定域名下的所有解析记录。
func (s *Service) GetRecords(accountID uint, domain string) ([]dnsprovider.DNSRecord, error) {
	account, err := s.getAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	if !account.Enabled {
		return nil, fmt.Errorf("dns account is disabled: %d", accountID)
	}

	provider, err := newProvider(account)
	if err != nil {
		return nil, err
	}

	records, err := provider.ListRecords(domain)
	if err != nil {
		return nil, fmt.Errorf("list records via %s: %w", account.ProviderType, err)
	}
	return records, nil
}

// AddRecord 添加一条 DNS 解析记录。
func (s *Service) AddRecord(accountID uint, domain, recordType, name, value string, ttl int) error {
	account, err := s.getAccountByID(accountID)
	if err != nil {
		return err
	}
	if !account.Enabled {
		return fmt.Errorf("dns account is disabled: %d", accountID)
	}

	provider, err := newProvider(account)
	if err != nil {
		return err
	}

	if err := provider.AddRecord(domain, recordType, name, value, ttl); err != nil {
		return fmt.Errorf("add record via %s: %w", account.ProviderType, err)
	}
	return nil
}

// DeleteRecord 删除一条 DNS 解析记录。
func (s *Service) DeleteRecord(accountID uint, domain, recordType, name string) error {
	account, err := s.getAccountByID(accountID)
	if err != nil {
		return err
	}
	if !account.Enabled {
		return fmt.Errorf("dns account is disabled: %d", accountID)
	}

	provider, err := newProvider(account)
	if err != nil {
		return err
	}

	if err := provider.DeleteRecord(domain, recordType, name); err != nil {
		return fmt.Errorf("delete record via %s: %w", account.ProviderType, err)
	}
	return nil
}

// AddTXTRecordAndWait 用于 SSL DNS-01 验证：添加 TXT 记录并轮询确认生效。
// 轮询间隔 2 秒，最长等待 30 秒。
func (s *Service) AddTXTRecordAndWait(accountID uint, domain, name, value string) error {
	if err := s.AddRecord(accountID, domain, "TXT", name, value, 60); err != nil {
		return err
	}

	// 轮询确认记录已生效
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		rec, err := s.getRecordInternal(accountID, domain, "TXT", name)
		if err == nil && rec != nil && rec.Value == value {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	// 超时不返回错误，记录可能已添加但尚未同步
	fmt.Printf("[dns] warn: TXT record not confirmed within 30s (domain=%s name=%s)\n", domain, name)
	return nil
}

// CleanupTXTRecord SSL DNS-01 验证完成后清理 TXT 记录。
func (s *Service) CleanupTXTRecord(accountID uint, domain, name string) error {
	return s.DeleteRecord(accountID, domain, "TXT", name)
}

// getRecordInternal 内部方法：获取单条记录（供 AddTXTRecordAndWait 轮询使用）。
func (s *Service) getRecordInternal(accountID uint, domain, recordType, name string) (*dnsprovider.DNSRecord, error) {
	account, err := s.getAccountByID(accountID)
	if err != nil {
		return nil, err
	}
	provider, err := newProvider(account)
	if err != nil {
		return nil, err
	}
	return provider.GetRecord(domain, recordType, name)
}
