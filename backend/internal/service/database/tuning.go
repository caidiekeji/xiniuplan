package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloudpivot/internal/model"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

// TuningService 数据库参数配置（DB-002）与实例状态监控（DB-004）
type TuningService struct {
	db *gorm.DB
}

func NewTuningService(db *gorm.DB) *TuningService {
	return &TuningService{db: db}
}

// VarEntry 参数/状态条目
type VarEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// DBStatus 实例核心状态指标（DB-004）
type DBStatus struct {
	CurrentQPS       float64       `json:"current_qps"`
	ThreadsConnected int           `json:"threads_connected"` // MySQL
	MaxConnections   int           `json:"max_connections"`   // MySQL
	SlowQueries      uint64        `json:"slow_queries"`      // MySQL 累计
	Uptime           uint64        `json:"uptime"`
	Extra            []VarEntry    `json:"extra"` // 引擎相关的其它关键指标
	PGPart           *PGStatusPart `json:"pg,omitempty"`
}

// PGStatusPart PostgreSQL 专属指标
type PGStatusPart struct {
	NumBackends    int    `json:"num_backends"`
	XactCommitted  uint64 `json:"xact_committed"`
	XactRolledBack uint64 `json:"xact_rolled_back"`
	BlocksRead     uint64 `json:"blocks_read"`
	BlocksHit      uint64 `json:"blocks_hit"`
	CacheHitRatio  string `json:"cache_hit_ratio"`
	Deadlocks      uint64 `json:"deadlocks"`
	TempFiles      uint64 `json:"temp_files"`
}

// varNameRegex 参数名合法性校验（防 SQL 注入：SET GLOBAL 不能用占位符）
var varNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`)

func (s *TuningService) getMySQLConn(instanceID uint) (*sql.DB, *model.DatabaseInstance, error) {
	svc := &mysqlConnService{db: s.db}
	return svc.connectInstance(instanceID, "")
}

// GetVariables 读取实例参数列表（DB-002）
func (s *TuningService) GetVariables(instanceID uint, search string) ([]VarEntry, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}

	switch instance.Type {
	case "mysql":
		return s.mysqlVariables(instanceID, search)
	case "postgres":
		return s.pgSettings(&instance, search)
	default:
		return nil, NewUnsupportedTypeError(instance.Type, "该实例暂不支持参数管理")
	}
}

// UpdateVariable 修改运行时参数（DB-002，MySQL SET GLOBAL + PERSIST / PG ALTER SYSTEM）
func (s *TuningService) UpdateVariable(instanceID uint, name, value string) error {
	if !varNameRegex.MatchString(name) {
		return fmt.Errorf("无效的参数名: %s", name)
	}
	if strings.ContainsAny(value, ";\\`\x00") {
		return fmt.Errorf("参数值包含非法字符")
	}
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return fmt.Errorf("实例不存在: %d", instanceID)
	}

	switch instance.Type {
	case "mysql":
		return s.mysqlSetVariable(instanceID, name, value)
	case "postgres":
		return s.pgSetSetting(&instance, name, value)
	default:
		return NewUnsupportedTypeError(instance.Type, "该实例暂不支持参数修改")
	}
}

// GetStatus 实例状态监控（DB-004）
func (s *TuningService) GetStatus(instanceID uint) (*DBStatus, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}
	switch instance.Type {
	case "mysql":
		return s.mysqlStatus(instanceID)
	case "postgres":
		return s.pgStatus(&instance)
	default:
		return nil, NewUnsupportedTypeError(instance.Type, "该实例暂不支持状态监控")
	}
}

// --- MySQL ---

func (s *TuningService) mysqlVariables(instanceID uint, search string) ([]VarEntry, error) {
	conn, _, err := s.getMySQLConn(instanceID)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := "SHOW GLOBAL VARIABLES"
	var args []interface{}
	if search != "" {
		query += " LIKE ?"
		args = append(args, "%"+search+"%")
	}
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("SHOW VARIABLES: %w", err)
	}
	defer rows.Close()

	entries := make([]VarEntry, 0, 64)
	for rows.Next() {
		var e VarEntry
		if err := rows.Scan(&e.Name, &e.Value); err == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (s *TuningService) mysqlSetVariable(instanceID uint, name, value string) error {
	conn, _, err := s.getMySQLConn(instanceID)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 1. 运行时生效
	if _, err := conn.Exec(fmt.Sprintf("SET GLOBAL %s = ?", name), value); err != nil {
		return fmt.Errorf("SET GLOBAL %s 失败: %w", name, err)
	}
	// 2. MySQL 8: SET PERSIST 持久化（5.7 会报错，忽略）
	if _, perr := conn.Exec(fmt.Sprintf("SET PERSIST %s = ?", name), value); perr != nil {
		fmt.Printf("[database-tuning] SET PERSIST %s failed (可能为 5.7): %v\n", name, perr)
	}
	return nil
}

func (s *TuningService) mysqlStatus(instanceID uint) (*DBStatus, error) {
	conn, _, err := s.getMySQLConn(instanceID)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 两次采样 QPS（间隔 1s）
	sample := func() (map[string]string, error) {
		rows, err := conn.Query("SHOW GLOBAL STATUS")
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		m := map[string]string{}
		for rows.Next() {
			var k, v string
			if err := rows.Scan(&k, &v); err == nil {
				m[k] = v
			}
		}
		return m, nil
	}
	s1, err := sample()
	if err != nil {
		return nil, fmt.Errorf("SHOW STATUS: %w", err)
	}
	time.Sleep(time.Second)
	s2, err := sample()
	if err != nil {
		return nil, err
	}

	toU := func(m map[string]string, k string) uint64 {
		var n uint64
		fmt.Sscanf(m[k], "%d", &n)
		return n
	}
	queries1, queries2 := toU(s1, "Queries"), toU(s2, "Queries")

	status := &DBStatus{
		CurrentQPS:       float64(queries2 - queries1),
		ThreadsConnected: int(toU(s2, "Threads_connected")),
		SlowQueries:      toU(s2, "Slow_queries"),
	}
	status.Uptime = toU(s2, "Uptime")

	// max_connections 来自 VARIABLES
	if vars, err := s.mysqlVariables(instanceID, "max_connections"); err == nil {
		for _, v := range vars {
			if v.Name == "max_connections" {
				var n int
				fmt.Sscanf(v.Value, "%d", &n)
				status.MaxConnections = n
			}
		}
	}

	status.Extra = []VarEntry{
		{Name: "Threads_running", Value: s2["Threads_running"]},
		{Name: "Innodb_buffer_pool_usage", Value: mysqlBufferPoolUsage(s2)},
		{Name: "Key_reads", Value: s2["Key_reads"]},
		{Name: "Aborted_connects", Value: s2["Aborted_connects"]},
		{Name: "Created_tmp_disk_tables", Value: s2["Created_tmp_disk_tables"]},
	}
	return status, nil
}

// mysqlBufferPoolUsage 计算 InnoDB 缓冲池使用率
func mysqlBufferPoolUsage(status map[string]string) string {
	toN := func(k string) uint64 {
		var n uint64
		fmt.Sscanf(status[k], "%d", &n)
		return n
	}
	total := toN("Innodb_buffer_pool_pages_total")
	if total == 0 {
		return "N/A"
	}
	free := toN("Innodb_buffer_pool_pages_free")
	return fmt.Sprintf("%.1f%%", float64(total-free)/float64(total)*100)
}

// --- PostgreSQL ---

func (s *TuningService) pgConnectRoot(instance *model.DatabaseInstance) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	engine := &pgEngine{db: s.db, instance: instance}
	return engine.connect(ctx, "postgres")
}

func (s *TuningService) pgSettings(instance *model.DatabaseInstance, search string) ([]VarEntry, error) {
	conn, err := s.pgConnectRoot(instance)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	query := "SELECT name, setting FROM pg_settings"
	var args []interface{}
	if search != "" {
		query += " WHERE name ILIKE $1"
		args = append(args, "%"+search+"%")
	}
	query += " ORDER BY name"

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("pg_settings: %w", err)
	}
	defer rows.Close()

	entries := make([]VarEntry, 0, 64)
	for rows.Next() {
		var e VarEntry
		if err := rows.Scan(&e.Name, &e.Value); err == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (s *TuningService) pgSetSetting(instance *model.DatabaseInstance, name, value string) error {
	conn, err := s.pgConnectRoot(instance)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	// ALTER SYSTEM 持久化 + pg_reload_conf；参数值经校验后拼接（ALTER SYSTEM 不支持参数化）
	if _, err := conn.Exec(context.Background(),
		fmt.Sprintf("ALTER SYSTEM SET %s = '%s'", quoteIdentifier(name), strings.ReplaceAll(value, "'", "''"))); err != nil {
		return fmt.Errorf("ALTER SYSTEM SET 失败: %w", err)
	}
	if _, err := conn.Exec(context.Background(), "SELECT pg_reload_conf()"); err != nil {
		return fmt.Errorf("pg_reload_conf 失败: %w", err)
	}
	return nil
}

func (s *TuningService) pgStatus(instance *model.DatabaseInstance) (*DBStatus, error) {
	conn, err := s.pgConnectRoot(instance)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	status := &DBStatus{PGPart: &PGStatusPart{}}
	row := conn.QueryRow(context.Background(), `
		SELECT numbackends, xact_commit, xact_rollback, blks_read, blks_hit, deadlocks, temp_files
		FROM pg_stat_database WHERE datname = current_database()`)
	var blksRead, blksHit uint64
	if err := row.Scan(&status.PGPart.NumBackends, &status.PGPart.XactCommitted, &status.PGPart.XactRolledBack,
		&blksRead, &blksHit, &status.PGPart.Deadlocks, &status.PGPart.TempFiles); err != nil {
		return nil, fmt.Errorf("pg_stat_database: %w", err)
	}
	status.PGPart.BlocksRead = blksRead
	status.PGPart.BlocksHit = blksHit
	if blksRead+blksHit > 0 {
		status.PGPart.CacheHitRatio = fmt.Sprintf("%.1f%%", float64(blksHit)/float64(blksHit+blksRead)*100)
	} else {
		status.PGPart.CacheHitRatio = "N/A"
	}

	var uptime float64
	if err := conn.QueryRow(context.Background(), "SELECT EXTRACT(EPOCH FROM now() - pg_postmaster_start_time())").Scan(&uptime); err == nil {
		status.Uptime = uint64(uptime)
	}

	// 两次采样 QPS
	count1 := pgQueryCount(conn, "SELECT sum(xact_commit+xact_rollback) FROM pg_stat_database")
	time.Sleep(time.Second)
	count2 := pgQueryCount(conn, "SELECT sum(xact_commit+xact_rollback) FROM pg_stat_database")
	status.CurrentQPS = float64(count2 - count1)

	return status, nil
}

func pgQueryCount(conn *pgx.Conn, query string) uint64 {
	var n uint64
	_ = conn.QueryRow(context.Background(), query).Scan(&n)
	return n
}
