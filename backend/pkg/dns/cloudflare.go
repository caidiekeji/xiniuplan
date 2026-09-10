package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	cloudflareEndpoint = "https://api.cloudflare.com/client/v4"
)

// CloudflareProvider Cloudflare DNS 提供商实现，使用 API v4 + Bearer Token 认证。
type CloudflareProvider struct {
	apiToken string
	client   *http.Client
}

// NewCloudflareProvider 创建 Cloudflare DNS 提供商实例。
func NewCloudflareProvider(apiToken string) *CloudflareProvider {
	return &CloudflareProvider{
		apiToken: apiToken,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *CloudflareProvider) Name() string { return "cloudflare" }

// cloudflareResponse Cloudflare API 统一响应结构。
type cloudflareResponse struct {
	Success    bool            `json:"success"`
	Errors     []cfError       `json:"errors"`
	Messages   []interface{}   `json:"messages"`
	Result     json.RawMessage `json:"result"`
	ResultInfo *cfResultInfo   `json:"result_info"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	Total      int `json:"total"`
}

// cfZone Cloudflare 区域（域名）。
type cfZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// cfDNSRecord Cloudflare DNS 记录。
type cfDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

// cloudflareRequest 发送 Cloudflare API 请求。
func (p *CloudflareProvider) cloudflareRequest(method, path string, query url.Values, body interface{}) (*cloudflareResponse, error) {
	reqURL := cloudflareEndpoint + path
	if query != nil {
		reqURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudflare %s %s request failed: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var cfResp cloudflareResponse
	if err := json.Unmarshal(respBody, &cfResp); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare API error [%d]: %s", cfResp.Errors[0].Code, cfResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API request failed (body: %s)", string(respBody))
	}

	return &cfResp, nil
}

// findZoneID 通过域名查找 Cloudflare Zone ID。
func (p *CloudflareProvider) findZoneID(domain string) (string, error) {
	query := url.Values{}
	query.Set("name", domain)
	query.Set("per_page", "1")

	resp, err := p.cloudflareRequest(http.MethodGet, "/zones", query, nil)
	if err != nil {
		return "", err
	}

	var zones []cfZone
	if err := json.Unmarshal(resp.Result, &zones); err != nil {
		return "", fmt.Errorf("parse zones: %w", err)
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("zone not found for domain: %s", domain)
	}

	return zones[0].ID, nil
}

// ListDomains 列出所有域名（Zone），自动分页。
func (p *CloudflareProvider) ListDomains() ([]Domain, error) {
	var domains []Domain
	page := 1
	perPage := 100

	for {
		query := url.Values{}
		query.Set("page", fmt.Sprintf("%d", page))
		query.Set("per_page", fmt.Sprintf("%d", perPage))

		resp, err := p.cloudflareRequest(http.MethodGet, "/zones", query, nil)
		if err != nil {
			return nil, err
		}

		var zones []cfZone
		if err := json.Unmarshal(resp.Result, &zones); err != nil {
			return nil, fmt.Errorf("parse zones: %w", err)
		}

		for _, z := range zones {
			domains = append(domains, Domain{Name: z.Name, ID: z.ID})
		}

		if resp.ResultInfo == nil || page >= resp.ResultInfo.TotalPages || len(zones) == 0 {
			break
		}
		page++
	}

	return domains, nil
}

// ListRecords 列出指定域名的所有解析记录，自动分页。
func (p *CloudflareProvider) ListRecords(domain string) ([]DNSRecord, error) {
	zoneID, err := p.findZoneID(domain)
	if err != nil {
		return nil, err
	}

	var records []DNSRecord
	page := 1
	perPage := 100

	for {
		query := url.Values{}
		query.Set("page", fmt.Sprintf("%d", page))
		query.Set("per_page", fmt.Sprintf("%d", perPage))

		path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
		resp, err := p.cloudflareRequest(http.MethodGet, path, query, nil)
		if err != nil {
			return nil, err
		}

		var cfRecords []cfDNSRecord
		if err := json.Unmarshal(resp.Result, &cfRecords); err != nil {
			return nil, fmt.Errorf("parse dns records: %w", err)
		}

		for _, r := range cfRecords {
			// Cloudflare 的 name 是完整域名（如 www.example.com），
			// 统一转换为相对主机记录（如 www）。
			name := r.Name
			if len(name) > len(domain) && name[len(name)-len(domain):] == domain {
				name = name[:len(name)-len(domain)-1]
			}
			if name == domain {
				name = "@"
			}
			records = append(records, DNSRecord{
				Type:  r.Type,
				Name:  name,
				Value: r.Content,
				TTL:   r.TTL,
				ID:    r.ID,
			})
		}

		if resp.ResultInfo == nil || page >= resp.ResultInfo.TotalPages || len(cfRecords) == 0 {
			break
		}
		page++
	}

	return records, nil
}

// AddRecord 添加一条解析记录。
func (p *CloudflareProvider) AddRecord(domain, recordType, name, value string, ttl int) error {
	zoneID, err := p.findZoneID(domain)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"type":    recordType,
		"name":    name,
		"content": value,
		"proxied": false,
	}
	if ttl > 0 {
		body["ttl"] = ttl
	} else {
		body["ttl"] = 1 // Cloudflare 自动 TTL 用 1 表示
	}

	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	_, err = p.cloudflareRequest(http.MethodPost, path, nil, body)
	return err
}

// DeleteRecord 按类型和名称删除解析记录。
func (p *CloudflareProvider) DeleteRecord(domain, recordType, name string) error {
	rec, err := p.GetRecord(domain, recordType, name)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("record not found: %s %s", recordType, name)
	}

	zoneID, err := p.findZoneID(domain)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, rec.ID)
	_, err = p.cloudflareRequest(http.MethodDelete, path, nil, nil)
	return err
}

// GetRecord 获取指定类型和名称的单条解析记录。
func (p *CloudflareProvider) GetRecord(domain, recordType, name string) (*DNSRecord, error) {
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
