package dns

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	aliyunEndpoint         = "https://alidns.aliyuncs.com/"
	aliyunAPIVersion       = "2015-01-09"
	aliyunSignatureMethod  = "HMAC-SHA1"
	aliyunSignatureVersion = "1.0"
)

// AliyunProvider 阿里云 DNS（云解析）提供商实现。
type AliyunProvider struct {
	accessKeyID     string
	accessKeySecret string
	client          *http.Client
}

// NewAliyunProvider 创建阿里云 DNS 提供商实例。
func NewAliyunProvider(accessKeyID, accessKeySecret string) *AliyunProvider {
	return &AliyunProvider{
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *AliyunProvider) Name() string { return "aliyun" }

// aliyunPercentEncode 阿里云专用百分号编码：
// 将 + 替换为 %20，* 替换为 %2A，%7E 替换为 ~。
func aliyunPercentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// sign 计算阿里云 RPC 风格 API 的 HMAC-SHA1 签名。
func (p *AliyunProvider) sign(params map[string]string) string {
	// 1. 按参数名字典序排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接规范化查询字符串
	var canonicalQuery strings.Builder
	for i, k := range keys {
		if i > 0 {
			canonicalQuery.WriteByte('&')
		}
		canonicalQuery.WriteString(aliyunPercentEncode(k))
		canonicalQuery.WriteByte('=')
		canonicalQuery.WriteString(aliyunPercentEncode(params[k]))
	}

	// 3. 构造待签名字符串
	stringToSign := "GET" + "&" + aliyunPercentEncode("/") + "&" + aliyunPercentEncode(canonicalQuery.String())

	// 4. HMAC-SHA1 签名（密钥末尾加 &）
	mac := hmac.New(sha1.New, []byte(p.accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// aliyunRequest 发送阿里云 API 请求并解析 JSON 响应。
func (p *AliyunProvider) aliyunRequest(action string, extraParams map[string]string) (map[string]interface{}, error) {
	params := map[string]string{
		"Format":           "JSON",
		"Version":          aliyunAPIVersion,
		"AccessKeyId":      p.accessKeyID,
		"SignatureMethod":  aliyunSignatureMethod,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": aliyunSignatureVersion,
		"SignatureNonce":   fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(1000000)),
		"Action":           action,
	}
	for k, v := range extraParams {
		params[k] = v
	}

	// 计算签名并附加
	params["Signature"] = p.sign(params)

	// 构造请求 URL
	queryParts := make([]string, 0, len(params))
	for k, v := range params {
		queryParts = append(queryParts, aliyunPercentEncode(k)+"="+aliyunPercentEncode(v))
	}
	reqURL := aliyunEndpoint + "?" + strings.Join(queryParts, "&")

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aliyun %s request failed: %w", action, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}

	// 检查阿里云错误响应
	if code, ok := result["Code"]; ok && code != "" {
		msg, _ := result["Message"].(string)
		return nil, fmt.Errorf("aliyun API error [%s]: %s", code, msg)
	}

	return result, nil
}

// ListDomains 列出所有域名（自动分页，最多 1000 条）。
func (p *AliyunProvider) ListDomains() ([]Domain, error) {
	var domains []Domain
	pageNumber := 1
	pageSize := 100

	for {
		result, err := p.aliyunRequest("DescribeDomains", map[string]string{
			"PageNumber": fmt.Sprintf("%d", pageNumber),
			"PageSize":   fmt.Sprintf("%d", pageSize),
		})
		if err != nil {
			return nil, err
		}

		domainsNode, ok := result["Domains"].(map[string]interface{})
		if !ok {
			break
		}
		domainList, ok := domainsNode["Domain"].([]interface{})
		if !ok || len(domainList) == 0 {
			break
		}

		for _, item := range domainList {
			d, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := d["DomainName"].(string)
			id, _ := d["DomainId"].(string)
			domains = append(domains, Domain{Name: name, ID: id})
		}

		total, _ := result["TotalCount"].(float64)
		if int64(len(domains)) >= int64(total) || len(domainList) < pageSize {
			break
		}
		pageNumber++
	}

	return domains, nil
}

// ListRecords 列出指定域名的所有解析记录（自动分页）。
func (p *AliyunProvider) ListRecords(domain string) ([]DNSRecord, error) {
	var records []DNSRecord
	pageNumber := 1
	pageSize := 100

	for {
		result, err := p.aliyunRequest("DescribeDomainRecords", map[string]string{
			"DomainName": domain,
			"PageNumber": fmt.Sprintf("%d", pageNumber),
			"PageSize":   fmt.Sprintf("%d", pageSize),
		})
		if err != nil {
			return nil, err
		}

		recordsNode, ok := result["DomainRecords"].(map[string]interface{})
		if !ok {
			break
		}
		recordList, ok := recordsNode["Record"].([]interface{})
		if !ok || len(recordList) == 0 {
			break
		}

		for _, item := range recordList {
			r, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			rec := DNSRecord{
				Type:  getString(r, "Type"),
				Name:  getString(r, "RR"),
				Value: getString(r, "Value"),
				TTL:   int(getFloat(r, "TTL")),
				ID:    getString(r, "RecordId"),
			}
			records = append(records, rec)
		}

		total, _ := result["TotalCount"].(float64)
		if int64(len(records)) >= int64(total) || len(recordList) < pageSize {
			break
		}
		pageNumber++
	}

	return records, nil
}

// AddRecord 添加一条解析记录。
func (p *AliyunProvider) AddRecord(domain, recordType, name, value string, ttl int) error {
	params := map[string]string{
		"DomainName": domain,
		"RR":         name,
		"Type":       recordType,
		"Value":      value,
	}
	if ttl > 0 {
		params["TTL"] = fmt.Sprintf("%d", ttl)
	}
	_, err := p.aliyunRequest("AddDomainRecord", params)
	return err
}

// DeleteRecord 按类型和名称删除解析记录。
func (p *AliyunProvider) DeleteRecord(domain, recordType, name string) error {
	rec, err := p.GetRecord(domain, recordType, name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("record not found: %s %s", recordType, name)
	}
	_, err = p.aliyunRequest("DeleteDomainRecord", map[string]string{
		"RecordId": rec.ID,
	})
	return err
}

// GetRecord 获取指定类型和名称的单条解析记录。
func (p *AliyunProvider) GetRecord(domain, recordType, name string) (*DNSRecord, error) {
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

// getString 从 map 中安全获取字符串。
func getString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// getFloat 从 map 中安全获取 float64。
func getFloat(m map[string]interface{}, key string) float64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	f, _ := v.(float64)
	return f
}
