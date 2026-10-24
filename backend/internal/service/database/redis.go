package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RedisService Redis 管理服务
type RedisService struct {
	db *gorm.DB
}

// NewRedisService 创建 Redis 管理服务
func NewRedisService(db *gorm.DB) *RedisService {
	return &RedisService{db: db}
}

// getRedisInstance 获取 Redis 实例并校验类型
func (s *RedisService) getRedisInstance(instanceID uint) (*model.DatabaseInstance, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}
	if instance.Type != "redis" {
		return nil, NewUnsupportedTypeError(instance.Type, "该操作仅支持 Redis 实例")
	}
	return &instance, nil
}

// newClient 创建 Redis 客户端
func (s *RedisService) newClient(instance *model.DatabaseInstance) *redis.Client {
	host := instance.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, instance.Port),
		Password:     instance.RootPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     5,
	})
}

// connectRedis 连接到指定 Redis 实例
func (s *RedisService) connectRedis(instanceID uint) (*redis.Client, *model.DatabaseInstance, error) {
	instance, err := s.getRedisInstance(instanceID)
	if err != nil {
		return nil, nil, err
	}
	client := s.newClient(instance)
	return client, instance, nil
}

// keySize 获取 key 的大小（元素数量或字符串长度）
func keySize(ctx context.Context, client *redis.Client, key, keyType string) int64 {
	switch keyType {
	case "string":
		n, err := client.StrLen(ctx, key).Result()
		if err != nil {
			return 0
		}
		return n
	case "hash":
		n, err := client.HLen(ctx, key).Result()
		if err != nil {
			return 0
		}
		return n
	case "list":
		n, err := client.LLen(ctx, key).Result()
		if err != nil {
			return 0
		}
		return n
	case "set":
		n, err := client.SCard(ctx, key).Result()
		if err != nil {
			return 0
		}
		return n
	case "zset":
		n, err := client.ZCard(ctx, key).Result()
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

// ScanKeys Scan 分页扫描 key
func (s *RedisService) ScanKeys(instanceID uint, pattern string, cursor uint64, count int64) (*dto.RedisScanResponse, error) {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if count <= 0 {
		count = 50
	}
	if pattern == "" {
		pattern = "*"
	}

	keys, nextCursor, err := client.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, fmt.Errorf("扫描 key 失败: %w", err)
	}

	result := make([]dto.RedisKeyInfo, 0, len(keys))
	for _, key := range keys {
		keyType, err := client.Type(ctx, key).Result()
		if err != nil {
			keyType = "none"
		}
		ttl := client.TTL(ctx, key).Val()
		var ttlSec int64 = -1
		if ttl > 0 {
			ttlSec = int64(ttl.Seconds())
		} else if ttl == -2*time.Nanosecond {
			ttlSec = -2
		}
		result = append(result, dto.RedisKeyInfo{
			Key:  key,
			Type: keyType,
			TTL:  ttlSec,
			Size: keySize(ctx, client, key, keyType),
		})
	}

	return &dto.RedisScanResponse{
		Keys:   result,
		Cursor: nextCursor,
	}, nil
}

// GetKeyValue 获取 key 的值（按类型返回）
func (s *RedisService) GetKeyValue(instanceID uint, key string) (*dto.RedisValueResponse, error) {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	keyType, err := client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取 key 类型失败: %w", err)
	}
	if keyType == "none" {
		return nil, fmt.Errorf("key 不存在: %s", key)
	}

	ttl := client.TTL(ctx, key).Val()
	var ttlSec int64 = -1
	if ttl > 0 {
		ttlSec = int64(ttl.Seconds())
	}

	resp := &dto.RedisValueResponse{
		Key:  key,
		Type: keyType,
		TTL:  ttlSec,
	}

	switch keyType {
	case "string":
		val, err := client.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("获取 string 值失败: %w", err)
		}
		resp.Value = val
	case "hash":
		val, err := client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("获取 hash 值失败: %w", err)
		}
		resp.Value = val
	case "list":
		length, _ := client.LLen(ctx, key).Result()
		stop := length - 1
		if stop > 999 {
			stop = 999
		}
		val, err := client.LRange(ctx, key, 0, stop).Result()
		if err != nil {
			return nil, fmt.Errorf("获取 list 值失败: %w", err)
		}
		resp.Value = val
	case "set":
		val, err := client.SMembers(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("获取 set 值失败: %w", err)
		}
		resp.Value = val
	case "zset":
		length, _ := client.ZCard(ctx, key).Result()
		stop := length - 1
		if stop > 999 {
			stop = 999
		}
		raw, err := client.ZRangeWithScores(ctx, key, 0, stop).Result()
		if err != nil {
			return nil, fmt.Errorf("获取 zset 值失败: %w", err)
		}
		type zsetItem struct {
			Member string  `json:"member"`
			Score  float64 `json:"score"`
		}
		items := make([]zsetItem, 0, len(raw))
		for _, z := range raw {
			items = append(items, zsetItem{Member: z.Member.(string), Score: z.Score})
		}
		resp.Value = items
	default:
		resp.Value = nil
	}

	return resp, nil
}

// SetKeyValue 设置 key 值（按类型处理）
func (s *RedisService) SetKeyValue(instanceID uint, key string, req dto.RedisSetKeyRequest) error {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return err
	}
	defer client.Close()

	expiration := time.Duration(0)
	if req.TTL > 0 {
		expiration = time.Duration(req.TTL) * time.Second
	}

	// 检测现有 key 类型，决定操作方式
	keyType, _ := client.Type(ctx, key).Result()

	// 如果指定了 field，按 hash 处理
	if req.Field != "" {
		if err := client.HSet(ctx, key, req.Field, fmt.Sprintf("%v", req.Value)).Err(); err != nil {
			return fmt.Errorf("HSET 失败: %w", err)
		}
		if expiration > 0 {
			client.Expire(ctx, key, expiration)
		}
		return nil
	}

	// 如果指定了 direction，按 list 处理
	if req.Direction != "" {
		val := fmt.Sprintf("%v", req.Value)
		var err error
		if strings.ToLower(req.Direction) == "left" {
			err = client.LPush(ctx, key, val).Err()
		} else {
			err = client.RPush(ctx, key, val).Err()
		}
		if err != nil {
			return fmt.Errorf("list push 失败: %w", err)
		}
		if expiration > 0 {
			client.Expire(ctx, key, expiration)
		}
		return nil
	}

	// 如果指定了 score，按 zset 处理
	if req.Score != nil {
		member := fmt.Sprintf("%v", req.Value)
		if err := client.ZAdd(ctx, key, redis.Z{Score: *req.Score, Member: member}).Err(); err != nil {
			return fmt.Errorf("ZADD 失败: %w", err)
		}
		if expiration > 0 {
			client.Expire(ctx, key, expiration)
		}
		return nil
	}

	// 如果指定了 member 但没有 score，按 set 处理
	if req.Member != nil {
		if err := client.SAdd(ctx, key, fmt.Sprintf("%v", req.Member)).Err(); err != nil {
			return fmt.Errorf("SADD 失败: %w", err)
		}
		if expiration > 0 {
			client.Expire(ctx, key, expiration)
		}
		return nil
	}

	// 默认按 string 处理（或覆盖已有 key）
	if keyType != "none" && keyType != "string" {
		return fmt.Errorf("key %s 已存在且类型为 %s，无法直接覆盖；请使用对应类型的操作参数", key, keyType)
	}
	if err := client.Set(ctx, key, fmt.Sprintf("%v", req.Value), expiration).Err(); err != nil {
		return fmt.Errorf("SET 失败: %w", err)
	}
	return nil
}

// DeleteKey 删除 key 或 key 中的子元素
func (s *RedisService) DeleteKey(instanceID uint, key string, req *dto.RedisDeleteKeyRequest) error {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return err
	}
	defer client.Close()

	if req == nil {
		// 删除整个 key
		if err := client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("DEL 失败: %w", err)
		}
		return nil
	}

	keyType, _ := client.Type(ctx, key).Result()

	switch keyType {
	case "hash":
		if req.Field == "" {
			return fmt.Errorf("删除 hash 子元素需要指定 field")
		}
		if err := client.HDel(ctx, key, req.Field).Err(); err != nil {
			return fmt.Errorf("HDEL 失败: %w", err)
		}
	case "list":
		if req.Member == nil {
			return fmt.Errorf("删除 list 元素需要指定 member")
		}
		if err := client.LRem(ctx, key, 0, fmt.Sprintf("%v", req.Member)).Err(); err != nil {
			return fmt.Errorf("LREM 失败: %w", err)
		}
	case "set":
		if req.Member == nil {
			return fmt.Errorf("删除 set 元素需要指定 member")
		}
		if err := client.SRem(ctx, key, fmt.Sprintf("%v", req.Member)).Err(); err != nil {
			return fmt.Errorf("SREM 失败: %w", err)
		}
	case "zset":
		if req.Member == nil {
			return fmt.Errorf("删除 zset 元素需要指定 member")
		}
		if err := client.ZRem(ctx, key, fmt.Sprintf("%v", req.Member)).Err(); err != nil {
			return fmt.Errorf("ZREM 失败: %w", err)
		}
	default:
		if err := client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("DEL 失败: %w", err)
		}
	}
	return nil
}

// ExpireKey 设置 key 过期时间
func (s *RedisService) ExpireKey(instanceID uint, key string, ttl int) error {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return err
	}
	defer client.Close()

	expiration := time.Duration(ttl) * time.Second
	if err := client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("EXPIRE 失败: %w", err)
	}
	return nil
}

// ExecuteCLI 执行任意 Redis 命令
func (s *RedisService) ExecuteCLI(instanceID uint, command string, args []string) (*dto.RedisCLIResponse, error) {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("命令不能为空")
	}

	// 构建参数：command + args
	fullArgs := make([]interface{}, 0, len(args)+1)
	fullArgs = append(fullArgs, command)
	for _, a := range args {
		fullArgs = append(fullArgs, a)
	}

	result, err := client.Do(ctx, fullArgs...).Result()
	if err != nil {
		if err == redis.Nil {
			return &dto.RedisCLIResponse{Type: "nil", Result: nil}, nil
		}
		return &dto.RedisCLIResponse{Type: "error", Result: err.Error()}, nil
	}

	resp := &dto.RedisCLIResponse{}
	switch v := result.(type) {
	case string:
		resp.Type = "string"
		resp.Result = v
	case int64:
		resp.Type = "integer"
		resp.Result = v
	case []interface{}:
		resp.Type = "array"
		// 转换为字符串数组以便 JSON 序列化
		strArr := make([]string, 0, len(v))
		for _, item := range v {
			strArr = append(strArr, fmt.Sprintf("%v", item))
		}
		resp.Result = strArr
	case []string:
		resp.Type = "array"
		resp.Result = v
	case nil:
		resp.Type = "nil"
		resp.Result = nil
	default:
		resp.Type = "bulk"
		resp.Result = fmt.Sprintf("%v", v)
	}
	return resp, nil
}

// GetMemoryInfo 获取内存分析信息
func (s *RedisService) GetMemoryInfo(instanceID uint) (*dto.RedisMemoryAnalysis, error) {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// INFO memory
	infoStr, err := client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("获取内存信息失败: %w", err)
	}
	memInfo := parseRedisInfo(infoStr)

	memory := dto.RedisMemoryInfo{
		UsedMemory:            parseInt64(memInfo["used_memory"]),
		UsedMemoryHuman:       memInfo["used_memory_human"],
		UsedMemoryPeak:        parseInt64(memInfo["used_memory_peak"]),
		UsedMemoryPeakHuman:   memInfo["used_memory_peak_human"],
		MaxMemory:             parseInt64(memInfo["maxmemory"]),
		MaxMemoryHuman:        memInfo["maxmemory_human"],
		MaxMemoryPolicy:       memInfo["maxmemory_policy"],
		MemFragmentationRatio: parseFloat64(memInfo["mem_fragmentation_ratio"]),
	}

	// DBSIZE
	dbSize, err := client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取 DBSIZE 失败: %w", err)
	}

	// 大 key 扫描（限制扫描数量避免超时）
	bigKeys := s.scanBigKeys(ctx, client, 5000, 20)

	return &dto.RedisMemoryAnalysis{
		Memory:  memory,
		DBSize:  dbSize,
		BigKeys: bigKeys,
	}, nil
}

// scanBigKeys 扫描大 key（模拟 --bigkeys）
func (s *RedisService) scanBigKeys(ctx context.Context, client *redis.Client, maxScan int64, topN int) []dto.RedisBigKey {
	type keyInfo struct {
		key  string
		typ  string
		size int64
	}
	var allKeys []keyInfo

	var cursor uint64
	scanned := int64(0)
	for scanned < maxScan {
		keys, next, err := client.Scan(ctx, cursor, "*", 500).Result()
		if err != nil {
			break
		}
		for _, key := range keys {
			keyType, _ := client.Type(ctx, key).Result()
			sz := keySize(ctx, client, key, keyType)
			allKeys = append(allKeys, keyInfo{key: key, typ: keyType, size: sz})
		}
		scanned += int64(len(keys))
		cursor = next
		if cursor == 0 {
			break
		}
	}

	// 按类型分组取 top N
	typeStats := make(map[string][]keyInfo)
	for _, ks := range allKeys {
		typeStats[ks.typ] = append(typeStats[ks.typ], ks)
	}

	var result []dto.RedisBigKey
	for typ, list := range typeStats {
		// 按 size 降序排序（简单选择排序 topN）
		for i := 0; i < len(list) && i < topN; i++ {
			maxIdx := i
			for j := i + 1; j < len(list); j++ {
				if list[j].size > list[maxIdx].size {
					maxIdx = j
				}
			}
			list[i], list[maxIdx] = list[maxIdx], list[i]
			result = append(result, dto.RedisBigKey{
				Key:  list[i].key,
				Type: typ,
				Size: list[i].size,
			})
		}
	}
	return result
}

// GetFullInfo 获取 Redis INFO 全量
func (s *RedisService) GetFullInfo(instanceID uint) (*dto.RedisInfoResponse, error) {
	ctx := context.Background()
	client, _, err := s.connectRedis(instanceID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	infoStr, err := client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取 INFO 失败: %w", err)
	}

	sections := make(map[string]map[string]string)
	var currentSection string

	for _, line := range strings.Split(infoStr, "\r\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") {
			if strings.HasPrefix(line, "# ") {
				currentSection = strings.TrimPrefix(line, "# ")
				sections[currentSection] = make(map[string]string)
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && currentSection != "" {
			sections[currentSection][parts[0]] = parts[1]
		}
	}

	return &dto.RedisInfoResponse{Sections: sections}, nil
}

// parseRedisInfo 解析 Redis INFO 输出为 map
func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(info, "\r\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func parseFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
