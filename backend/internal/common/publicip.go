package common

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 公网 IP 缓存：避免每个 API 请求都同步访问外网。
//
// 背景（2026-10-23）：dashboard/overview 与 system/info 都会内联调用 ipify.org
// 拉公网 IP，超时 3 秒。服务器出网不通时该请求要等 DNS/连接失败才返回，
// 叠加 CPU 采样阻塞后仪表盘首屏可达 5 秒。公网 IP 属于"几乎不变"的元数据，
// 改为后台定时刷新 + 内存缓存，API 路径零阻塞。
const (
	publicIPTTL     = 10 * time.Minute
	publicIPTimeout = 1500 * time.Millisecond
)

type publicIPCache struct {
	mu        sync.RWMutex
	ip        string
	fetchedAt time.Time
}

var (
	pubIP       = &publicIPCache{}
	pubIPOnce   sync.Once
	pubIPClient = &http.Client{Timeout: publicIPTimeout}
)

// GetPublicIP 返回公网 IP（读内存缓存，绝不阻塞等待外网）
//
// 首次调用时启动后台刷新协程；此刻若尚无缓存则返回空串，
// 几秒后刷新完成，后续调用即可拿到值。
func GetPublicIP() string {
	pubIPOnce.Do(func() { go refreshPublicIPLoop() })

	pubIP.mu.RLock()
	defer pubIP.mu.RUnlock()
	if pubIP.ip == "" || time.Since(pubIP.fetchedAt) > publicIPTTL {
		return pubIP.ip // 过期时返回旧值，不触发同步探测
	}
	return pubIP.ip
}

// refreshPublicIPLoop 后台定时刷新公网 IP
func refreshPublicIPLoop() {
	for {
		fetchPublicIP()
		time.Sleep(publicIPTTL)
	}
}

// fetchPublicIP 拉取公网 IP 并更新缓存
func fetchPublicIP() {
	resp, err := pubIPClient.Get("https://api.ipify.org?format=text")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return
	}
	pubIP.mu.Lock()
	pubIP.ip = ip
	pubIP.fetchedAt = time.Now()
	pubIP.mu.Unlock()
}
