package api

import (
	"net/url"
	"strconv"
	"strings"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	dbsvc "cloudpivot/internal/service/database"

	"github.com/gin-gonic/gin"
)

// RedisHandler Redis 管理 API Handler
type RedisHandler struct {
	redisSvc *dbsvc.RedisService
}

// NewRedisHandler 创建 Redis 管理 Handler
func NewRedisHandler() *RedisHandler {
	return &RedisHandler{
		redisSvc: dbsvc.NewRedisService(model.DB),
	}
}

// RegisterRedisRoutes 注册 Redis 管理路由
func RegisterRedisRoutes(r *gin.RouterGroup) {
	h := NewRedisHandler()

	// Key 浏览器：Scan 分页
	r.GET("/redis/:id/keys", middleware.RequirePermission(model.PermRedisView), h.ScanKeys)
	// 单个 Key 操作（*key 捕获含 / 的 key 名）
	r.GET("/redis/:id/keys/*key", middleware.RequirePermission(model.PermRedisView), h.GetKey)
	r.PUT("/redis/:id/keys/*key", middleware.RequirePermission(model.PermRedisManage), h.SetKey)
	r.DELETE("/redis/:id/keys/*key", middleware.RequirePermission(model.PermRedisManage), h.DeleteKey)
	// 设置过期时间（key 通过 body 传递，避免 catch-all 路径冲突）
	r.POST("/redis/:id/keys/expire", middleware.RequirePermission(model.PermRedisManage), h.ExpireKey)

	// Redis CLI
	r.POST("/redis/:id/cli", middleware.RequirePermission(model.PermRedisManage), h.ExecuteCLI)

	// 内存分析
	r.GET("/redis/:id/memory", middleware.RequirePermission(model.PermRedisView), h.GetMemory)

	// INFO 全量
	r.GET("/redis/:id/info", middleware.RequirePermission(model.PermRedisView), h.GetInfo)
}

// parseRedisKey 从通配路径参数解析 key（去除前导 / 并 URL 解码）
func parseRedisKey(c *gin.Context) string {
	key := c.Param("key")
	key = strings.TrimPrefix(key, "/")
	if decoded, err := url.PathUnescape(key); err == nil {
		return decoded
	}
	return key
}

// ========== Key 浏览器 ==========

// ScanKeys GET /redis/:id/keys?pattern=&cursor=&count=
func (h *RedisHandler) ScanKeys(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	pattern := c.DefaultQuery("pattern", "*")
	cursor, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 64)
	count, _ := strconv.ParseInt(c.DefaultQuery("count", "50"), 10, 64)

	result, err := h.redisSvc.ScanKeys(id, pattern, cursor, count)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// ========== 单个 Key 操作 ==========

// GetKey GET /redis/:id/keys/:key
func (h *RedisHandler) GetKey(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	key := parseRedisKey(c)
	if key == "" {
		common.Fail(c, common.CodeParamError, "key 不能为空")
		return
	}

	result, err := h.redisSvc.GetKeyValue(id, key)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// SetKey PUT /redis/:id/keys/:key
func (h *RedisHandler) SetKey(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	key := parseRedisKey(c)
	if key == "" {
		common.Fail(c, common.CodeParamError, "key 不能为空")
		return
	}

	var req dto.RedisSetKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	if err := h.redisSvc.SetKeyValue(id, key, req); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// DeleteKey DELETE /redis/:id/keys/:key
func (h *RedisHandler) DeleteKey(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	key := parseRedisKey(c)
	if key == "" {
		common.Fail(c, common.CodeParamError, "key 不能为空")
		return
	}

	// 可选：删除子元素（hash field / list element / set member / zset member）
	var req *dto.RedisDeleteKeyRequest
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			// 绑定失败时按删除整个 key 处理
			req = nil
		}
	}

	if err := h.redisSvc.DeleteKey(id, key, req); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// ExpireKey POST /redis/:id/keys/expire
func (h *RedisHandler) ExpireKey(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}

	var req dto.RedisExpireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if req.Key == "" {
		common.Fail(c, common.CodeParamError, "key 不能为空")
		return
	}
	if req.TTL <= 0 {
		common.Fail(c, common.CodeParamError, "ttl 必须大于 0")
		return
	}

	if err := h.redisSvc.ExpireKey(id, req.Key, req.TTL); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// ========== Redis CLI ==========

// ExecuteCLI POST /redis/:id/cli
func (h *RedisHandler) ExecuteCLI(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.RedisCLIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	result, err := h.redisSvc.ExecuteCLI(id, req.Command, req.Args)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// ========== 内存分析 ==========

// GetMemory GET /redis/:id/memory
func (h *RedisHandler) GetMemory(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	result, err := h.redisSvc.GetMemoryInfo(id)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// ========== INFO 全量 ==========

// GetInfo GET /redis/:id/info
func (h *RedisHandler) GetInfo(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	result, err := h.redisSvc.GetFullInfo(id)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}
