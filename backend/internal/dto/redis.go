package dto

// RedisKeyInfo Redis Key 信息
type RedisKeyInfo struct {
	Key  string `json:"key"`
	Type string `json:"type"` // string | hash | list | set | zset | none
	TTL  int64  `json:"ttl"`  // 秒，-1=永不过期，-2=不存在
	Size int64  `json:"size"` // 元素数量或字符串长度
}

// RedisKeyGroup Redis Key 分组（按前缀）
type RedisKeyGroup struct {
	Prefix string         `json:"prefix"`
	Count  int            `json:"count"`
	Keys   []RedisKeyInfo `json:"keys"`
}

// RedisScanResponse Scan 分页响应
type RedisScanResponse struct {
	Keys   []RedisKeyInfo `json:"keys"`
	Cursor uint64         `json:"cursor"` // 0=扫描完成
}

// RedisValueResponse Redis Key 值响应
type RedisValueResponse struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	TTL   int64       `json:"ttl"`
	Value interface{} `json:"value"`
}

// RedisSetKeyRequest 设置 Key 请求
type RedisSetKeyRequest struct {
	Value interface{} `json:"value" binding:"required"`
	TTL   int         `json:"ttl"` // 秒，0=永不过期
	// Hash 字段操作
	Field string `json:"field"`
	// List 操作
	Direction string `json:"direction"` // left | right
	// ZSet 操作
	Score *float64 `json:"score"`
	// Set 成员
	Member interface{} `json:"member"`
}

// RedisDeleteKeyRequest 删除 Key 请求（支持 Hash/List/Set/ZSet 子元素删除）
type RedisDeleteKeyRequest struct {
	Field  string      `json:"field"`  // Hash field
	Member interface{} `json:"member"` // Set/ZSet member
	Index  int         `json:"index"`  // List index
}

// RedisExpireRequest 设置 Key 过期时间请求
type RedisExpireRequest struct {
	Key string `json:"key" binding:"required"`
	TTL int    `json:"ttl" binding:"required"` // 秒
}

// RedisCLIRequest Redis CLI 请求
type RedisCLIRequest struct {
	Command string   `json:"command" binding:"required"`
	Args    []string `json:"args"`
}

// RedisCLIResponse Redis CLI 响应
type RedisCLIResponse struct {
	Type   string      `json:"type"` // string | integer | array | error | nil | bulk
	Result interface{} `json:"result"`
}

// RedisMemoryInfo Redis 内存信息
type RedisMemoryInfo struct {
	UsedMemory            int64   `json:"used_memory"`
	UsedMemoryHuman       string  `json:"used_memory_human"`
	UsedMemoryPeak        int64   `json:"used_memory_peak"`
	UsedMemoryPeakHuman   string  `json:"used_memory_peak_human"`
	MaxMemory             int64   `json:"maxmemory"`
	MaxMemoryHuman        string  `json:"maxmemory_human"`
	MaxMemoryPolicy       string  `json:"maxmemory_policy"`
	MemFragmentationRatio float64 `json:"mem_fragmentation_ratio"`
}

// RedisBigKey 大 Key 统计
type RedisBigKey struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

// RedisMemoryAnalysis Redis 内存分析响应
type RedisMemoryAnalysis struct {
	Memory  RedisMemoryInfo `json:"memory"`
	DBSize  int64           `json:"db_size"`
	BigKeys []RedisBigKey   `json:"big_keys"`
}

// RedisInfoResponse Redis INFO 全量响应
type RedisInfoResponse struct {
	Sections map[string]map[string]string `json:"sections"`
}
