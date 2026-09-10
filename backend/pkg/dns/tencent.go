package dns

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	tencentEndpoint = "https://dnspod.tencentcloudapi.com/"
	tencentService  = "dnspod"
	tencentVersion  = "2021-03-23"
	tencentHost     = "dnspod.tencentcloudapi.com"
)

// TencentProvider 腾讯云 DNSPod 提供商实现，使用 API 3.0 TC3-HMAC-SHA256 签名。
type TencentProvider struct {
	secretID  string
	secretKey string
	client    *http.Client
}

// NewTencentProvider 创建腾讯云 DNSPod 提供商实例。
func NewTencentProvider(secretID, secretKey string) *TencentProvider {
	return &TencentProvider{
		secretID:  secretID,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *TencentProvider) Name() string { return "tencent" }

// sha256Hex 计算字符串的 SHA256 并返回十六进制字符串。
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// hmacSHA256 计算 HMAC-SHA256，返回原始字节。
func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

// tencentSign 计算腾讯云 API 3.0 TC3-HMAC-SHA256 签名，返回 Authorization 头值。
func (p *TencentProvider) tencentSign(action string, payload []byte, timestamp int64) string {
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	actionLower := strings.ToLower(action)

	// 1. 拼接规范请求串
	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\nx-tc-action:%s\n", tencentHost, actionLower)
	signedHeaders := "content-type;host;x-tc-action"
	hashedPayload := sha256Hex(string(payload))

	canonicalRequest := strings.Join([]string{
		"POST",
		"/",
		"",
		canonicalHeaders,
		signedHeaders,
		hashedPayload,
	}, "\n")

	// 2. 拼接待签名字符串
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, tencentService)
	stringToSign := strings.Join([]string{
		"TC3-HMAC-SHA256",
		fmt.Sprintf("%d", timestamp),
		credentialScope,
		sha256Hex(canonicalRequest),
	}, "\n")

	// 3. 计算签名
	secretDate := hmacSHA256([]byte("TC3"+p.secretKey), date)
	secretService := hmacSHA256(secretDate, tencentService)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	// 4. 拼接 Authorization
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		p.secretID, credentialScope, signedHeaders, signature)
}

// tencentRequest 发送腾讯云 API 请求并解析响应中的 Response 字段。
func (p *TencentProvider) tencentRequest(action string, body map[string]interface{}) (map[string]interface{}, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	timestamp := time.Now().Unix()
	authorization := p.tencentSign(action, payload, timestamp)

	req, err := http.NewRequest(http.MethodPost, tencentEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Host", tencentHost)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", tencentVersion)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("Authorization", authorization)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tencent %s request failed: %w", action, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var raw struct {
		Response map[string]interface{} `json:"Response"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}

	if raw.Response == nil {
		return nil, fmt.Errorf("empty response from tencent API (body: %s)", string(respBody))
	}

	// 检查错误
	if errNode, ok := raw.Response["Error"].(map[string]interface{}); ok {
		code, _ := errNode["Code"].(string)
		msg, _ := errNode["Message"].(string)
		return nil, fmt.Errorf("tencent API error [%s]: %s", code, msg)
	}

	return raw.Response, nil
}

// ListDomains 列出所有域名（自动分页）。
func (p *TencentProvider) ListDomains() ([]Domain, error) {
	var domains []Domain
	offset := uint64(0)
	limit := uint64(100)

	for {
		result, err := p.tencentRequest("DescribeDomainList", map[string]interface{}{
			"Offset": offset,
			"Limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		domainList, ok := result["DomainList"].([]interface{})
		if !ok || len(domainList) == 0 {
			break
		}

		for _, item := range domainList {
			d, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			domains = append(domains, Domain{
				Name: getString(d, "Name"),
				ID:   fmt.Sprintf("%.0f", getFloat(d, "DomainId")),
			})
		}

		total, _ := result["TotalCount"].(float64)
		if uint64(len(domains)) >= uint64(total) || uint64(len(domainList)) < limit {
			break
		}
		offset += limit
	}

	return domains, nil
}

// ListRecords 列出指定域名的所有解析记录（自动分页）。
func (p *TencentProvider) ListRecords(domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	offset := uint64(0)
	limit := uint64(100)

	for {
		result, err := p.tencentRequest("DescribeRecordList", map[string]interface{}{
			"Domain": domain,
			"Offset": offset,
			"Limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		recordList, ok := result["RecordList"].([]interface{})
		if !ok || len(recordList) == 0 {
			break
		}

		for _, item := range recordList {
			r, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			records = append(records, DNSRecord{
				Type:  getString(r, "Type"),
				Name:  getString(r, "Name"),
				Value: getString(r, "Value"),
				TTL:   int(getFloat(r, "TTL")),
				ID:    fmt.Sprintf("%.0f", getFloat(r, "RecordId")),
			})
		}

		total, _ := result["TotalCount"].(float64)
		if uint64(len(records)) >= uint64(total) || uint64(len(recordList)) < limit {
			break
		}
		offset += limit
	}

	return records, nil
}

// AddRecord 添加一条解析记录。
func (p *TencentProvider) AddRecord(domain, recordType, name, value string, ttl int) error {
	body := map[string]interface{}{
		"Domain":     domain,
		"SubDomain":  name,
		"RecordType": recordType,
		"RecordLine": "默认",
		"Value":      value,
	}
	if ttl > 0 {
		body["TTL"] = ttl
	}
	_, err := p.tencentRequest("CreateRecord", body)
	return err
}

// DeleteRecord 按类型和名称删除解析记录。
func (p *TencentProvider) DeleteRecord(domain, recordType, name string) error {
	rec, err := p.GetRecord(domain, recordType, name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("record not found: %s %s", recordType, name)
	}
	_, err = p.tencentRequest("DeleteRecord", map[string]interface{}{
		"Domain":   domain,
		"RecordId": rec.ID,
	})
	return err
}

// GetRecord 获取指定类型和名称的单条解析记录。
func (p *TencentProvider) GetRecord(domain, recordType, name string) (*DNSRecord, error) {
	records, err := p.ListRecords(domain)
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].Type == recordType && records[i].Name == name {
			return &records[i], nil
		}
	}
	return nil, nil
}
