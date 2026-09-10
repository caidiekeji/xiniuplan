package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"
	dnssvc "cloudpivot/internal/service/dns"
	dnsprovider "cloudpivot/pkg/dns"

	"github.com/gin-gonic/gin"
)

// DNSHandler DNS 管理 API Handler。
type DNSHandler struct {
	svc *dnssvc.Service
}

// NewDNSHandler 创建 DNS Handler。
func NewDNSHandler(svc *dnssvc.Service) *DNSHandler {
	return &DNSHandler{svc: svc}
}

// RegisterDNSRoutes 注册 DNS 管理路由。
// 路由前缀由调用方的 RouterGroup 决定（通常为 /api/v1）。
func RegisterDNSRoutes(r *gin.RouterGroup) {
	svc := dnssvc.NewService(model.DB)
	h := NewDNSHandler(svc)

	dns := r.Group("/dns")
	{
		// 账号管理
		dns.GET("/accounts", h.ListAccounts)
		dns.POST("/accounts", h.CreateAccount)
		dns.GET("/accounts/:id", h.GetAccount)
		dns.PUT("/accounts/:id", h.UpdateAccount)
		dns.DELETE("/accounts/:id", h.DeleteAccount)

		// 域名与解析记录
		dns.GET("/accounts/:id/domains", h.GetDomains)
		dns.GET("/accounts/:id/domains/:domain/records", h.GetRecords)
		dns.POST("/accounts/:id/domains/:domain/records", h.AddRecord)
		dns.DELETE("/accounts/:id/domains/:domain/records", h.DeleteRecord)
	}
}

// parseAccountID 从 URL 参数中解析账号 ID。
func parseAccountID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的账号 ID: "+idStr)
		return 0, false
	}
	return uint(id), true
}

// toRecordResponse 将提供商记录转换为响应 DTO。
func toRecordResponse(records []dnsprovider.DNSRecord) []dto.DNSRecordResponse {
	result := make([]dto.DNSRecordResponse, 0, len(records))
	for _, r := range records {
		result = append(result, dto.DNSRecordResponse{
			Type:  r.Type,
			Name:  r.Name,
			Value: r.Value,
			TTL:   r.TTL,
			ID:    r.ID,
		})
	}
	return result
}

// ListAccounts GET /dns/accounts - 列出所有 DNS 账号。
func (h *DNSHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.svc.ListAccounts()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 DNS 账号列表失败: "+err.Error())
		return
	}
	common.Success(c, accounts)
}

// CreateAccount POST /dns/accounts - 创建 DNS 账号。
func (h *DNSHandler) CreateAccount(c *gin.Context) {
	var req dto.CreateDNSAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	account, err := h.svc.CreateAccount(req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "创建 DNS 账号失败: "+err.Error())
		return
	}
	common.Success(c, account)
}

// GetAccount GET /dns/accounts/:id - 获取单个 DNS 账号详情。
func (h *DNSHandler) GetAccount(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}

	accounts, err := h.svc.ListAccounts()
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取 DNS 账号失败: "+err.Error())
		return
	}
	for i := range accounts {
		if accounts[i].ID == id {
			common.Success(c, accounts[i])
			return
		}
	}
	common.Fail(c, common.CodeNotFound, "DNS 账号不存在")
}

// UpdateAccount PUT /dns/accounts/:id - 更新 DNS 账号。
func (h *DNSHandler) UpdateAccount(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}

	var req dto.UpdateDNSAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	account, err := h.svc.UpdateAccount(id, req)
	if err != nil {
		common.Fail(c, common.CodeServerError, "更新 DNS 账号失败: "+err.Error())
		return
	}
	common.Success(c, account)
}

// DeleteAccount DELETE /dns/accounts/:id - 删除 DNS 账号。
func (h *DNSHandler) DeleteAccount(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteAccount(id); err != nil {
		common.Fail(c, common.CodeServerError, "删除 DNS 账号失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}

// GetDomains GET /dns/accounts/:id/domains - 获取域名列表。
func (h *DNSHandler) GetDomains(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}

	domains, err := h.svc.GetDomains(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取域名列表失败: "+err.Error())
		return
	}
	common.Success(c, domains)
}

// GetRecords GET /dns/accounts/:id/domains/:domain/records - 获取解析记录列表。
func (h *DNSHandler) GetRecords(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}
	domain := c.Param("domain")
	if domain == "" {
		common.Fail(c, common.CodeParamError, "域名不能为空")
		return
	}

	records, err := h.svc.GetRecords(id, domain)
	if err != nil {
		common.Fail(c, common.CodeServerError, "获取解析记录失败: "+err.Error())
		return
	}
	common.Success(c, toRecordResponse(records))
}

// AddRecord POST /dns/accounts/:id/domains/:domain/records - 添加解析记录。
func (h *DNSHandler) AddRecord(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}
	domain := c.Param("domain")
	if domain == "" {
		common.Fail(c, common.CodeParamError, "域名不能为空")
		return
	}

	var req dto.AddDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.AddRecord(id, domain, req.Type, req.Name, req.Value, req.TTL); err != nil {
		common.Fail(c, common.CodeServerError, "添加解析记录失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"added": true})
}

// DeleteRecord DELETE /dns/accounts/:id/domains/:domain/records - 删除解析记录。
func (h *DNSHandler) DeleteRecord(c *gin.Context) {
	id, ok := parseAccountID(c)
	if !ok {
		return
	}
	domain := c.Param("domain")
	if domain == "" {
		common.Fail(c, common.CodeParamError, "域名不能为空")
		return
	}

	var req dto.DeleteDNSRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.DeleteRecord(id, domain, req.Type, req.Name); err != nil {
		common.Fail(c, common.CodeServerError, "删除解析记录失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"deleted": true})
}
