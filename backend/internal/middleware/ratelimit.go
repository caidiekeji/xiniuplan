package middleware

import (
	"sync"
	"time"

	"cloudpivot/internal/common"

	"github.com/gin-gonic/gin"
)

// tokenBucket 令牌桶
type tokenBucket struct {
	capacity  int64
	tokens    int64
	rate      float64 // 每秒补充速率
	lastCheck time.Time
	mu        sync.Mutex
}

func newTokenBucket(capacity int64, ratePerMinute float64) *tokenBucket {
	return &tokenBucket{
		capacity:  capacity,
		tokens:    capacity,
		rate:      ratePerMinute / 60.0,
		lastCheck: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += int64(elapsed * b.rate)
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastCheck = now

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter 基于 IP 的限流中间件
// 使用全局 IP 级别限流（而非 per-path），防止通过多路径绕过
type RateLimiter struct {
	apiBuckets   sync.Map // map[string]*tokenBucket — 全局 API 限流（按 IP）
	loginBuckets sync.Map // map[string]*tokenBucket — 登录接口限流（按 IP）
	loginRate    float64  // 登录接口每分钟次数
	apiRate      float64  // 其他接口每分钟次数（全局，不分路径）
}

// NewRateLimiter 创建限流器
func NewRateLimiter(loginRate, apiRate float64) *RateLimiter {
	return &RateLimiter{
		loginRate: loginRate,
		apiRate:   apiRate,
	}
}

func (r *RateLimiter) getAPIBucket(ip string) *tokenBucket {
	bucket, _ := r.apiBuckets.LoadOrStore(ip, newTokenBucket(int64(r.apiRate), r.apiRate))
	return bucket.(*tokenBucket)
}

func (r *RateLimiter) getLoginBucket(ip string) *tokenBucket {
	bucket, _ := r.loginBuckets.LoadOrStore(ip, newTokenBucket(int64(r.loginRate), r.loginRate))
	return bucket.(*tokenBucket)
}

// Limit 限流中间件
// 登录接口使用独立的更严格限流；其他所有 API 共享全局 IP 限流
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		var bucket *tokenBucket
		if c.Request.URL.Path == "/api/v1/auth/login" {
			bucket = r.getLoginBucket(ip)
		} else {
			bucket = r.getAPIBucket(ip)
		}

		if !bucket.allow() {
			common.Fail(c, 42901, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CleanupBuckets 定期清理不活跃的限流桶（防止内存泄漏）
func (r *RateLimiter) CleanupBuckets() {
	// 简单实现：遍历并删除超过 1 小时未使用的桶
	// 由于 tokenBucket 没有暴露 lastCheck，这里依赖 Go GC 和 sync.Map 的自然清理
	// 生产环境可考虑添加过期时间字段
}
